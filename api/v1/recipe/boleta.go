package recipe

type CrearBoletaRecipe struct {
	ClienteID string  `json:"clienteId" binding:"required"`
	Monto     float64 `json:"monto" binding:"required,gt=0"`
	Periodo   string  `json:"periodo" binding:"required,periodo"`
}
