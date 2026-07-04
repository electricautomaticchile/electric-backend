package recipe

type CrearDispositivoRecipe struct {
	NumeroDispositivo string `json:"numeroDispositivo" binding:"required,max=50"`
	Nombre            string `json:"nombre" binding:"required,max=100"`
	Tipo              string `json:"tipo" binding:"required,oneof=medidor sensor controlador gateway"`
	ClienteID         string `json:"clienteId" binding:"required"`
	EmpresaID         string `json:"empresaId,omitempty"`
}

type ActualizarDispositivoRecipe struct {
	Nombre string `json:"nombre,omitempty" binding:"omitempty,max=100"`
	Tipo   string `json:"tipo,omitempty" binding:"omitempty,oneof=medidor sensor controlador gateway"`
}

type AsignarDispositivoRecipe struct {
	ClienteID string `json:"clienteId" binding:"required"`
}

type ActualizarLecturaRecipe struct {
	Voltage        float64 `json:"voltage" binding:"min=0,max=500"`
	Current        float64 `json:"current" binding:"min=0,max=200"`
	ActivePower    float64 `json:"activePower" binding:"min=0"`
	Energy         float64 `json:"energy" binding:"min=0"`
	Cost           float64 `json:"cost" binding:"min=0"`
	Frecuencia     float64 `json:"frecuencia" binding:"min=0,max=100"`
	FactorPotencia float64 `json:"factorPotencia" binding:"min=0,max=1"`
	// Ubicación GPS opcional reportada por el ESP32 en la lectura.
	Latitud  *float64 `json:"latitud,omitempty" binding:"omitempty,min=-90,max=90"`
	Longitud *float64 `json:"longitud,omitempty" binding:"omitempty,min=-180,max=180"`
}

type CambiarEstadoDispositivoRecipe struct {
	Estado string `json:"estado" binding:"required,oneof=activo inactivo mantenimiento"`
}
