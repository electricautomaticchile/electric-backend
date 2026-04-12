package email

// EmailService define la interfaz común para proveedores de email.
type EmailService interface {
	EnviarEmail(to []string, subject, htmlBody, textBody string) error
	// Existentes
	EnviarNotificacionAlerta(destinatario, nombreCliente, tipoAlerta, mensaje string) error
	EnviarNotificacionTicket(destinatario, nombreUsuario, numeroTicket, asunto, mensaje string) error
	EnviarNotificacionBoleta(destinatario, nombreCliente, numeroBoleta, monto, fechaVencimiento string) error
	EnviarRecuperacionPassword(destinatario, nombreUsuario, token string) error
	EnviarBienvenida(destinatario, nombreUsuario, tipoUsuario string) error
	EnviarCredenciales(destinatario, nombreCliente, numeroCliente, passwordTemporal string) error
	// Boletas y suministro
	EnviarBoletaVenciendo(destinatario, nombreCliente, periodo, monto, fechaVencimiento string) error
	EnviarBoletaVencida(destinatario, nombreCliente, periodo, monto string) error
	EnviarAdvertenciaCorte(destinatario, nombreCliente string, numBoletas int, montoTotal string) error
	EnviarAvisoCorte(destinatario, nombreCliente, titulo, mensaje string, numBoletas int, montoTotal string) error
	EnviarServicioSuspendido(destinatario, nombreCliente string, numBoletas int, montoTotal string) error
	EnviarPagoConfirmado(destinatario, nombreCliente, periodo, monto string, servicioRepuesto bool) error
}
