package recipe

type CrearEmpresaRecipe struct {
	Nombre    string `json:"nombre" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Telefono  string `json:"telefono,omitempty"`
	Direccion string `json:"direccion,omitempty"`
	RUT       string `json:"rut,omitempty"`
}

type ActualizarEmpresaRecipe struct {
	Nombre    string `json:"nombre,omitempty"`
	Email     string `json:"email,omitempty"`
	Telefono  string `json:"telefono,omitempty"`
	Direccion string `json:"direccion,omitempty"`
	RUT       string `json:"rut,omitempty"`
}

type CambiarEstadoEmpresaRecipe struct {
	Activo bool `json:"activo" binding:"required"`
}
