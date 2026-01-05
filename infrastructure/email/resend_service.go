package email

import (
	"bytes"
	"electric-backend/config"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"time"
)

type ResendService struct {
	apiKey    string
	from      string
	client    *http.Client
	templates map[string]*template.Template
}

type EmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
	Text    string   `json:"text,omitempty"`
}

type EmailResponse struct {
	ID string `json:"id"`
}

func NewResendService() *ResendService {
	service := &ResendService{
		apiKey:    config.AppConfig.ResendAPIKey,
		from:      config.AppConfig.EmailFrom,
		client:    &http.Client{Timeout: 10 * time.Second},
		templates: make(map[string]*template.Template),
	}
	
	service.loadTemplates()
	return service
}

func (s *ResendService) loadTemplates() {
	templatePath := "infrastructure/email/templates"
	
	templateFiles := []string{
		"alerta.html",
		"ticket.html",
		"boleta.html",
		"recuperacion.html",
		"bienvenida.html",
		"credenciales.html",
	}
	
	for _, file := range templateFiles {
		name := file[:len(file)-5]
		tmpl, err := template.ParseFiles(filepath.Join(templatePath, file))
		if err == nil {
			s.templates[name] = tmpl
		}
	}
}

func (s *ResendService) renderTemplate(name string, data interface{}) (string, error) {
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

func (s *ResendService) EnviarEmail(to []string, subject, htmlBody, textBody string) error {
	if s.apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY no configurado")
	}

	emailReq := EmailRequest{
		From:    s.from,
		To:      to,
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	}

	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return fmt.Errorf("error serializando email: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error enviando email: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error de Resend (status %d): %s", resp.StatusCode, string(body))
	}

	var emailResp EmailResponse
	if err := json.Unmarshal(body, &emailResp); err != nil {
		return fmt.Errorf("error parseando respuesta: %w", err)
	}

	return nil
}

func (s *ResendService) EnviarNotificacionAlerta(destinatario, nombreCliente, tipoAlerta, mensaje string) error {
	subject := fmt.Sprintf("⚠️ Alerta: %s", tipoAlerta)
	
	data := map[string]string{
		"NombreCliente": nombreCliente,
		"TipoAlerta":    tipoAlerta,
		"Mensaje":       mensaje,
	}
	
	html, err := s.renderTemplate("alerta", data)
	if err != nil {
		return err
	}
	
	text := fmt.Sprintf("Hola %s,\n\nSe ha generado una nueva alerta: %s\n%s\n\nRevisa tu dashboard en: https://electricautomaticchile.com/cliente", nombreCliente, tipoAlerta, mensaje)
	
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}

func (s *ResendService) EnviarNotificacionTicket(destinatario, nombreUsuario, numeroTicket, asunto, mensaje string) error {
	subject := fmt.Sprintf("🎫 Ticket #%s: %s", numeroTicket, asunto)
	
	data := map[string]string{
		"NombreUsuario": nombreUsuario,
		"NumeroTicket":  numeroTicket,
		"Asunto":        asunto,
		"Mensaje":       mensaje,
	}
	
	html, err := s.renderTemplate("ticket", data)
	if err != nil {
		return err
	}
	
	text := fmt.Sprintf("Hola %s,\n\nActualización en ticket #%s: %s\n\n%s\n\nRevisa tu ticket en: https://electricautomaticchile.com/cliente", nombreUsuario, numeroTicket, asunto, mensaje)
	
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}

func (s *ResendService) EnviarNotificacionBoleta(destinatario, nombreCliente, numeroBoleta, monto, fechaVencimiento string) error {
	subject := fmt.Sprintf("💰 Nueva Boleta #%s - $%s", numeroBoleta, monto)
	
	data := map[string]string{
		"NombreCliente":    nombreCliente,
		"NumeroBoleta":     numeroBoleta,
		"Monto":            monto,
		"FechaVencimiento": fechaVencimiento,
	}
	
	html, err := s.renderTemplate("boleta", data)
	if err != nil {
		return err
	}
	
	text := fmt.Sprintf("Hola %s,\n\nNueva boleta generada:\nNúmero: #%s\nMonto: $%s\nVencimiento: %s\n\nRevisa tu boleta en: https://electricautomaticchile.com/cliente", nombreCliente, numeroBoleta, monto, fechaVencimiento)
	
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}

func (s *ResendService) EnviarRecuperacionPassword(destinatario, nombreUsuario, token string) error {
	subject := "🔐 Recuperación de Contraseña"
	resetURL := fmt.Sprintf("https://electricautomaticchile.com/reset-password?token=%s", token)
	
	data := map[string]string{
		"NombreUsuario": nombreUsuario,
		"ResetURL":      resetURL,
	}
	
	html, err := s.renderTemplate("recuperacion", data)
	if err != nil {
		return err
	}
	
	text := fmt.Sprintf("Hola %s,\n\nRecibimos una solicitud para restablecer tu contraseña.\n\nHaz clic en el siguiente enlace para crear una nueva contraseña:\n%s\n\nEste enlace expirará en 1 hora.\n\nSi no solicitaste este cambio, ignora este correo.", nombreUsuario, resetURL)
	
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}

func (s *ResendService) EnviarBienvenida(destinatario, nombreUsuario, tipoUsuario string) error {
	subject := "🎉 Bienvenido a Electric Automatic Chile"
	dashboardURL := "https://electricautomaticchile.com/cliente"
	if tipoUsuario == "empresa" {
		dashboardURL = "https://electricautomaticchile.com/empresa"
	}
	
	data := map[string]string{
		"NombreUsuario": nombreUsuario,
		"DashboardURL":  dashboardURL,
	}
	
	html, err := s.renderTemplate("bienvenida", data)
	if err != nil {
		return err
	}
	
	text := fmt.Sprintf("Hola %s,\n\n¡Bienvenido a Electric Automatic Chile!\n\nTu cuenta ha sido creada exitosamente.\n\nAccede a tu dashboard en: %s\n\n¿Necesitas ayuda? Contáctanos en soporte@electricautomaticchile.com", nombreUsuario, dashboardURL)
	
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}

func (s *ResendService) EnviarCredenciales(destinatario, nombreCliente, numeroCliente, passwordTemporal string) error {
	subject := "🔑 Credenciales de Acceso - Electric Automatic Chile"
	
	data := map[string]string{
		"NombreCliente":    nombreCliente,
		"NumeroCliente":    numeroCliente,
		"PasswordTemporal": passwordTemporal,
	}
	
	html, err := s.renderTemplate("credenciales", data)
	if err != nil {
		return err
	}
	
	text := fmt.Sprintf(`Hola %s,

Tu cuenta en Electric Automatic Chile ha sido creada exitosamente.

CREDENCIALES DE ACCESO:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Número de Cliente: %s
Contraseña Temporal: %s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️ IMPORTANTE - SEGURIDAD:
• Esta es una contraseña temporal
• Debes cambiarla en tu primer inicio de sesión
• No compartas estas credenciales con nadie
• Guarda esta información en un lugar seguro

Inicia sesión en: https://electricautomaticchile.com/login

¿Necesitas ayuda? Contáctanos en soporte@electricautomaticchile.com`, nombreCliente, numeroCliente, passwordTemporal)
	
	return s.EnviarEmail([]string{destinatario}, subject, html, text)
}
