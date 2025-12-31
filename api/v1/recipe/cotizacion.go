package recipe

type CrearCotizacionRecipe struct {
	Nombre   string `json:"nombre" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Empresa  string `json:"empresa,omitempty"`
	Telefono string `json:"telefono,omitempty"`
	Servicio string `json:"servicio" binding:"required"`
	Plazo    string `json:"plazo,omitempty"`
	Mensaje  string `json:"mensaje" binding:"required"`
}

type ActualizarCotizacionRecipe struct {
	Nombre    string `json:"nombre,omitempty"`
	Email     string `json:"email,omitempty"`
	Empresa   string `json:"empresa,omitempty"`
	Telefono  string `json:"telefono,omitempty"`
	Estado    string `json:"estado,omitempty"`
	Prioridad string `json:"prioridad,omitempty"`
}

type ActualizarEstadoCotizacionRecipe struct {
	Estado string `json:"estado" binding:"required"`
}
