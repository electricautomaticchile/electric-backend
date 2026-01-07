package recipe

type ActualizarConfiguracionRecipe struct {
	Clave string      `json:"clave" binding:"required,max=100"`
	Valor interface{} `json:"valor" binding:"required"`
}
