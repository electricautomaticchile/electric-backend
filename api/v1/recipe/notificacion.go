package recipe

type CrearNotificacionRecipe struct {
	DestinatarioID string `json:"destinatarioId" binding:"required"`
	Titulo         string `json:"titulo" binding:"required,max=200"`
	Mensaje        string `json:"mensaje" binding:"required,max=1000"`
	Tipo           string `json:"tipo" binding:"required,oneof=alerta ticket sistema consumo facturacion"`
}
