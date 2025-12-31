package recipe

type DashboardClienteRecipe struct {
	ClienteID string `json:"clienteId"`
}

type ConsumoClienteRecipe struct {
	FechaInicio string `json:"fechaInicio"`
	FechaFin    string `json:"fechaFin"`
}
