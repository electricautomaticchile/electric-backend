package services

import (
	"context"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/export"
	"fmt"
)

type ExportService struct {
	clienteRepo     ports.PortCliente
	dispositivoRepo ports.PortDispositivo
	alertaRepo      ports.PortAlerta
	boletaRepo      ports.PortBoleta
	excelService    *export.ExcelService
	pdfService      *export.PDFService
}

func NewExportService(
	clienteRepo ports.PortCliente,
	dispositivoRepo ports.PortDispositivo,
	alertaRepo ports.PortAlerta,
	boletaRepo ports.PortBoleta,
) *ExportService {
	return &ExportService{
		clienteRepo:     clienteRepo,
		dispositivoRepo: dispositivoRepo,
		alertaRepo:      alertaRepo,
		boletaRepo:      boletaRepo,
		excelService:    export.NewExcelService(),
		pdfService:      export.NewPDFService(),
	}
}

func (s *ExportService) ExportarClientesExcel(ctx context.Context, empresaID string) ([]byte, error) {
	clientes, err := s.clienteRepo.FindAll(ctx, empresaID)
	if err != nil {
		return nil, err
	}

	data := make([]map[string]interface{}, len(clientes))
	for i, cliente := range clientes {
		data[i] = map[string]interface{}{
			"nombre":         cliente.Nombre,
			"correo":         cliente.Correo,
			"telefono":       cliente.Telefono,
			"direccion":      cliente.Direccion,
			"ciudad":         cliente.Ciudad,
			"tipoCliente":    cliente.TipoCliente,
			"activo":         cliente.Activo,
			"fechaRegistro":  cliente.FechaRegistro.Format("02/01/2006"),
		}
	}

	return s.excelService.ExportarClientes(data)
}

func (s *ExportService) ExportarClientesPDF(ctx context.Context, empresaID string) ([]byte, error) {
	clientes, err := s.clienteRepo.FindAll(ctx, empresaID)
	if err != nil {
		return nil, err
	}

	data := make([]map[string]interface{}, len(clientes))
	for i, cliente := range clientes {
		data[i] = map[string]interface{}{
			"nombre":      cliente.Nombre,
			"correo":      cliente.Correo,
			"telefono":    cliente.Telefono,
			"tipoCliente": cliente.TipoCliente,
			"activo":      cliente.Activo,
		}
	}

	return s.pdfService.ExportarClientes(data)
}

func (s *ExportService) ExportarDispositivosExcel(ctx context.Context, empresaID string) ([]byte, error) {
	dispositivos, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return nil, err
	}

	data := make([]map[string]interface{}, len(dispositivos))
	for i, disp := range dispositivos {
		clienteNombre := "Sin asignar"
		if !disp.ClienteID.IsZero() {
			cliente, _ := s.clienteRepo.FindByID(ctx, disp.ClienteID.Hex())
			if cliente != nil {
				clienteNombre = cliente.Nombre
			}
		}

		consumoActual := 0.0
		if disp.UltimaLectura != nil {
			consumoActual = disp.UltimaLectura.Energy
		}

		data[i] = map[string]interface{}{
			"idDispositivo":    disp.NumeroDispositivo,
			"tipo":             disp.Tipo,
			"modelo":           disp.Nombre,
			"cliente":          clienteNombre,
			"estado":           disp.Estado,
			"consumoActual":    fmt.Sprintf("%.2f", consumoActual),
			"fechaInstalacion": disp.FechaCreacion.Format("02/01/2006"),
		}
	}

	return s.excelService.ExportarDispositivos(data)
}

func (s *ExportService) ExportarDispositivosPDF(ctx context.Context, empresaID string) ([]byte, error) {
	dispositivos, err := s.dispositivoRepo.FindAll(ctx, empresaID)
	if err != nil {
		return nil, err
	}

	data := make([]map[string]interface{}, len(dispositivos))
	for i, disp := range dispositivos {
		clienteNombre := "Sin asignar"
		if !disp.ClienteID.IsZero() {
			cliente, _ := s.clienteRepo.FindByID(ctx, disp.ClienteID.Hex())
			if cliente != nil {
				clienteNombre = cliente.Nombre
			}
		}

		consumoActual := 0.0
		if disp.UltimaLectura != nil {
			consumoActual = disp.UltimaLectura.Energy
		}

		data[i] = map[string]interface{}{
			"idDispositivo":    disp.NumeroDispositivo,
			"tipo":             disp.Tipo,
			"modelo":           disp.Nombre,
			"cliente":          clienteNombre,
			"estado":           disp.Estado,
			"consumoActual":    fmt.Sprintf("%.2f", consumoActual),
			"fechaInstalacion": disp.FechaCreacion.Format("02/01/2006"),
		}
	}

	return s.pdfService.ExportarDispositivos(data)
}

func (s *ExportService) ExportarAlertasExcel(ctx context.Context, empresaID string) ([]byte, error) {
	alertas, err := s.alertaRepo.FindByEmpresa(ctx, empresaID)
	if err != nil {
		return nil, err
	}

	data := make([]map[string]interface{}, len(alertas))
	for i, alerta := range alertas {
		dispositivoNombre := alerta.Dispositivo
		if dispositivoNombre == "" {
			dispositivoNombre = "N/A"
		}

		fechaResolucion := ""
		if alerta.FechaResolucion != nil {
			fechaResolucion = alerta.FechaResolucion.Format("02/01/2006 15:04")
		}

		data[i] = map[string]interface{}{
			"tipo":             alerta.Tipo,
			"mensaje":          alerta.Mensaje,
			"dispositivo":      dispositivoNombre,
			"resuelta":         alerta.Resuelta,
			"fechaCreacion":    alerta.FechaCreacion.Format("02/01/2006 15:04"),
			"fechaResolucion":  fechaResolucion,
		}
	}

	return s.excelService.ExportarAlertas(data)
}

func (s *ExportService) ExportarBoletasExcel(ctx context.Context, empresaID string) ([]byte, error) {
	clientes, err := s.clienteRepo.FindAll(ctx, empresaID)
	if err != nil {
		return nil, err
	}

	var todasBoletas []map[string]interface{}
	
	for _, cliente := range clientes {
		boletas, err := s.boletaRepo.FindByCliente(ctx, cliente.ID)
		if err != nil {
			continue
		}

		for _, boleta := range boletas {
			fechaPago := ""
			if boleta.FechaPago != nil {
				fechaPago = boleta.FechaPago.Format("02/01/2006")
			}

			todasBoletas = append(todasBoletas, map[string]interface{}{
				"numero":           boleta.ID.Hex(),
				"cliente":          cliente.Nombre,
				"periodo":          boleta.Periodo,
				"consumo":          "N/A",
				"monto":            fmt.Sprintf("%.2f", boleta.Monto),
				"estado":           boleta.Estado,
				"fechaEmision":     boleta.FechaCreacion.Format("02/01/2006"),
				"fechaVencimiento": fechaPago,
			})
		}
	}

	return s.excelService.ExportarBoletas(todasBoletas)
}

func (s *ExportService) ExportarBoletaPDF(ctx context.Context, boletaID string) ([]byte, error) {
	boleta, err := s.boletaRepo.FindByID(ctx, boletaID)
	if err != nil {
		return nil, err
	}

	cliente, err := s.clienteRepo.FindByID(ctx, boleta.ClienteID.Hex())
	if err != nil {
		return nil, err
	}

	fechaPago := ""
	if boleta.FechaPago != nil {
		fechaPago = boleta.FechaPago.Format("02/01/2006")
	}

	data := map[string]interface{}{
		"numero":           boleta.ID.Hex(),
		"cliente":          cliente.Nombre,
		"direccion":        cliente.Direccion,
		"periodo":          boleta.Periodo,
		"consumo":          0.0,
		"tarifa":           0.0,
		"monto":            boleta.Monto,
		"fechaEmision":     boleta.FechaCreacion.Format("02/01/2006"),
		"fechaVencimiento": fechaPago,
	}

	return s.pdfService.ExportarBoleta(data)
}
