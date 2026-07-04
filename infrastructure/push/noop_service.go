package push

import "log"

// NoopService es el adaptador de push usado cuando no hay credenciales de
// Firebase disponibles. Registra la operación pero no envía nada, permitiendo
// que el arranque del servidor no falle (igual que los adaptadores no-op de
// email y SMS).
type NoopService struct{}

func NewNoopService() *NoopService {
	log.Println("push: Firebase deshabilitado; usando no-op")
	return &NoopService{}
}

func (s *NoopService) EnviarPush(tokens []string, titulo, cuerpo string, data map[string]string) error {
	log.Printf("push noop: omitido envio a %d token(s) titulo=%q", len(tokens), titulo)
	return nil
}
