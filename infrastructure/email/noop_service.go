package email

import (
	"fmt"
	"log"
)

type NoopService struct {
	from string
}

func NewNoopService(from string) *NoopService {
	if from == "" {
		from = "noreply@electricautomaticchile.com"
	}
	log.Printf("email: proveedor externo deshabilitado; usando no-op desde %s", from)
	return &NoopService{from: from}
}

func (s *NoopService) EnviarEmail(to []string, subject, htmlBody, textBody string) error {
	log.Printf("email noop: omitido envio a=%v asunto=%q", to, subject)
	return nil
}

func (s *NoopService) EnviarNotificacionAlerta(destinatario, nombreCliente, tipoAlerta, mensaje string) error {
	return s.EnviarEmail(
		[]string{destinatario},
		fmt.Sprintf("Alerta: %s", tipoAlerta),
		mensaje,
		fmt.Sprintf("Hola %s, se genero una alerta: %s", nombreCliente, mensaje),
	)
}

func (s *NoopService) EnviarNotificacionTicket(destinatario, nombreUsuario, numeroTicket, asunto, mensaje string) error {
	return s.EnviarEmail(
		[]string{destinatario},
		fmt.Sprintf("Ticket #%s: %s", numeroTicket, asunto),
		mensaje,
		fmt.Sprintf("Hola %s, actualizacion del ticket #%s: %s", nombreUsuario, numeroTicket, mensaje),
	)
}

func (s *NoopService) EnviarNotificacionBoleta(destinatario, nombreCliente, numeroBoleta, monto, fechaVencimiento string) error {
	return s.EnviarEmail(
		[]string{destinatario},
		fmt.Sprintf("Nueva boleta #%s", numeroBoleta),
		"",
		fmt.Sprintf("Hola %s, boleta #%s por $%s vence el %s.", nombreCliente, numeroBoleta, monto, fechaVencimiento),
	)
}

func (s *NoopService) EnviarRecuperacionPassword(destinatario, nombreUsuario, token string) error {
	return s.EnviarEmail(
		[]string{destinatario},
		"Recuperacion de contrasena",
		"",
		fmt.Sprintf("Hola %s, token de recuperacion: %s", nombreUsuario, token),
	)
}

func (s *NoopService) EnviarCredenciales(destinatario, nombreCliente, numeroCliente, passwordTemporal string) error {
	return s.EnviarEmail(
		[]string{destinatario},
		"Credenciales de acceso",
		"",
		fmt.Sprintf("Hola %s, numero cliente: %s, password temporal: %s", nombreCliente, numeroCliente, passwordTemporal),
	)
}

func (s *NoopService) EnviarBoletaVenciendo(destinatario, nombreCliente, periodo, monto, fechaVencimiento string) error {
	return s.EnviarEmail(
		[]string{destinatario},
		fmt.Sprintf("Boleta de %s vence el %s", periodo, fechaVencimiento),
		"",
		fmt.Sprintf("Hola %s, tu boleta de %s por $%s vence el %s.", nombreCliente, periodo, monto, fechaVencimiento),
	)
}

func (s *NoopService) EnviarBoletaVencida(destinatario, nombreCliente, periodo, monto string) error {
	return s.EnviarEmail(
		[]string{destinatario},
		fmt.Sprintf("Boleta vencida: %s", periodo),
		"",
		fmt.Sprintf("Hola %s, tu boleta de %s por $%s ha vencido.", nombreCliente, periodo, monto),
	)
}

func (s *NoopService) EnviarAdvertenciaCorte(destinatario, nombreCliente string, numBoletas int, montoTotal string) error {
	return s.EnviarEmail(
		[]string{destinatario},
		fmt.Sprintf("Advertencia: %d boletas vencidas", numBoletas),
		"",
		fmt.Sprintf("Hola %s, tienes %d boletas vencidas por $%s.", nombreCliente, numBoletas, montoTotal),
	)
}

func (s *NoopService) EnviarAvisoCorte(destinatario, nombreCliente, titulo, mensaje string, numBoletas int, montoTotal string) error {
	return s.EnviarEmail(
		[]string{destinatario},
		titulo,
		"",
		fmt.Sprintf("Hola %s, %s Deuda total: $%s.", nombreCliente, mensaje, montoTotal),
	)
}

func (s *NoopService) EnviarServicioSuspendido(destinatario, nombreCliente string, numBoletas int, montoTotal string) error {
	return s.EnviarEmail(
		[]string{destinatario},
		"Servicio suspendido",
		"",
		fmt.Sprintf("Hola %s, tu servicio fue suspendido por %d boletas vencidas ($%s).", nombreCliente, numBoletas, montoTotal),
	)
}

func (s *NoopService) EnviarPagoConfirmado(destinatario, nombreCliente, periodo, monto string, servicioRepuesto bool) error {
	mensaje := fmt.Sprintf("Hola %s, tu pago de $%s (%s) fue recibido.", nombreCliente, monto, periodo)
	if servicioRepuesto {
		mensaje += " Tu suministro fue repuesto."
	}
	return s.EnviarEmail([]string{destinatario}, "Pago confirmado", "", mensaje)
}
