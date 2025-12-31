package recipe

type RegistroEmpresaRecipe struct {
	NombreEmpresa string `json:"nombreEmpresa" binding:"required"`
	RazonSocial   string `json:"razonSocial" binding:"required"`
	Rut           string `json:"rut" binding:"required"`
	Correo        string `json:"correo" binding:"required,email"`
	Telefono      string `json:"telefono" binding:"required"`
	Direccion     string `json:"direccion" binding:"required"`
	Ciudad        string `json:"ciudad" binding:"required"`
	Region        string `json:"region" binding:"required"`
	ContactoPrincipal ContactoPrincipalRecipe `json:"contactoPrincipal" binding:"required"`
}

type ContactoPrincipalRecipe struct {
	Nombre   string `json:"nombre" binding:"required"`
	Cargo    string `json:"cargo" binding:"required"`
	Telefono string `json:"telefono" binding:"required"`
	Correo   string `json:"correo" binding:"required,email"`
}
