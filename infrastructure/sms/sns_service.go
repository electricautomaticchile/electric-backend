package sms

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
)

// SNSService envía SMS transaccionales usando AWS SNS.
//
// Credenciales: cadena estándar de AWS (AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY
// / AWS_REGION, o rol IAM). Los números deben ir en formato E.164 (+569XXXXXXXX).
type SNSService struct {
	client    *sns.Client
	senderID  string
	smsType   string
}

// NewSNSService crea el servicio SNS. region por defecto us-east-1.
func NewSNSService(senderID, region string) (*SNSService, error) {
	if region == "" {
		region = "us-east-1"
	}
	if senderID == "" {
		senderID = "ElectricCL"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("cargando config AWS: %w", err)
	}

	log.Printf("sms: proveedor AWS SNS activo (region=%s, sender=%s)", region, senderID)
	return &SNSService{client: sns.NewFromConfig(cfg), senderID: senderID, smsType: "Transactional"}, nil
}

// EnviarSMS publica un SMS a un número E.164 vía SNS.
func (s *SNSService) EnviarSMS(to, mensaje string) error {
	if to == "" {
		return fmt.Errorf("teléfono vacío")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := s.client.Publish(ctx, &sns.PublishInput{
		PhoneNumber: aws.String(to),
		Message:     aws.String(mensaje),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"AWS.SNS.SMS.SMSType":  {DataType: aws.String("String"), StringValue: aws.String(s.smsType)},
			"AWS.SNS.SMS.SenderID": {DataType: aws.String("String"), StringValue: aws.String(s.senderID)},
		},
	})
	if err != nil {
		return fmt.Errorf("SNS Publish: %w", err)
	}
	return nil
}

func (s *SNSService) EnviarNotificacionConsumo(telefono, nombreCliente string, consumoKWh, costoEstimado float64, diasTranscurridos int) error {
	mensaje := fmt.Sprintf(
		"Hola %s, tu resumen de consumo (%d dias): %.1f kWh, $%s estimado. Revisa %s - %s",
		nombreCliente, diasTranscurridos, consumoKWh, formatCLP(costoEstimado), urlPago, firma,
	)
	return s.EnviarSMS(telefono, mensaje)
}

func (s *SNSService) EnviarAvisoCorteServicio(telefono, nombreCliente string, boletasImpagas int, montoTotal float64) error {
	mensaje := fmt.Sprintf(
		"Aviso: Hola %s, tienes %d boleta(s) vencida(s) por $%s. Paga en %s - %s",
		nombreCliente, boletasImpagas, formatCLP(montoTotal), urlPago, firma,
	)
	return s.EnviarSMS(telefono, mensaje)
}

func (s *SNSService) EnviarConfirmacionPago(telefono, nombreCliente, periodo string, monto float64) error {
	mensaje := fmt.Sprintf(
		"Pago confirmado: Hola %s, recibimos tu pago de $%s correspondiente a %s. - %s",
		nombreCliente, formatCLP(monto), periodo, firma,
	)
	return s.EnviarSMS(telefono, mensaje)
}
