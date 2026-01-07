package recipe

type CrearClienteRecipe struct {
	Nombre        string `json:"nombre" binding:"required,max=100"`
	Correo        string `json:"correo" binding:"required,email,max=100"`
	NumeroCliente string `json:"numeroCliente" binding:"omitempty,numero_cliente"`
	Telefono      string `json:"telefono,omitempty" binding:"omitempty,telefono_cl"`
	Direccion     string `json:"direccion,omitempty" binding:"omitempty,max=200"`
	Ciudad        string `json:"ciudad,omitempty" binding:"omitempty,max=50"`
	Rut           string `json:"rut,omitempty" binding:"omitempty,rut"`
	TipoCliente   string `json:"tipoCliente,omitempty" binding:"omitempty,max=50"`
	Empresa       string `json:"empresa,omitempty" binding:"omitempty,max=100"`
	EmpresaID     string `json:"empresaId" binding:"required"`
}

type ActualizarClienteRecipe struct {
	Nombre    string `json:"nombre,omitempty" binding:"omitempty,max=100"`
	Correo    string `json:"correo,omitempty" binding:"omitempty,email,max=100"`
	Telefono  string `json:"telefono,omitempty" binding:"omitempty,telefono_cl"`
	Direccion string `json:"direccion,omitempty" binding:"omitempty,max=200"`
	Ciudad    string `json:"ciudad,omitempty" binding:"omitempty,max=50"`
	Rut       string `json:"rut,omitempty" binding:"omitempty,rut"`
	Activo    *bool  `json:"activo,omitempty"`
}
