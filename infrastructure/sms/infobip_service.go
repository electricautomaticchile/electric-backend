package sms

import (
	"bytes"
	"electric-backend/config"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type InfobipService struct {
	apiKey     string
	baseURL    string
	fromNumber string
	client     *http.Client
}

type InfobipMessage struct {
	From         string                   `json:"from"`
	Destinations []InfobipDestination     `json:"destinations"`
	Text         string                   `json:"text"`
}

type InfobipDestination struct {
	To string `json:"to"`
}

type InfobipRequest struct {
	Messages []InfobipMessage `json:"messages"`
}

type InfobipResponse struct {
	Messages []struct {
		MessageID string `json:"messageId"`
		Status    struct {
			GroupID     int    `json:"groupId"`
			GroupName   string `json:"groupName"`
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"status"`
	} `json:"messages"`
}

func NewInfobipService() *InfobipService {
	return &InfobipService{
		apiKey:     config.AppConfig.InfobipAPIKey,
		baseURL:    config.AppConfig.InfobipBaseURL,
		fromNumber: config.AppConfig.SMSFromNumber,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *InfobipService) EnviarSMS(to, mensaje string) error {
	if s.apiKey == "" {
		return fmt.Errorf("credenciales de Infobip no configuradas")
	}

	if !strings.HasPrefix(to, "+") {
		to = "+56" + strings.TrimPrefix(to, "56")
	}

	reqBody := InfobipRequest{
		Messages: []InfobipMessage{
			{
				From: s.fromNumber,
				Destinations: []InfobipDestination{
					{To: to},
				},
				Text: mensaje,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error serializando request: %w", err)
	}

	apiURL := fmt.Sprintf("%s/sms/2/text/advanced", s.baseURL)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Authorization", "App "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error enviando SMS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error de Infobip (status %d)", resp.StatusCode)
	}

	var infobipResp InfobipResponse
	if err := json.NewDecoder(resp.Body).Decode(&infobipResp); err != nil {
		return fmt.Errorf("error parseando respuesta: %w", err)
	}

	if len(infobipResp.Messages) == 0 {
		return fmt.Errorf("no se recibió respuesta del servidor")
	}

	return nil
}

func (s *InfobipService) EnviarNotificacionConsumo(telefono, nombreCliente string, consumoKWh float64, costoEstimado float64, diasTranscurridos int) error {
	mensaje := fmt.Sprintf(
		"Hola %s, llevas %d dias de consumo:\n"+
			"⚡ %.2f kWh\n"+
			"💰 $%.0f aprox.\n"+
			"Reduce tu consumo para ahorrar.\n"+
			"- Electric Automatic Chile",
		nombreCliente,
		diasTranscurridos,
		consumoKWh,
		costoEstimado,
	)

	return s.EnviarSMS(telefono, mensaje)
}

func (s *InfobipService) EnviarAvisoCorteServicio(telefono, nombreCliente string, boletasImpagas int, montoTotal float64) error {
	mensaje := fmt.Sprintf(
		"⚠️ AVISO IMPORTANTE\n"+
			"Hola %s, tienes %d boletas impagas por $%.0f.\n"+
			"Tu servicio sera cortado pronto.\n"+
			"Paga en: electricautomaticchile.com/cliente\n"+
			"- Electric Automatic Chile",
		nombreCliente,
		boletasImpagas,
		montoTotal,
	)

	return s.EnviarSMS(telefono, mensaje)
}

func (s *InfobipService) EnviarConfirmacionPago(telefono, nombreCliente, numeroBoleta string, monto float64) error {
	mensaje := fmt.Sprintf(
		"✅ Pago confirmado\n"+
			"Hola %s, recibimos tu pago de $%.0f para la boleta #%s.\n"+
			"Gracias por tu preferencia.\n"+
			"- Electric Automatic Chile",
		nombreCliente,
		monto,
		numeroBoleta,
	)

	return s.EnviarSMS(telefono, mensaje)
}
