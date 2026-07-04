package email

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

// SESService envía emails transaccionales usando AWS SES v2.
//
// Credenciales: usa la cadena estándar de AWS (variables de entorno
// AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY / AWS_REGION, o rol IAM).
// El remitente (from) debe estar verificado en SES.
type SESService struct {
	client *sesv2.Client
	from   string
}

// NewSESService crea el servicio SES. region por defecto us-east-1.
func NewSESService(from, region string) (*SESService, error) {
	if from == "" {
		from = "noreply@electricautomaticchile.com"
	}
	if region == "" {
		region = "us-east-1"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("cargando config AWS: %w", err)
	}

	log.Printf("email: proveedor AWS SES activo (region=%s, from=%s)", region, from)
	return &SESService{client: sesv2.NewFromConfig(cfg), from: from}, nil
}

// EnviarEmail realiza el envío real vía SES.
func (s *SESService) EnviarEmail(to []string, subject, htmlBody, textBody string) error {
	if len(to) == 0 {
		return fmt.Errorf("destinatario vacío")
	}
	if textBody == "" {
		textBody = subject
	}

	body := &types.Body{
		Text: &types.Content{Data: aws.String(textBody), Charset: aws.String("UTF-8")},
	}
	if htmlBody != "" {
		body.Html = &types.Content{Data: aws.String(htmlBody), Charset: aws.String("UTF-8")}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := s.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.from),
		Destination:      &types.Destination{ToAddresses: to},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: aws.String(subject), Charset: aws.String("UTF-8")},
				Body:    body,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("SES SendEmail: %w", err)
	}
	return nil
}

func (s *SESService) EnviarNotificacionAlerta(destinatario, nombreCliente, tipoAlerta, mensaje string) error {
	return s.EnviarEmail([]string{destinatario}, fmt.Sprintf("Alerta: %s", tipoAlerta), mensaje,
		fmt.Sprintf("Hola %s, se generó una alerta: %s", nombreCliente, mensaje))
}

func (s *SESService) EnviarNotificacion(destinatario, nombreCliente, tipo, titulo, mensaje string) error {
	asunto := titulo
	if asunto == "" {
		asunto = "Nueva notificación"
	}
	texto := fmt.Sprintf("Hola %s,\n\n%s\n\n%s\n\n— ElectricAutomaticChile",
		nombreCliente, titulo, mensaje)
	html := renderNotificacionHTML(nombreCliente, tipo, titulo, mensaje)
	return s.EnviarEmail([]string{destinatario}, asunto, html, texto)
}

// renderNotificacionHTML arma un correo HTML sencillo y con la marca para
// cualquier notificación del sistema.
func renderNotificacionHTML(nombreCliente, tipo, titulo, mensaje string) string {
	etiqueta := tipo
	if etiqueta == "" {
		etiqueta = "notificación"
	}
	return fmt.Sprintf(`<!DOCTYPE html><html lang="es"><body style="margin:0;padding:0;background:#f4f4f5;font-family:Arial,Helvetica,sans-serif;">`+
		`<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:#f4f4f5;padding:24px 0;">`+
		`<tr><td align="center">`+
		`<table role="presentation" width="560" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;overflow:hidden;border:1px solid #e5e7eb;">`+
		`<tr><td style="background:#0a0a0a;padding:20px 28px;color:#ffffff;font-size:18px;font-weight:bold;">ElectricAutomaticChile</td></tr>`+
		`<tr><td style="padding:28px;">`+
		`<span style="display:inline-block;background:#eff6ff;color:#1d4ed8;font-size:11px;text-transform:uppercase;letter-spacing:.5px;padding:4px 10px;border-radius:999px;margin-bottom:14px;">%s</span>`+
		`<h1 style="margin:0 0 12px;font-size:20px;color:#111827;">%s</h1>`+
		`<p style="margin:0 0 8px;color:#374151;font-size:14px;">Hola %s,</p>`+
		`<p style="margin:0;color:#374151;font-size:15px;line-height:1.5;">%s</p>`+
		`</td></tr>`+
		`<tr><td style="padding:18px 28px;background:#fafafa;border-top:1px solid #e5e7eb;color:#9ca3af;font-size:12px;">`+
		`Este es un mensaje automático de ElectricAutomaticChile. No respondas a este correo.</td></tr>`+
		`</table></td></tr></table></body></html>`,
		etiqueta, titulo, nombreCliente, mensaje)
}

func (s *SESService) EnviarNotificacionTicket(destinatario, nombreUsuario, numeroTicket, asunto, mensaje string) error {
	return s.EnviarEmail([]string{destinatario}, fmt.Sprintf("Ticket #%s: %s", numeroTicket, asunto), mensaje,
		fmt.Sprintf("Hola %s, actualización del ticket #%s: %s", nombreUsuario, numeroTicket, mensaje))
}

func (s *SESService) EnviarNotificacionBoleta(destinatario, nombreCliente, numeroBoleta, monto, fechaVencimiento string) error {
	return s.EnviarEmail([]string{destinatario}, fmt.Sprintf("Nueva boleta #%s", numeroBoleta), "",
		fmt.Sprintf("Hola %s, boleta #%s por $%s vence el %s.", nombreCliente, numeroBoleta, monto, fechaVencimiento))
}

func (s *SESService) EnviarRecuperacionPassword(destinatario, nombreUsuario, token string) error {
	return s.EnviarEmail([]string{destinatario}, "Recuperación de contraseña", "",
		fmt.Sprintf("Hola %s, token de recuperación: %s", nombreUsuario, token))
}

func (s *SESService) EnviarBienvenida(destinatario, nombreUsuario, tipoUsuario string) error {
	return s.EnviarEmail([]string{destinatario}, "Bienvenido a Electricautomaticchile", "",
		fmt.Sprintf("Hola %s, tu cuenta %s fue creada.", nombreUsuario, tipoUsuario))
}

func (s *SESService) EnviarCredenciales(destinatario, nombreCliente, numeroCliente, passwordTemporal string) error {
	return s.EnviarEmail([]string{destinatario}, "Credenciales de acceso", "",
		fmt.Sprintf("Hola %s, número cliente: %s, password temporal: %s", nombreCliente, numeroCliente, passwordTemporal))
}

func (s *SESService) EnviarBoletaVenciendo(destinatario, nombreCliente, periodo, monto, fechaVencimiento string) error {
	return s.EnviarEmail([]string{destinatario}, fmt.Sprintf("Boleta de %s vence el %s", periodo, fechaVencimiento), "",
		fmt.Sprintf("Hola %s, tu boleta de %s por $%s vence el %s.", nombreCliente, periodo, monto, fechaVencimiento))
}

func (s *SESService) EnviarBoletaVencida(destinatario, nombreCliente, periodo, monto string) error {
	return s.EnviarEmail([]string{destinatario}, fmt.Sprintf("Boleta vencida: %s", periodo), "",
		fmt.Sprintf("Hola %s, tu boleta de %s por $%s ha vencido.", nombreCliente, periodo, monto))
}

func (s *SESService) EnviarAdvertenciaCorte(destinatario, nombreCliente string, numBoletas int, montoTotal string) error {
	return s.EnviarEmail([]string{destinatario}, fmt.Sprintf("Advertencia: %d boletas vencidas", numBoletas), "",
		fmt.Sprintf("Hola %s, tienes %d boletas vencidas por $%s.", nombreCliente, numBoletas, montoTotal))
}

func (s *SESService) EnviarAvisoCorte(destinatario, nombreCliente, titulo, mensaje string, numBoletas int, montoTotal string) error {
	return s.EnviarEmail([]string{destinatario}, titulo, "",
		fmt.Sprintf("Hola %s, %s Deuda total: $%s.", nombreCliente, mensaje, montoTotal))
}

func (s *SESService) EnviarServicioSuspendido(destinatario, nombreCliente string, numBoletas int, montoTotal string) error {
	return s.EnviarEmail([]string{destinatario}, "Servicio suspendido", "",
		fmt.Sprintf("Hola %s, tu servicio fue suspendido por %d boletas vencidas ($%s).", nombreCliente, numBoletas, montoTotal))
}

func (s *SESService) EnviarPagoConfirmado(destinatario, nombreCliente, periodo, monto string, servicioRepuesto bool) error {
	mensaje := fmt.Sprintf("Hola %s, tu pago de $%s (%s) fue recibido.", nombreCliente, monto, periodo)
	if servicioRepuesto {
		mensaje += " Tu suministro fue repuesto."
	}
	return s.EnviarEmail([]string{destinatario}, "Pago confirmado", "", mensaje)
}
