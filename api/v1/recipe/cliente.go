package recipe

type CrearClienteRecipe struct {
	Nombre        string `json:"nombre" binding:"required"`
	Correo        string `json:"correo" binding:"required,email"`
	NumeroCliente string `json:"numeroCliente"`
	Telefono      string `json:"telefono,omitempty"`
	Direccion     string `json:"direccion,omitempty"`
	Ciudad        string `json:"ciudad,omitempty"`
	Rut           string `json:"rut,omitempty"`
	TipoCliente   string `json:"tipoCliente,omitempty"`
	Empresa       string `json:"empresa,omitempty"`
	EmpresaID     string `json:"empresaId" binding:"required"`
}

type ActualizarClienteRecipe struct {
	Nombre    string `json:"nombre,omitempty"`
	Correo    string `json:"correo,omitempty"`
	Telefono  string `json:"telefono,omitempty"`
	Direccion string `json:"direccion,omitempty"`
	Ciudad    string `json:"ciudad,omitempty"`
	Rut       string `json:"rut,omitempty"`
	Activo    *bool  `json:"activo,omitempty"`
}
