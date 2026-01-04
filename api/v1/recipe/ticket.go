package recipe

type CrearTicketRecipe struct {
	Asunto      string `json:"asunto" binding:"required"`
	Descripcion string `json:"descripcion" binding:"required"`
	Categoria   string `json:"categoria"`
	Prioridad   string `json:"prioridad"`
	ClienteID   string `json:"clienteId,omitempty"`
	EmpresaID   string `json:"empresaId,omitempty"`
}

type AgregarRespuestaRecipe struct {
	Mensaje string `json:"mensaje" binding:"required"`
}

type ActualizarEstadoTicketRecipe struct {
	Estado string `json:"estado" binding:"required"`
}
