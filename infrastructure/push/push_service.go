package push

import "context"

// PushService define el contrato para el envío de notificaciones push.
// EnviarPush envía una notificación a la lista de tokens indicada. Los tokens
// que resulten inválidos (no registrados) son eliminados automáticamente.
type PushService interface {
	EnviarPush(tokens []string, titulo, cuerpo string, data map[string]string) error
}

// TokenEliminador permite al servicio de push limpiar tokens inválidos.
// Se satisface con el repositorio de tokens FCM sin generar dependencias
// circulares entre paquetes.
type TokenEliminador interface {
	DeleteByToken(ctx context.Context, token string) error
}
