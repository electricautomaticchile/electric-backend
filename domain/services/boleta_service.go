package services

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/email"
	"electric-backend/infrastructure/entities"
	"electric-backend/infrastructure/sms"
	"electric-backend/types"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BoletaService struct {
	boletaRepo       ports.PortBoleta
	clienteRepo      ports.PortCliente
	dispositivoRepo  ports.PortDispositivo
	notificacionRepo ports.PortNotificacion
	tarifaRepo       ports.PortTarifa
	emailService     email.EmailService
	smsService       sms.SMSService
	wsNotifier       *WebSocketNotifierService
	// Callback para ejecutar corte real en el dispositivo
	OnCortarServicio func(clienteID string)
	// Callback para restablecer servicio cuando el cliente paga
	OnRestablecerServicio func(clienteID string)
	// Callback para marcar corte pendiente (persistente ante reinicio)
	onMarcarCortePendiente func(clienteID string)
}

func NewBoletaService(
	boletaRepo ports.PortBoleta,
	clienteRepo ports.PortCliente,
	emailService email.EmailService,
) *BoletaService {
	return &BoletaService{
		boletaRepo:   boletaRepo,
		clienteRepo:  clienteRepo,
		emailService: emailService,
	}
}

// SetDependencies configura dependencias opcionales post-construcción
func (s *BoletaService) SetDependencies(
	dispositivoRepo ports.PortDispositivo,
	notificacionRepo ports.PortNotificacion,
	tarifaRepo ports.PortTarifa,
	smsService sms.SMSService,
	wsNotifier *WebSocketNotifierService,
) {
	s.dispositivoRepo = dispositivoRepo
	s.notificacionRepo = notificacionRepo
	s.tarifaRepo = tarifaRepo
	s.smsService = smsService
	s.wsNotifier = wsNotifier
}

// SetServicioElectrico conecta los callbacks de corte/reposición con el ServicioElectricoService
func (s *BoletaService) SetServicioElectrico(svc *ServicioElectricoService) {
	s.OnCortarServicio = svc.CortarServicio
	s.OnRestablecerServicio = svc.RestablecerServicio
	s.onMarcarCortePendiente = svc.MarcarCortePendiente
}

// ─── CRUD ────────────────────────────────────────────────────────────────────

func (s *BoletaService) ObtenerPorCliente(ctx context.Context, clienteID string) ([]*models.BoletaModel, error) {
	boletas, err := s.boletaRepo.FindByCliente(ctx, clienteID)
	if err != nil {
		return []*models.BoletaModel{}, nil
	}

	result := make([]*models.BoletaModel, len(boletas))
	for i, boleta := range boletas {
		result[i] = s.entityToModel(boleta)
	}
	return result, nil
}

func (s *BoletaService) ObtenerPorID(ctx context.Context, id string) (*models.BoletaModel, error) {
	boleta, err := s.boletaRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.entityToModel(boleta), nil
}

func (s *BoletaService) ObtenerEmpresaCliente(ctx context.Context, clienteID string) (string, error) {
	cliente, err := s.clienteRepo.FindByID(ctx, clienteID)
	if err != nil {
		return "", err
	}
	return cliente.EmpresaID, nil
}

func (s *BoletaService) Crear(ctx context.Context, r *recipe.CrearBoletaRecipe) (*models.BoletaModel, error) {
	clienteID, err := primitive.ObjectIDFromHex(r.ClienteID)
	if err != nil {
		return nil, types.ThrowRecipe("ClienteID inválido", "clienteId")
	}

	cliente, err := s.clienteRepo.FindByID(ctx, r.ClienteID)
	if err != nil {
		return nil, err
	}
	if cliente.EmpresaID == "" {
		return nil, types.ThrowPower("El cliente no tiene empresa asociada")
	}
	empresaID, err := primitive.ObjectIDFromHex(cliente.EmpresaID)
	if err != nil {
		return nil, types.ThrowRecipe("EmpresaID inválido", "empresaId")
	}

	// Calcular fecha de vencimiento: último día del mes siguiente
	ahora := time.Now()
	mesSiguiente := ahora.AddDate(0, 2, 0) // 2 meses adelante
	ultimoDiaMesSiguiente := time.Date(mesSiguiente.Year(), mesSiguiente.Month(), 0, 23, 59, 59, 0, ahora.Location())

	entity := &entities.BoletaEntity{
		ClienteID:        clienteID,
		EmpresaID:        empresaID,
		Monto:            r.Monto,
		Periodo:          r.Periodo,
		Mes:              int(ahora.Month()),
		Anio:             ahora.Year(),
		FechaVencimiento: &ultimoDiaMesSiguiente,
	}

	if err := s.boletaRepo.Create(ctx, entity); err != nil {
		return nil, err
	}

	go s.enviarNotificacionBoleta(ctx, r.ClienteID, entity)

	return s.entityToModel(entity), nil
}

// ─── RESUMEN DE DEUDA ────────────────────────────────────────────────────────

func (s *BoletaService) ObtenerResumenDeuda(ctx context.Context, clienteID string) (*models.DeudaResumenModel, error) {
	pendientes, err := s.boletaRepo.FindPendientesByCliente(ctx, clienteID)
	if err != nil {
		return nil, err
	}

	vencidas, err := s.boletaRepo.FindVencidasByCliente(ctx, clienteID)
	if err != nil {
		return nil, err
	}

	var montoTotal, montoVencido float64
	var proximoVencimiento *time.Time

	for _, b := range pendientes {
		montoTotal += b.Monto
		if b.FechaVencimiento != nil {
			if proximoVencimiento == nil || b.FechaVencimiento.Before(*proximoVencimiento) {
				fv := *b.FechaVencimiento
				proximoVencimiento = &fv
			}
		}
	}

	for _, b := range vencidas {
		montoVencido += b.Monto
	}

	// Determinar nivel de alerta
	nivelAlerta := "normal"
	numVencidas := len(vencidas)
	switch {
	case numVencidas >= 3:
		nivelAlerta = "corte"
	case numVencidas == 2:
		nivelAlerta = "critico"
	case numVencidas == 1:
		nivelAlerta = "advertencia"
	}

	return &models.DeudaResumenModel{
		BoletasPendientes:  len(pendientes),
		BoletasVencidas:    numVencidas,
		MontoTotal:         montoTotal,
		MontoVencido:       montoVencido,
		ProximoVencimiento: proximoVencimiento,
		NivelAlerta:        nivelAlerta,
	}, nil
}

// ─── CONFIRMAR PAGO ──────────────────────────────────────────────────────────

func (s *BoletaService) ConfirmarPago(ctx context.Context, boletaID string) (*models.ConfirmarPagoResponse, error) {
	boleta, err := s.boletaRepo.FindByID(ctx, boletaID)
	if err != nil {
		return nil, err
	}

	if boleta.Estado == "pagado" {
		return nil, fmt.Errorf("esta boleta ya fue pagada")
	}

	// Marcar como pagada
	ahora := time.Now()
	boleta.Estado = "pagado"
	boleta.FechaPago = &ahora

	if err := s.boletaRepo.Update(ctx, boletaID, boleta); err != nil {
		return nil, err
	}

	clienteID := boleta.ClienteID.Hex()

	// Verificar si se debe reponer el servicio (< 2 boletas vencidas restantes)
	servicioRepuesto := false
	vencidas, _ := s.boletaRepo.FindVencidasByCliente(ctx, clienteID)
	if len(vencidas) < 2 && s.OnRestablecerServicio != nil {
		s.OnRestablecerServicio(clienteID)
		servicioRepuesto = true
	}

	// Notificar pago
	go s.enviarNotificacionPago(ctx, clienteID, boleta, servicioRepuesto)

	mensaje := "Pago confirmado correctamente."
	if servicioRepuesto {
		mensaje = "Pago confirmado. Suministro repuesto automáticamente."
	}

	return &models.ConfirmarPagoResponse{
		BoletaID:         boletaID,
		Estado:           "pagado",
		FechaPago:        ahora,
		ServicioRepuesto: servicioRepuesto,
		Mensaje:          mensaje,
	}, nil
}

// ─── SCHEDULER: VERIFICAR VENCIMIENTOS (diario) ─────────────────────────────

func (s *BoletaService) VerificarVencimientos(ctx context.Context) error {
	// PASO 1: Marcar boletas por vencer (3 días antes)
	porVencer, err := s.boletaRepo.FindPorVencer(ctx, 3)
	if err != nil {
		return fmt.Errorf("error buscando boletas por vencer: %w", err)
	}

	for _, boleta := range porVencer {
		if err := s.boletaRepo.UpdateEstado(ctx, boleta.ID.Hex(), "por_vencer"); err != nil {
			log.Printf("Error actualizando boleta %s a por_vencer: %v", boleta.ID.Hex(), err)
			continue
		}

		if !boleta.NotificacionesEnviadas.PorVencer3Dias {
			go s.enviarAvisoPorVencer(ctx, boleta)
			s.boletaRepo.UpdateNotificacionEnviada(ctx, boleta.ID.Hex(), "porVencer3dias")
		}
	}

	// PASO 2: Marcar boletas vencidas
	vencidas, err := s.boletaRepo.FindVencidas(ctx)
	if err != nil {
		return fmt.Errorf("error buscando boletas vencidas: %w", err)
	}

	for _, boleta := range vencidas {
		if err := s.boletaRepo.UpdateEstado(ctx, boleta.ID.Hex(), "vencido"); err != nil {
			log.Printf("Error actualizando boleta %s a vencido: %v", boleta.ID.Hex(), err)
			continue
		}

		if !boleta.NotificacionesEnviadas.Vencida {
			go s.enviarAvisoVencida(ctx, boleta)
			s.boletaRepo.UpdateNotificacionEnviada(ctx, boleta.ID.Hex(), "vencida")
		}
	}

	// PASO 3: Escalada por cliente — verificar clientes con múltiples boletas vencidas
	if err := s.ejecutarEscaladaCortes(ctx); err != nil {
		log.Printf("Error en escalada de cortes: %v", err)
	}

	return nil
}

func (s *BoletaService) ejecutarEscaladaCortes(ctx context.Context) error {
	// Query optimizado: solo trae IDs de clientes con 2+ boletas vencidas
	clienteIDs, err := s.boletaRepo.FindClienteIDsConBoletasVencidas(ctx)
	if err != nil {
		return err
	}

	if len(clienteIDs) == 0 {
		return nil
	}

	log.Printf("Escalada cortes: %d cliente(s) con boletas vencidas", len(clienteIDs))

	for _, clienteID := range clienteIDs {
		cliente, err := s.clienteRepo.FindByID(ctx, clienteID)
		if err != nil || cliente == nil {
			continue
		}

		vencidas, err := s.boletaRepo.FindVencidasByCliente(ctx, clienteID)
		if err != nil || len(vencidas) == 0 {
			continue
		}

		numVencidas := len(vencidas)
		log.Printf("Cliente %s (%s): %d boletas vencidas", cliente.Nombre, clienteID, numVencidas)

		switch {
		case numVencidas == 2:
			// Advertencia formal
			for _, b := range vencidas {
				if !b.NotificacionesEnviadas.Advertencia2Boletas {
					go s.enviarAdvertenciaFormal(ctx, cliente, vencidas)
					s.boletaRepo.UpdateNotificacionEnviada(ctx, b.ID.Hex(), "advertencia2boletas")
				}
			}

		case numVencidas >= 3:
			boletaMasReciente := vencidas[len(vencidas)-1]

			if !boletaMasReciente.NotificacionesEnviadas.CorteEjecutado {
				// Marcar corte pendiente en MongoDB (persistente ante reinicio)
				if s.onMarcarCortePendiente != nil {
					s.onMarcarCortePendiente(cliente.ID)
				}
				// Aviso 30 seg → espera → corte (todo en una goroutine)
				go s.ejecutarCorteConAviso(ctx, cliente, vencidas)
				// Marcar todas las notificaciones como enviadas para no repetir
				s.boletaRepo.UpdateNotificacionEnviada(ctx, boletaMasReciente.ID.Hex(), "aviso48h")
				s.boletaRepo.UpdateNotificacionEnviada(ctx, boletaMasReciente.ID.Hex(), "aviso24h")
				s.boletaRepo.UpdateNotificacionEnviada(ctx, boletaMasReciente.ID.Hex(), "corteEjecutado")
			}
		}
	}

	return nil
}

// VerificarEscaladaCortes — solo ejecuta la escalada de cortes.
// Se llama cada 5 min para detectar nuevas situaciones sin reiniciar el servidor.
func (s *BoletaService) VerificarEscaladaCortes(ctx context.Context) error {
	return s.ejecutarEscaladaCortes(ctx)
}

// ─── GENERACIÓN AUTOMÁTICA DE BOLETAS (fin de mes) ──────────────────────────

func (s *BoletaService) GenerarBoletasMensuales(ctx context.Context) error {
	clientes, err := s.clienteRepo.FindAll(ctx, "")
	if err != nil {
		return fmt.Errorf("error obteniendo clientes: %w", err)
	}

	ahora := time.Now()
	mesActual := int(ahora.Month())
	anioActual := ahora.Year()

	// Fecha de vencimiento: último día del mes siguiente
	mesSiguiente := ahora.AddDate(0, 2, 0)
	ultimoDiaMesSiguiente := time.Date(mesSiguiente.Year(), mesSiguiente.Month(), 0, 23, 59, 59, 0, ahora.Location())

	meses := []string{"", "Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
		"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"}
	periodo := fmt.Sprintf("%s %d", meses[mesActual], anioActual)

	for _, cliente := range clientes {
		if !cliente.Activo {
			continue
		}

		// Verificar que no exista boleta para este periodo
		boletasCliente, _ := s.boletaRepo.FindByCliente(ctx, cliente.ID)
		yaExiste := false
		for _, b := range boletasCliente {
			if b.Mes == mesActual && b.Anio == anioActual {
				yaExiste = true
				break
			}
		}
		if yaExiste {
			continue
		}

		// Obtener consumo del mes desde dispositivos
		var consumoKwh float64
		var costoTotal float64

		if s.dispositivoRepo != nil {
			dispositivos, err := s.dispositivoRepo.FindByCliente(ctx, cliente.ID)
			if err == nil {
				for _, d := range dispositivos {
					if d.UltimaLectura != nil {
						consumoKwh += d.UltimaLectura.ConsumoKWh
						costoTotal += d.UltimaLectura.CostoEstimado
					}
				}
			}
		}

		if consumoKwh == 0 {
			continue // Sin consumo, no generar boleta
		}

		clienteOID, _ := primitive.ObjectIDFromHex(cliente.ID)

		entity := &entities.BoletaEntity{
			ClienteID:        clienteOID,
			Monto:            costoTotal,
			MontoTotal:       costoTotal,
			Periodo:          periodo,
			Mes:              mesActual,
			Anio:             anioActual,
			ConsumoKwh:       consumoKwh,
			FechaVencimiento: &ultimoDiaMesSiguiente,
		}

		if err := s.boletaRepo.Create(ctx, entity); err != nil {
			log.Printf("Error creando boleta para cliente %s: %v", cliente.ID, err)
			continue
		}

		go s.enviarNotificacionBoleta(ctx, cliente.ID, entity)
		log.Printf("Boleta generada: cliente=%s periodo=%s monto=%.0f kwh=%.1f", cliente.ID, periodo, costoTotal, consumoKwh)
	}

	return nil
}

// ─── NOTIFICACIONES ─────────────────────────────────────────────────────────

func (s *BoletaService) enviarNotificacionBoleta(ctx context.Context, clienteID string, boleta *entities.BoletaEntity) {
	cliente, err := s.clienteRepo.FindByID(ctx, clienteID)
	if err != nil {
		log.Printf("Error obteniendo cliente para email de boleta: %v", err)
		return
	}

	monto := formatearCLP(boleta.Monto)
	fechaVenc := "Sin fecha"
	if boleta.FechaVencimiento != nil {
		fechaVenc = boleta.FechaVencimiento.Format("02/01/2006")
	}

	// Email
	if cliente.Correo != "" {
		numeroBoleta := boleta.ID.Hex()[:8]
		if err := s.emailService.EnviarNotificacionBoleta(
			cliente.Correo, cliente.Nombre, numeroBoleta, monto, fechaVenc,
		); err != nil {
			log.Printf("Error enviando email de boleta: %v", err)
		}
	}

	// SMS
	if s.smsService != nil && cliente.NotificacionesSMS && cliente.Telefono != "" {
		msg := fmt.Sprintf(
			"Hola %s, tu boleta de %s por $%s fue generada. Vence el %s.\nPaga en: electricautomaticchile.com/cliente\n- ElectricAutomaticChile",
			cliente.Nombre, boleta.Periodo, monto, fechaVenc,
		)
		if err := s.smsService.EnviarSMS(cliente.Telefono, msg); err != nil {
			log.Printf("Error enviando SMS de boleta: %v", err)
		}
	}

	// Notificación en app
	s.crearNotificacionApp(ctx, clienteID, "Nueva boleta generada",
		fmt.Sprintf("Tu boleta de %s por $%s fue generada. Vence el %s.", boleta.Periodo, monto, fechaVenc),
		"facturacion", "info")
}

func (s *BoletaService) enviarNotificacionPago(ctx context.Context, clienteID string, boleta *entities.BoletaEntity, servicioRepuesto bool) {
	cliente, err := s.clienteRepo.FindByID(ctx, clienteID)
	if err != nil {
		return
	}

	monto := formatearCLP(boleta.Monto)

	// SMS
	if s.smsService != nil && cliente.Telefono != "" {
		msg := fmt.Sprintf("✅ Pago confirmado de $%s (%s).", monto, boleta.Periodo)
		if servicioRepuesto {
			msg += " Tu suministro electrico fue repuesto automaticamente."
		}
		msg += "\n- ElectricAutomaticChile"
		s.smsService.EnviarSMS(cliente.Telefono, msg)
	}

	// Email
	if s.emailService != nil && cliente.Correo != "" {
		s.emailService.EnviarPagoConfirmado(cliente.Correo, cliente.Nombre, boleta.Periodo, monto, servicioRepuesto)
	}

	// Notificación en app
	titulo := "Pago confirmado"
	mensaje := fmt.Sprintf("Tu pago de $%s (%s) fue recibido correctamente.", monto, boleta.Periodo)
	if servicioRepuesto {
		mensaje += " Tu suministro eléctrico fue repuesto automáticamente."
	}
	s.crearNotificacionApp(ctx, clienteID, titulo, mensaje, "facturacion", "info")
}

func (s *BoletaService) enviarAvisoPorVencer(ctx context.Context, boleta *entities.BoletaEntity) {
	clienteID := boleta.ClienteID.Hex()
	cliente, err := s.clienteRepo.FindByID(ctx, clienteID)
	if err != nil {
		return
	}

	monto := formatearCLP(boleta.Monto)
	fechaVenc := ""
	if boleta.FechaVencimiento != nil {
		fechaVenc = boleta.FechaVencimiento.Format("02/01/2006")
	}

	// SMS
	if s.smsService != nil && cliente.NotificacionesSMS && cliente.Telefono != "" {
		msg := fmt.Sprintf(
			"⚠️ Hola %s, tu boleta de %s por $%s vence el %s. Paga a tiempo para evitar el corte del servicio.\nelectricautomaticchile.com/cliente\n- ElectricAutomaticChile",
			cliente.Nombre, boleta.Periodo, monto, fechaVenc,
		)
		s.smsService.EnviarSMS(cliente.Telefono, msg)
	}

	// Email
	if s.emailService != nil && cliente.Correo != "" {
		s.emailService.EnviarBoletaVenciendo(cliente.Correo, cliente.Nombre, boleta.Periodo, monto, fechaVenc)
	}

	s.crearNotificacionApp(ctx, clienteID, "Boleta por vencer",
		fmt.Sprintf("Tu boleta de %s por $%s vence el %s. Paga a tiempo.", boleta.Periodo, monto, fechaVenc),
		"facturacion", "warning")
}

func (s *BoletaService) enviarAvisoVencida(ctx context.Context, boleta *entities.BoletaEntity) {
	clienteID := boleta.ClienteID.Hex()
	cliente, err := s.clienteRepo.FindByID(ctx, clienteID)
	if err != nil {
		return
	}

	monto := formatearCLP(boleta.Monto)

	// SMS
	if s.smsService != nil && cliente.NotificacionesSMS && cliente.Telefono != "" {
		msg := fmt.Sprintf(
			"🔴 Hola %s, tu boleta de %s por $%s ha vencido. Paga lo antes posible para evitar la suspension del servicio.\nelectricautomaticchile.com/cliente\n- ElectricAutomaticChile",
			cliente.Nombre, boleta.Periodo, monto,
		)
		s.smsService.EnviarSMS(cliente.Telefono, msg)
	}

	// Email
	if s.emailService != nil && cliente.Correo != "" {
		s.emailService.EnviarBoletaVencida(cliente.Correo, cliente.Nombre, boleta.Periodo, monto)
	}

	s.crearNotificacionApp(ctx, clienteID, "Boleta vencida",
		fmt.Sprintf("Tu boleta de %s por $%s ha vencido. Paga para evitar el corte.", boleta.Periodo, monto),
		"alerta", "warning")
}

func (s *BoletaService) enviarAdvertenciaFormal(ctx context.Context, cliente *models.ClienteModel, vencidas []*entities.BoletaEntity) {
	var montoTotal float64
	for _, b := range vencidas {
		montoTotal += b.Monto
	}
	monto := formatearCLP(montoTotal)

	// SMS
	if s.smsService != nil && cliente.NotificacionesSMS && cliente.Telefono != "" {
		msg := fmt.Sprintf(
			"⚠️ ADVERTENCIA: Hola %s, tienes %d boletas vencidas por $%s en total.\nAl tercer impago se suspendera tu suministro electrico.\nPaga en: electricautomaticchile.com/cliente\n- ElectricAutomaticChile",
			cliente.Nombre, len(vencidas), monto,
		)
		s.smsService.EnviarSMS(cliente.Telefono, msg)
	}

	// Email
	if s.emailService != nil && cliente.Correo != "" {
		s.emailService.EnviarAdvertenciaCorte(cliente.Correo, cliente.Nombre, len(vencidas), monto)
	}

	s.crearNotificacionApp(ctx, cliente.ID, "⚠️ 2 boletas vencidas",
		fmt.Sprintf("Tienes %d boletas vencidas por $%s. Al tercer impago se suspenderá tu suministro.", len(vencidas), monto),
		"alerta", "critica")
}

func (s *BoletaService) enviarAviso48h(ctx context.Context, cliente *models.ClienteModel, vencidas []*entities.BoletaEntity) {
	var montoTotal float64
	for _, b := range vencidas {
		montoTotal += b.Monto
	}
	monto := formatearCLP(montoTotal)

	// SMS
	if s.smsService != nil && cliente.Telefono != "" {
		msg := fmt.Sprintf(
			"⚠️ AVISO 48 HORAS: Hola %s, en 48 horas se generara tu 3ra boleta impaga.\nSi no pagas antes, tu suministro electrico sera CORTADO automaticamente.\nDeuda actual: $%s\nPaga ahora: electricautomaticchile.com/cliente\n- ElectricAutomaticChile",
			cliente.Nombre, monto,
		)
		s.smsService.EnviarSMS(cliente.Telefono, msg)
	}

	// Email
	if s.emailService != nil && cliente.Correo != "" {
		s.emailService.EnviarAvisoCorte(cliente.Correo, cliente.Nombre,
			"Aviso: en 48 horas se cortará tu suministro eléctrico",
			fmt.Sprintf("En 48 horas se generará tu 3ra boleta impaga y tu suministro eléctrico será CORTADO automáticamente si no pagas antes. Deuda actual: $%s.", monto),
			len(vencidas), monto)
	}

	s.crearNotificacionApp(ctx, cliente.ID, "⚠️ Aviso: corte en 48 horas",
		fmt.Sprintf("En 48 horas se generará tu 3ra boleta y se cortará el suministro si no pagas. Deuda: $%s.", monto),
		"alerta", "critica")
}

func (s *BoletaService) enviarAviso24h(ctx context.Context, cliente *models.ClienteModel, vencidas []*entities.BoletaEntity) {
	var montoTotal float64
	for _, b := range vencidas {
		montoTotal += b.Monto
	}
	monto := formatearCLP(montoTotal)

	// SMS
	if s.smsService != nil && cliente.Telefono != "" {
		msg := fmt.Sprintf(
			"🔴 ULTIMO AVISO 24 HORAS: Hola %s, en 24 horas se generara tu 3ra boleta impaga y tu suministro electrico sera CORTADO.\nPaga $%s AHORA para evitarlo.\nelectricautomaticchile.com/cliente\n- ElectricAutomaticChile",
			cliente.Nombre, monto,
		)
		s.smsService.EnviarSMS(cliente.Telefono, msg)
	}

	// Email
	if s.emailService != nil && cliente.Correo != "" {
		s.emailService.EnviarAvisoCorte(cliente.Correo, cliente.Nombre,
			"Último aviso: en 24 horas se cortará tu suministro eléctrico",
			fmt.Sprintf("En 24 horas se generará tu 3ra boleta impaga y tu suministro eléctrico será CORTADO. Paga $%s AHORA para evitarlo.", monto),
			len(vencidas), monto)
	}

	s.crearNotificacionApp(ctx, cliente.ID, "🔴 Último aviso: corte en 24 horas",
		fmt.Sprintf("En 24 horas se generará tu 3ra boleta y se cortará el suministro. Paga $%s ahora.", monto),
		"alerta", "error")
}

func (s *BoletaService) enviarAviso30Seg(ctx context.Context, cliente *models.ClienteModel, vencidas []*entities.BoletaEntity) {
	var montoTotal float64
	for _, b := range vencidas {
		montoTotal += b.Monto
	}
	monto := formatearCLP(montoTotal)

	// SMS
	if s.smsService != nil && cliente.Telefono != "" {
		msg := fmt.Sprintf(
			"🔴 CORTE INMINENTE: Hola %s, tu suministro electrico se cortara en 30 SEGUNDOS por 3 boletas impagas ($%s).\nSi ya pagaste, ignora este mensaje.\n- ElectricAutomaticChile",
			cliente.Nombre, monto,
		)
		s.smsService.EnviarSMS(cliente.Telefono, msg)
	}

	// Email
	if s.emailService != nil && cliente.Correo != "" {
		s.emailService.EnviarAvisoCorte(cliente.Correo, cliente.Nombre,
			"CORTE INMINENTE: Tu suministro se cortará en 30 segundos",
			fmt.Sprintf("Tu suministro eléctrico se cortará en 30 SEGUNDOS por 3 boletas impagas ($%s). Si ya pagaste, ignora este mensaje.", monto),
			len(vencidas), monto)
	}

	s.crearNotificacionApp(ctx, cliente.ID, "🔴 Corte en 30 segundos",
		fmt.Sprintf("Tu suministro se cortará en 30 segundos por 3 boletas impagas ($%s). Paga ahora para evitarlo.", monto),
		"alerta", "error")
}

func (s *BoletaService) ejecutarCorteAutomatico(ctx context.Context, cliente *models.ClienteModel, vencidas []*entities.BoletaEntity) {
	var montoTotal float64
	for _, b := range vencidas {
		montoTotal += b.Monto
	}
	monto := formatearCLP(montoTotal)

	if s.OnCortarServicio != nil {
		s.OnCortarServicio(cliente.ID)
		log.Printf("CORTE AUTOMÁTICO ejecutado: cliente=%s boletas=%d monto=$%s", cliente.ID, len(vencidas), monto)
	} else {
		log.Printf("CORTE AUTOMÁTICO (sin dispositivo): cliente=%s boletas=%d monto=$%s", cliente.ID, len(vencidas), monto)
	}

	// SMS
	if s.smsService != nil && cliente.Telefono != "" {
		msg := fmt.Sprintf(
			"🔴 Tu suministro electrico fue SUSPENDIDO por %d boletas vencidas ($%s).\nPaga tus boletas para restablecer el servicio:\nelectricautomaticchile.com/cliente\n- ElectricAutomaticChile",
			len(vencidas), monto,
		)
		s.smsService.EnviarSMS(cliente.Telefono, msg)
	}

	// Email
	if s.emailService != nil && cliente.Correo != "" {
		s.emailService.EnviarServicioSuspendido(cliente.Correo, cliente.Nombre, len(vencidas), monto)
	}

	s.crearNotificacionApp(ctx, cliente.ID, "🔴 Suministro suspendido",
		fmt.Sprintf("Tu suministro fue suspendido por %d boletas vencidas ($%s). Paga para restablecer.", len(vencidas), monto),
		"alerta", "error")
}

// ejecutarCorteConAviso envía aviso de 30 seg, espera, y luego ejecuta el corte.
// El estado "corte_pendiente" ya fue persistido en MongoDB antes de llamar esta función,
// así que si el servidor se reinicia, EjecutarCortesPendientes lo retomará.
func (s *BoletaService) ejecutarCorteConAviso(ctx context.Context, cliente *models.ClienteModel, vencidas []*entities.BoletaEntity) {
	s.enviarAviso30Seg(ctx, cliente, vencidas)
	time.Sleep(30 * time.Second)
	s.ejecutarCorteAutomatico(ctx, cliente, vencidas)
}

// formatearCLP formatea un monto como pesos chilenos con separador de miles
func formatearCLP(monto float64) string {
	n := int64(monto)
	s := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return result
}

func (s *BoletaService) crearNotificacionApp(ctx context.Context, clienteID, titulo, mensaje, tipo, severidad string) {
	if s.notificacionRepo == nil {
		return
	}

	destinatarioID, err := primitive.ObjectIDFromHex(clienteID)
	if err != nil {
		return
	}

	entity := &entities.NotificacionEntity{
		DestinatarioID: destinatarioID,
		Titulo:         titulo,
		Mensaje:        mensaje,
		Tipo:           tipo,
		Severidad:      severidad,
		Importante:     severidad == "critica" || severidad == "error",
	}

	if err := s.notificacionRepo.Create(ctx, entity); err != nil {
		log.Printf("Error creando notificación app: %v", err)
	}

	if s.wsNotifier != nil {
		s.wsNotifier.NotificarNuevaNotificacion(entity)
	}
}

// ─── ENTITY TO MODEL ────────────────────────────────────────────────────────

func (s *BoletaService) entityToModel(entity *entities.BoletaEntity) *models.BoletaModel {
	model := &models.BoletaModel{
		ID:               entity.ID.Hex(),
		ClienteID:        entity.ClienteID.Hex(),
		Monto:            entity.Monto,
		Periodo:          entity.Periodo,
		Mes:              entity.Mes,
		Anio:             entity.Anio,
		ConsumoKwh:       entity.ConsumoKwh,
		Estado:           entity.Estado,
		FechaCreacion:    entity.FechaCreacion,
		FechaVencimiento: entity.FechaVencimiento,
		FechaPago:        entity.FechaPago,
		MotivoCorte:      entity.MotivoCorte,
	}

	if !entity.EmpresaID.IsZero() {
		model.EmpresaID = entity.EmpresaID.Hex()
	}
	if !entity.DispositivoID.IsZero() {
		model.DispositivoID = entity.DispositivoID.Hex()
	}

	return model
}
