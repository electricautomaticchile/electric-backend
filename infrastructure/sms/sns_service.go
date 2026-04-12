package sms

import (
	"context"
	"electric-backend/config"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

const firma = "Electricautomaticchile"
const urlPago = "electricautomaticchile.com/cliente"

// SMSService define la interfaz para envío de SMS
type SMSService interface {
	EnviarSMS(to, mensaje string) error
	EnviarNotificacionConsumo(telefono, nombreCliente string, consumoKWh, costoEstimado float64, diasTranscurridos int) error
	EnviarAvisoCorteServicio(telefono, nombreCliente string, boletasImpagas int, montoTotal float64) error
	EnviarConfirmacionPago(telefono, nombreCliente, numeroBoleta string, monto float64) error
}

// SNSService implementa SMSService usando Amazon SNS
type SNSService struct {
	client *sns.Client
}

func NewSNSService() (*SNSService, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(config.AppConfig.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS para SNS: %w", err)
	}

	return &SNSService{
		client: sns.NewFromConfig(awsCfg),
	}, nil
}

func (s *SNSService) EnviarSMS(to, mensaje string) error {
	if !strings.HasPrefix(to, "+") {
		to = "+56" + strings.TrimPrefix(to, "56")
	}

	_, err := s.client.Publish(context.TODO(), &sns.PublishInput{
		PhoneNumber: &to,
		Message:     &mensaje,
	})
	if err != nil {
		return fmt.Errorf("error enviando SMS via SNS: %w", err)
	}

	return nil
}

// Resumen quincenal de consumo (día 1 y 15 del mes)
func (s *SNSService) EnviarNotificacionConsumo(telefono, nombreCliente string, consumoKWh, costoEstimado float64, diasTranscurridos int) error {
	mensaje := fmt.Sprintf(
		"Hola %s, tu resumen de consumo (%d dias):\n"+
			"⚡ %.1f kWh consumidos\n"+
			"💰 $%s estimado este mes\n"+
			"Revisa tu detalle en %s\n"+
			"- %s",
		nombreCliente,
		diasTranscurridos,
		consumoKWh,
		formatCLP(costoEstimado),
		urlPago,
		firma,
	)
	return s.EnviarSMS(telefono, mensaje)
}

// Aviso de corte por boletas impagas
func (s *SNSService) EnviarAvisoCorteServicio(telefono, nombreCliente string, boletasImpagas int, montoTotal float64) error {
	mensaje := fmt.Sprintf(
		"⚠️ AVISO: Hola %s, tienes %d boleta(s) vencida(s) por $%s.\n"+
			"Tu suministro electrico sera suspendido si no regularizas tu deuda.\n"+
			"Paga en: %s\n"+
			"- %s",
		nombreCliente,
		boletasImpagas,
		formatCLP(montoTotal),
		urlPago,
		firma,
	)
	return s.EnviarSMS(telefono, mensaje)
}

// Confirmación de pago recibido
func (s *SNSService) EnviarConfirmacionPago(telefono, nombreCliente, periodo string, monto float64) error {
	mensaje := fmt.Sprintf(
		"✅ Pago confirmado\n"+
			"Hola %s, recibimos tu pago de $%s correspondiente a %s.\n"+
			"Gracias por regularizar tu cuenta.\n"+
			"- %s",
		nombreCliente,
		formatCLP(monto),
		periodo,
		firma,
	)
	return s.EnviarSMS(telefono, mensaje)
}

// formatCLP formatea un número como pesos chilenos sin decimales con separador de miles
func formatCLP(monto float64) string {
	n := int64(monto)
	s := fmt.Sprintf("%d", n)
	// Agregar puntos como separador de miles
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return result
}
