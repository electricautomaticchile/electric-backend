package sms

import (
	"fmt"
	"log"
)

const firma = "Electricautomaticchile"
const urlPago = "electricautomaticchile.com/cliente"

type SMSService interface {
	EnviarSMS(to, mensaje string) error
	EnviarNotificacionConsumo(telefono, nombreCliente string, consumoKWh, costoEstimado float64, diasTranscurridos int) error
	EnviarAvisoCorteServicio(telefono, nombreCliente string, boletasImpagas int, montoTotal float64) error
	EnviarConfirmacionPago(telefono, nombreCliente, numeroBoleta string, monto float64) error
}

type NoopService struct{}

func NewNoopService() *NoopService {
	log.Println("sms: proveedor externo deshabilitado; usando no-op")
	return &NoopService{}
}

func (s *NoopService) EnviarSMS(to, mensaje string) error {
	log.Printf("sms noop: omitido envio a=%s mensaje=%q", to, mensaje)
	return nil
}

func (s *NoopService) EnviarNotificacionConsumo(telefono, nombreCliente string, consumoKWh, costoEstimado float64, diasTranscurridos int) error {
	mensaje := fmt.Sprintf(
		"Hola %s, tu resumen de consumo (%d dias): %.1f kWh, $%s estimado. Revisa %s - %s",
		nombreCliente,
		diasTranscurridos,
		consumoKWh,
		formatCLP(costoEstimado),
		urlPago,
		firma,
	)
	return s.EnviarSMS(telefono, mensaje)
}

func (s *NoopService) EnviarAvisoCorteServicio(telefono, nombreCliente string, boletasImpagas int, montoTotal float64) error {
	mensaje := fmt.Sprintf(
		"Aviso: Hola %s, tienes %d boleta(s) vencida(s) por $%s. Paga en %s - %s",
		nombreCliente,
		boletasImpagas,
		formatCLP(montoTotal),
		urlPago,
		firma,
	)
	return s.EnviarSMS(telefono, mensaje)
}

func (s *NoopService) EnviarConfirmacionPago(telefono, nombreCliente, periodo string, monto float64) error {
	mensaje := fmt.Sprintf(
		"Pago confirmado: Hola %s, recibimos tu pago de $%s correspondiente a %s. - %s",
		nombreCliente,
		formatCLP(monto),
		periodo,
		firma,
	)
	return s.EnviarSMS(telefono, mensaje)
}

func formatCLP(monto float64) string {
	n := int64(monto)
	s := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return result
}
