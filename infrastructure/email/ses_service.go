package email

import (
	"bytes"
	"context"
	"electric-backend/config"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type SESService struct {
	client    *ses.Client
	from      string
	templates map[string]*template.Template
}

func NewSESService() *SESService {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(config.AppConfig.AWSRegion),
	)
	if err != nil {
		fmt.Printf("⚠️ SES: error cargando config AWS: %v\n", err)
	}

	service := &SESService{
		client:    ses.NewFromConfig(awsCfg),
		from:      config.AppConfig.EmailFrom,
		templates: make(map[string]*template.Template),
	}
	service.loadTemplates()
	return service
}

func (s *SESService) loadTemplates() {
	templatePath := "infrastructure/email/templates"
	for _, file := range []string{"alerta.html", "ticket.html", "boleta.html", "recuperacion.html", "bienvenida.html", "credenciales.html"} {
		name := file[:len(file)-5]
		if tmpl, err := template.ParseFiles(filepath.Join(templatePath, file)); err == nil {
			s.templates[name] = tmpl
		}
	}
}

func (s *SESService) renderTemplate(name string, data interface{}) (string, error) {
	tmpl, exists := s.templates[name]
	if !exists {
		return "", fmt.Errorf("template %s no encontrado", name)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *SESService) EnviarEmail(to []string, subject, htmlBody, textBody string) error {
	toAddresses := make([]string, len(to))
	copy(toAddresses, to)

	input := &ses.SendEmailInput{
		Source: aws.String(s.from),
		Destination: &types.Destination{
			ToAddresses: toAddresses,
		},
		Message: &types.Message{
			Subject: &types.Content{Data: aws.String(subject)},
			Body: &types.Body{
				Html: &types.Content{Data: aws.String(htmlBody)},
				Text: &types.Content{Data: aws.String(textBody)},
			},
		},
	}

	_, err := s.client.SendEmail(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error enviando email via SES: %w", err)
	}
	return nil
}

func (s *SESService) EnviarNotificacionAlerta(destinatario, nombreCliente, tipoAlerta, mensaje string) error {
	subject := fmt.Sprintf("⚠️ Alerta: %s", tipoAlerta)
	data := map[string]string{"NombreCliente": nombreCliente, "TipoAlerta": tipoAlerta, "Mensaje": mensaje}
	html, err := s.renderTemplate("alerta", data)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("Hola %s,\n\nSe ha generado una nueva alerta: %s\n%s\n\nRevisa tu dashboard en: https://electricautomaticchile.com/cliente", nombreCliente, tipoAlerta, mensaje)
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}

func (s *SESService) EnviarNotificacionTicket(destinatario, nombreUsuario, numeroTicket, asunto, mensaje string) error {
	subject := fmt.Sprintf("🎫 Ticket #%s: %s", numeroTicket, asunto)
	data := map[string]string{"NombreUsuario": nombreUsuario, "NumeroTicket": numeroTicket, "Asunto": asunto, "Mensaje": mensaje}
	html, err := s.renderTemplate("ticket", data)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("Hola %s,\n\nActualización en ticket #%s: %s\n\n%s\n\nRevisa tu ticket en: https://electricautomaticchile.com/cliente", nombreUsuario, numeroTicket, asunto, mensaje)
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}

func (s *SESService) EnviarNotificacionBoleta(destinatario, nombreCliente, numeroBoleta, monto, fechaVencimiento string) error {
	subject := fmt.Sprintf("💰 Nueva Boleta #%s - $%s", numeroBoleta, monto)
	data := map[string]string{"NombreCliente": nombreCliente, "NumeroBoleta": numeroBoleta, "Monto": monto, "FechaVencimiento": fechaVencimiento}
	html, err := s.renderTemplate("boleta", data)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("Hola %s,\n\nNueva boleta generada:\nNúmero: #%s\nMonto: $%s\nVencimiento: %s\n\nRevisa tu boleta en: https://electricautomaticchile.com/cliente", nombreCliente, numeroBoleta, monto, fechaVencimiento)
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}

func (s *SESService) EnviarRecuperacionPassword(destinatario, nombreUsuario, token string) error {
	subject := "🔐 Recuperación de Contraseña"
	resetURL := fmt.Sprintf("https://electricautomaticchile.com/reset-password?token=%s", token)
	data := map[string]string{"NombreUsuario": nombreUsuario, "ResetURL": resetURL}
	html, err := s.renderTemplate("recuperacion", data)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("Hola %s,\n\nRecibimos una solicitud para restablecer tu contraseña.\n\nHaz clic en el siguiente enlace:\n%s\n\nEste enlace expirará en 1 hora.", nombreUsuario, resetURL)
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}

func (s *SESService) EnviarBienvenida(destinatario, nombreUsuario, tipoUsuario string) error {
	subject := "🎉 Bienvenido a Electric Automatic Chile"
	dashboardURL := "https://electricautomaticchile.com/cliente"
	if tipoUsuario == "empresa" {
		dashboardURL = "https://electricautomaticchile.com/empresa"
	}
	data := map[string]string{"NombreUsuario": nombreUsuario, "DashboardURL": dashboardURL}
	html, err := s.renderTemplate("bienvenida", data)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("Hola %s,\n\n¡Bienvenido a Electric Automatic Chile!\n\nAccede a tu dashboard en: %s", nombreUsuario, dashboardURL)
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}

func (s *SESService) EnviarCredenciales(destinatario, nombreCliente, numeroCliente, passwordTemporal string) error {
	subject := "🔑 Credenciales de Acceso - Electric Automatic Chile"
	data := map[string]string{"NombreCliente": nombreCliente, "NumeroCliente": numeroCliente, "PasswordTemporal": passwordTemporal}
	html, err := s.renderTemplate("credenciales", data)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("Hola %s,\n\nTu cuenta ha sido creada.\n\nNúmero de Cliente: %s\nContraseña Temporal: %s\n\nInicia sesión en: https://electricautomaticchile.com/login", nombreCliente, numeroCliente, passwordTemporal)
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}
