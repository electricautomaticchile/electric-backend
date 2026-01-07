package recipe

type CrearCotizacionRecipe struct {
	Nombre   string `json:"nombre" binding:"required,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Empresa  string `json:"empresa,omitempty" binding:"omitempty,max=100"`
	Telefono string `json:"telefono,omitempty" binding:"omitempty,telefono_cl"`
	Servicio string `json:"servicio" binding:"required,max=100"`
	Plazo    string `json:"plazo,omitempty" binding:"omitempty,oneof=urgente pronto normal"`
	Mensaje  string `json:"mensaje" binding:"required,max=2000"`
}

type ActualizarCotizacionRecipe struct {
	Nombre    string `json:"nombre,omitempty" binding:"omitempty,max=100"`
	Email     string `json:"email,omitempty" binding:"omitempty,email,max=100"`
	Empresa   string `json:"empresa,omitempty" binding:"omitempty,max=100"`
	Telefono  string `json:"telefono,omitempty" binding:"omitempty,telefono_cl"`
	Estado    string `json:"estado,omitempty" binding:"omitempty,oneof=pendiente contactado cotizado aceptado rechazado"`
	Prioridad string `json:"prioridad,omitempty" binding:"omitempty,oneof=baja media alta critica"`
}

type ActualizarEstadoCotizacionRecipe struct {
	Estado string `json:"estado" binding:"required,oneof=pendiente contactado cotizado aceptado rechazado"`
}
