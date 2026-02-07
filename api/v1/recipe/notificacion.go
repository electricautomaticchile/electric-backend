package recipe

type CrearNotificacionRecipe struct {
	DestinatarioID string                 `json:"destinatarioId" binding:"required"`
	DispositivoID  string                 `json:"dispositivoId,omitempty"`
	Titulo         string                 `json:"titulo" binding:"required,max=200"`
	Mensaje        string                 `json:"mensaje" binding:"required,max=1000"`
	Tipo           string                 `json:"tipo" binding:"required,oneof=alerta ticket sistema consumo facturacion"`
	Severidad      string                 `json:"severidad,omitempty"`
	Importante     bool                   `json:"importante,omitempty"`
	Metadatos      map[string]interface{} `json:"metadatos,omitempty"`
}
