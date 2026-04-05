package sms

import (
	"context"
	"electric-backend/config"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

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

func (s *SNSService) EnviarNotificacionConsumo(telefono, nombreCliente string, consumoKWh, costoEstimado float64, diasTranscurridos int) error {
	mensaje := fmt.Sprintf(
		"Hola %s, llevas %d dias de consumo:\n"+
			"⚡ %.2f kWh\n"+
			"💰 $%.0f aprox.\n"+
			"Reduce tu consumo para ahorrar.\n"+
			"- Electricatomaticchile",
		nombreCliente, diasTranscurridos, consumoKWh, costoEstimado,
	)
	return s.EnviarSMS(telefono, mensaje)
}

func (s *SNSService) EnviarAvisoCorteServicio(telefono, nombreCliente string, boletasImpagas int, montoTotal float64) error {
	mensaje := fmt.Sprintf(
		"⚠️ AVISO IMPORTANTE\n"+
			"Hola %s, tienes %d boletas impagas por $%.0f.\n"+
			"Tu servicio sera cortado pronto.\n"+
			"Paga en: electricautomaticchile.com/cliente\n"+
			"- Electricautomaticchile",
		nombreCliente, boletasImpagas, montoTotal,
	)
	return s.EnviarSMS(telefono, mensaje)
}

func (s *SNSService) EnviarConfirmacionPago(telefono, nombreCliente, numeroBoleta string, monto float64) error {
	mensaje := fmt.Sprintf(
		"✅ Pago confirmado\n"+
			"Hola %s, recibimos tu pago de $%.0f para la boleta #%s.\n"+
			"Gracias por tu preferencia.\n"+
			"- Electricautomaticchile",
		nombreCliente, monto, numeroBoleta,
	)
	return s.EnviarSMS(telefono, mensaje)
}
