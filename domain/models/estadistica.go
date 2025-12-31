package models

type EstadisticaConsumoModel struct {
	ClienteID       string  `json:"clienteId"`
	ConsumoTotal    float64 `json:"consumoTotal"`
	CostoTotal      float64 `json:"costoTotal"`
	PromedioVoltage float64 `json:"promedioVoltage"`
	PromedioCurrent float64 `json:"promedioCurrent"`
}

type EstadisticaGlobalModel struct {
	TotalClientes      int     `json:"totalClientes"`
	TotalDispositivos  int     `json:"totalDispositivos"`
	ConsumoTotalKWh    float64 `json:"consumoTotalKwh"`
	DispositivosActivos int    `json:"dispositivosActivos"`
}
