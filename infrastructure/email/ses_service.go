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
	subject := "🎉 Bienvenido a Electricautomaticchile"
	dashboardURL := "https://electricautomaticchile.com/cliente"
	if tipoUsuario == "empresa" {
		dashboardURL = "https://electricautomaticchile.com/empresa"
	}
	data := map[string]string{"NombreUsuario": nombreUsuario, "DashboardURL": dashboardURL}
	html, err := s.renderTemplate("bienvenida", data)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("Hola %s,\n\n¡Bienvenido a Electricautomaticchile!\n\nAccede a tu dashboard en: %s", nombreUsuario, dashboardURL)
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}

func (s *SESService) EnviarCredenciales(destinatario, nombreCliente, numeroCliente, passwordTemporal string) error {
	subject := "🔑 Credenciales de Acceso - Electricautomaticchile"
	data := map[string]string{"NombreCliente": nombreCliente, "NumeroCliente": numeroCliente, "PasswordTemporal": passwordTemporal}
	html, err := s.renderTemplate("credenciales", data)
	if err != nil {
		return err
	}
	text := fmt.Sprintf("Hola %s,\n\nTu cuenta ha sido creada.\n\nNúmero de Cliente: %s\nContraseña Temporal: %s\n\nInicia sesión en: https://electricautomaticchile.com/login", nombreCliente, numeroCliente, passwordTemporal)
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}

func (s *SESService) EnviarBoletaVenciendo(destinatario, nombreCliente, periodo, monto, fechaVencimiento string) error {
	data := map[string]string{"NombreCliente": nombreCliente, "Periodo": periodo, "Monto": monto, "FechaVencimiento": fechaVencimiento}
	html, err := s.renderTemplate("boleta_venciendo", data)
	if err != nil {
		return s.EnviarEmail([]string{destinatario}, fmt.Sprintf("⚠️ Tu boleta de %s vence el %s", periodo, fechaVencimiento),
			fmt.Sprintf("Hola %s, tu boleta de %s por $%s vence el %s.", nombreCliente, periodo, monto, fechaVencimiento),
			fmt.Sprintf("Hola %s, tu boleta de %s por $%s vence el %s.", nombreCliente, periodo, monto, fechaVencimiento))
	}
	text := fmt.Sprintf("Hola %s, tu boleta de %s por $%s vence el %s. Paga en: https://electricautomaticchile.com/cliente", nombreCliente, periodo, monto, fechaVencimiento)
	return s.EnviarEmail([]string{destinatario}, fmt.Sprintf("⚠️ Tu boleta de %s vence el %s", periodo, fechaVencimiento), html, text)
}

func (s *SESService) EnviarBoletaVencida(destinatario, nombreCliente, periodo, monto string) error {
	data := map[string]string{"NombreCliente": nombreCliente, "Periodo": periodo, "Monto": monto}
	html, err := s.renderTemplate("boleta_vencida", data)
	if err != nil {
		return s.EnviarEmail([]string{destinatario}, fmt.Sprintf("🔴 Boleta vencida: %s", periodo),
			fmt.Sprintf("Hola %s, tu boleta de %s por $%s ha vencido.", nombreCliente, periodo, monto),
			fmt.Sprintf("Hola %s, tu boleta de %s por $%s ha vencido.", nombreCliente, periodo, monto))
	}
	text := fmt.Sprintf("Hola %s, tu boleta de %s por $%s ha vencido. Paga en: https://electricautomaticchile.com/cliente", nombreCliente, periodo, monto)
	return s.EnviarEmail([]string{destinatario}, fmt.Sprintf("🔴 Boleta vencida: %s", periodo), html, text)
}

func (s *SESService) EnviarAdvertenciaCorte(destinatario, nombreCliente string, numBoletas int, montoTotal string) error {
	data := map[string]interface{}{"NombreCliente": nombreCliente, "NumBoletas": numBoletas, "MontoTotal": montoTotal}
	html, err := s.renderTemplate("advertencia_corte", data)
	if err != nil {
		return s.EnviarEmail([]string{destinatario}, fmt.Sprintf("⚠️ Advertencia: %d boletas vencidas", numBoletas),
			fmt.Sprintf("Hola %s, tienes %d boletas vencidas por $%s.", nombreCliente, numBoletas, montoTotal),
			fmt.Sprintf("Hola %s, tienes %d boletas vencidas por $%s.", nombreCliente, numBoletas, montoTotal))
	}
	text := fmt.Sprintf("Hola %s, tienes %d boletas vencidas por $%s. Al tercer impago se suspenderá tu suministro.", nombreCliente, numBoletas, montoTotal)
	return s.EnviarEmail([]string{destinatario}, fmt.Sprintf("⚠️ Advertencia: %d boletas vencidas — riesgo de corte", numBoletas), html, text)
}

func (s *SESService) EnviarAvisoCorte(destinatario, nombreCliente, titulo, mensaje string, numBoletas int, montoTotal string) error {
	data := map[string]interface{}{"NombreCliente": nombreCliente, "Titulo": titulo, "Mensaje": mensaje, "NumBoletas": numBoletas, "MontoTotal": montoTotal}
	html, err := s.renderTemplate("aviso_corte", data)
	if err != nil {
		return s.EnviarEmail([]string{destinatario}, "🔴 "+titulo,
			fmt.Sprintf("Hola %s, %s Deuda: $%s.", nombreCliente, mensaje, montoTotal),
			fmt.Sprintf("Hola %s, %s Deuda: $%s.", nombreCliente, mensaje, montoTotal))
	}
	text := fmt.Sprintf("Hola %s, %s Deuda total: $%s. Paga en: https://electricautomaticchile.com/cliente", nombreCliente, mensaje, montoTotal)
	return s.EnviarEmail([]string{destinatario}, "🔴 "+titulo, html, text)
}

func (s *SESService) EnviarServicioSuspendido(destinatario, nombreCliente string, numBoletas int, montoTotal string) error {
	data := map[string]interface{}{"NombreCliente": nombreCliente, "NumBoletas": numBoletas, "MontoTotal": montoTotal}
	html, err := s.renderTemplate("servicio_suspendido", data)
	if err != nil {
		return s.EnviarEmail([]string{destinatario}, "🔴 Tu suministro eléctrico fue suspendido",
			fmt.Sprintf("Hola %s, tu suministro fue suspendido por %d boletas vencidas ($%s).", nombreCliente, numBoletas, montoTotal),
			fmt.Sprintf("Hola %s, tu suministro fue suspendido por %d boletas vencidas ($%s).", nombreCliente, numBoletas, montoTotal))
	}
	text := fmt.Sprintf("Hola %s, tu suministro fue suspendido por %d boletas vencidas ($%s). Paga en: https://electricautomaticchile.com/cliente", nombreCliente, numBoletas, montoTotal)
	return s.EnviarEmail([]string{destinatario}, "🔴 Tu suministro eléctrico fue suspendido", html, text)
}

func (s *SESService) EnviarPagoConfirmado(destinatario, nombreCliente, periodo, monto string, servicioRepuesto bool) error {
	data := map[string]interface{}{"NombreCliente": nombreCliente, "Periodo": periodo, "Monto": monto, "ServicioRepuesto": servicioRepuesto}
	html, err := s.renderTemplate("pago_confirmado", data)
	if err != nil {
		msg := fmt.Sprintf("Hola %s, tu pago de $%s (%s) fue recibido.", nombreCliente, monto, periodo)
		return s.EnviarEmail([]string{destinatario}, "✅ Pago confirmado", msg, msg)
	}
	text := fmt.Sprintf("Hola %s, tu pago de $%s (%s) fue recibido correctamente.", nombreCliente, monto, periodo)
	if servicioRepuesto {
		text += " Tu suministro eléctrico fue repuesto automáticamente."
	}
	return s.EnviarEmail([]string{destinatario}, fmt.Sprintf("✅ Pago confirmado — %s", periodo), html, text)
}
