package recipe

type EnviarLeadMagnetRecipe struct {
	Email  string `json:"email" binding:"required,email"`
	Nombre string `json:"nombre" binding:"required"`
}
