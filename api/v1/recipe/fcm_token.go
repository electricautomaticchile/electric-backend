package recipe

type RegistrarFCMTokenRecipe struct {
	Token      string `json:"token" binding:"required"`
	Plataforma string `json:"plataforma" binding:"required,oneof=android ios"`
}
