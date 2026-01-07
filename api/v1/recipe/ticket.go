package recipe

type CrearTicketRecipe struct {
	Asunto      string `json:"asunto" binding:"required,max=200"`
	Descripcion string `json:"descripcion" binding:"required,max=2000"`
	Categoria   string `json:"categoria" binding:"omitempty,oneof=tecnico facturacion soporte general"`
	Prioridad   string `json:"prioridad" binding:"omitempty,oneof=baja media alta critica"`
	ClienteID   string `json:"clienteId,omitempty"`
	EmpresaID   string `json:"empresaId,omitempty"`
}

type AgregarRespuestaRecipe struct {
	Mensaje string `json:"mensaje" binding:"required,max=2000"`
}

type ActualizarEstadoTicketRecipe struct {
	Estado string `json:"estado" binding:"required,oneof=abierto en_proceso resuelto cerrado"`
}
