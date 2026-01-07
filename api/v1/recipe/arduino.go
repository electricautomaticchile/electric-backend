package recipe

type RecibirDatosArduinoRecipe struct {
	DeviceID    string  `json:"deviceId" binding:"required,max=50"`
	Voltage     float64 `json:"voltage" binding:"min=0,max=500"`
	Current     float64 `json:"current" binding:"min=0,max=200"`
	ActivePower float64 `json:"activePower" binding:"min=0"`
	Energy      float64 `json:"energy" binding:"min=0"`
}

type RegistrarDispositivoArduinoRecipe struct {
	NumeroDispositivo string `json:"numeroDispositivo" binding:"required,max=50"`
	ClienteID         string `json:"clienteId" binding:"required"`
}
