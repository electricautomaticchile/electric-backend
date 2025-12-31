package recipe

type LoginRecipe struct {
	NumeroCliente string `json:"numeroCliente" binding:"required"`
	Password      string `json:"password" binding:"required"`
}

type CambiarPasswordRecipe struct {
	PasswordActual string `json:"passwordActual" binding:"required"`
	PasswordNuevo  string `json:"passwordNuevo" binding:"required,min=6"`
}

type SolicitarRecuperacionRecipe struct {
	Email string `json:"email" binding:"required,email"`
}

type RestablecerPasswordRecipe struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}
