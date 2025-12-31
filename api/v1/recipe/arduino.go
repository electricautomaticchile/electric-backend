package recipe

type RecibirDatosArduinoRecipe struct {
	DeviceID    string  `json:"deviceId" binding:"required"`
	Voltage     float64 `json:"voltage"`
	Current     float64 `json:"current"`
	ActivePower float64 `json:"activePower"`
	Energy      float64 `json:"energy"`
}

type RegistrarDispositivoArduinoRecipe struct {
	NumeroDispositivo string `json:"numeroDispositivo" binding:"required"`
	ClienteID         string `json:"clienteId" binding:"required"`
}
