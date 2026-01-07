package recipe

type EnviarLeadMagnetRecipe struct {
	Email  string `json:"email" binding:"required,email,max=100"`
	Nombre string `json:"nombre" binding:"required,max=100"`
}
