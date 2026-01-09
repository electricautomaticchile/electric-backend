package recipe

type CrearUsuarioEmpresaRecipe struct {
	Nombre   string `json:"nombre" binding:"required,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Role     string `json:"role" binding:"required,oneof=EMPRESA_ADMIN EMPRESA_OPERADOR EMPRESA_SOPORTE EMPRESA_FINANCIERO"`
	Telefono string `json:"telefono,omitempty" binding:"omitempty,telefono_cl"`
	Cargo    string `json:"cargo,omitempty" binding:"omitempty,max=50"`
}

type ActualizarUsuarioEmpresaRecipe struct {
	Nombre   string `json:"nombre,omitempty" binding:"omitempty,max=100"`
	Telefono string `json:"telefono,omitempty" binding:"omitempty,telefono_cl"`
	Cargo    string `json:"cargo,omitempty" binding:"omitempty,max=50"`
	Role     string `json:"role,omitempty" binding:"omitempty,oneof=EMPRESA_ADMIN EMPRESA_OPERADOR EMPRESA_SOPORTE EMPRESA_FINANCIERO"`
	Activo   *bool  `json:"activo,omitempty"`
}

type LoginEmpresaRecipe struct {
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required"`
}
