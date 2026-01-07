package recipe

type RegistroEmpresaRecipe struct {
	NombreEmpresa string `json:"nombreEmpresa" binding:"required,max=100"`
	RazonSocial   string `json:"razonSocial" binding:"required,max=150"`
	Rut           string `json:"rut" binding:"required,rut"`
	Correo        string `json:"correo" binding:"required,email,max=100"`
	Telefono      string `json:"telefono" binding:"required,telefono_cl"`
	Direccion     string `json:"direccion" binding:"required,max=200"`
	Ciudad        string `json:"ciudad" binding:"required,max=50"`
	Region        string `json:"region" binding:"required,max=50"`
	ContactoPrincipal ContactoPrincipalRecipe `json:"contactoPrincipal" binding:"required"`
}

type ContactoPrincipalRecipe struct {
	Nombre   string `json:"nombre" binding:"required,max=100"`
	Cargo    string `json:"cargo" binding:"required,max=50"`
	Telefono string `json:"telefono" binding:"required,telefono_cl"`
	Correo   string `json:"correo" binding:"required,email,max=100"`
}
