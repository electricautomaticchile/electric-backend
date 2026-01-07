package recipe

type CrearEmpresaRecipe struct {
	Nombre    string `json:"nombre" binding:"required,max=100"`
	Email     string `json:"email" binding:"required,email,max=100"`
	Telefono  string `json:"telefono,omitempty" binding:"omitempty,telefono_cl"`
	Direccion string `json:"direccion,omitempty" binding:"omitempty,max=200"`
	RUT       string `json:"rut,omitempty" binding:"omitempty,rut"`
}

type ActualizarEmpresaRecipe struct {
	Nombre    string `json:"nombre,omitempty" binding:"omitempty,max=100"`
	Email     string `json:"email,omitempty" binding:"omitempty,email,max=100"`
	Telefono  string `json:"telefono,omitempty" binding:"omitempty,telefono_cl"`
	Direccion string `json:"direccion,omitempty" binding:"omitempty,max=200"`
	RUT       string `json:"rut,omitempty" binding:"omitempty,rut"`
}

type CambiarEstadoEmpresaRecipe struct {
	Activo bool `json:"activo" binding:"required"`
}
