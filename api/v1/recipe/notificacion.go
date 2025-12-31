package recipe

type CrearNotificacionRecipe struct {
	DestinatarioID string `json:"destinatarioId" binding:"required"`
	Titulo         string `json:"titulo" binding:"required"`
	Mensaje        string `json:"mensaje" binding:"required"`
	Tipo           string `json:"tipo" binding:"required"`
}
