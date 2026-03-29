package email

// EmailService define la interfaz común para proveedores de email.
type EmailService interface {
	EnviarEmail(to []string, subject, htmlBody, textBody string) error
	EnviarNotificacionAlerta(destinatario, nombreCliente, tipoAlerta, mensaje string) error
	EnviarNotificacionTicket(destinatario, nombreUsuario, numeroTicket, asunto, mensaje string) error
	EnviarNotificacionBoleta(destinatario, nombreCliente, numeroBoleta, monto, fechaVencimiento string) error
	EnviarRecuperacionPassword(destinatario, nombreUsuario, token string) error
	EnviarBienvenida(destinatario, nombreUsuario, tipoUsuario string) error
	EnviarCredenciales(destinatario, nombreCliente, numeroCliente, passwordTemporal string) error
}
