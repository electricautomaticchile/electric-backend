package recipe

type CrearDispositivoRecipe struct {
	NumeroDispositivo string `json:"numeroDispositivo" binding:"required"`
	Nombre            string `json:"nombre" binding:"required"`
	Tipo              string `json:"tipo" binding:"required"`
	ClienteID         string `json:"clienteId" binding:"required"`
	EmpresaID         string `json:"empresaId" binding:"required"`
}

type ActualizarDispositivoRecipe struct {
	Nombre string `json:"nombre,omitempty"`
	Tipo   string `json:"tipo,omitempty"`
}

type ActualizarLecturaRecipe struct {
	Voltage     float64 `json:"voltage"`
	Current     float64 `json:"current"`
	ActivePower float64 `json:"activePower"`
	Energy      float64 `json:"energy"`
	Cost        float64 `json:"cost"`
}

type CambiarEstadoDispositivoRecipe struct {
	Estado string `json:"estado" binding:"required"`
}
