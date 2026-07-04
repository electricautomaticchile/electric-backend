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
