package recipe

type ContactoPrincipalRecipe struct {
	Nombre   string `json:"nombre,omitempty"`
	Cargo    string `json:"cargo,omitempty"`
	Telefono string `json:"telefono,omitempty"`
	Correo   string `json:"correo,omitempty"`
}

type ConfiguracionesRecipe struct {
	Notificaciones bool   `json:"notificaciones,omitempty"`
	Tema           string `json:"tema,omitempty"`
	MaxUsuarios    int    `json:"maxUsuarios,omitempty"`
}

type CrearEmpresaRecipe struct {
	Nombre    string `json:"nombre" binding:"required,max=100"`
	Email     string `json:"email" binding:"required,email,max=100"`
	Telefono  string `json:"telefono,omitempty" binding:"omitempty,telefono_cl"`
	Direccion string `json:"direccion,omitempty" binding:"omitempty,max=200"`
	RUT       string `json:"rut,omitempty" binding:"omitempty,rut"`
}

type ActualizarEmpresaRecipe struct {
	NombreEmpresa     string                    `json:"nombreEmpresa,omitempty"`
	RazonSocial       string                    `json:"razonSocial,omitempty"`
	Rut               string                    `json:"rut,omitempty"`
	Correo            string                    `json:"correo,omitempty"`
	Telefono          string                    `json:"telefono,omitempty"`
	Direccion         string                    `json:"direccion,omitempty"`
	Ciudad            string                    `json:"ciudad,omitempty"`
	Region            string                    `json:"region,omitempty"`
	ContactoPrincipal *ContactoPrincipalRecipe  `json:"contactoPrincipal,omitempty"`
	Configuraciones   *ConfiguracionesRecipe    `json:"configuraciones,omitempty"`
}

type CambiarEstadoEmpresaRecipe struct {
	Activo bool `json:"activo" binding:"required"`
}
