package recipe

type LoginRecipe struct {
	NumeroCliente string `json:"numeroCliente" binding:"required,numero_cliente"`
	Password      string `json:"password" binding:"required"`
}

type CambiarPasswordRecipe struct {
	PasswordActual string `json:"passwordActual"`
	PasswordNuevo  string `json:"passwordNuevo" binding:"required,password_strong"`
}

type SolicitarRecuperacionRecipe struct {
	Email string `json:"email" binding:"required,email,max=100"`
}

type RestablecerPasswordRecipe struct {
	Token    string `json:"token" binding:"required,max=500"`
	Password string `json:"password" binding:"required,password_strong"`
}
