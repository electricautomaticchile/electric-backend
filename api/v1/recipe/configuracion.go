package recipe

type ActualizarConfiguracionRecipe struct {
	Clave string      `json:"clave" binding:"required"`
	Valor interface{} `json:"valor" binding:"required"`
}
