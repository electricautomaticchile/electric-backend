package recipe

type CrearTicketRecipe struct {
	Titulo      string `json:"titulo" binding:"required"`
	Descripcion string `json:"descripcion" binding:"required"`
	Prioridad   string `json:"prioridad" binding:"required"`
	ClienteID   string `json:"clienteId,omitempty"`
}

type AgregarRespuestaRecipe struct {
	Mensaje string `json:"mensaje" binding:"required"`
}

type ActualizarEstadoTicketRecipe struct {
	Estado string `json:"estado" binding:"required"`
}
