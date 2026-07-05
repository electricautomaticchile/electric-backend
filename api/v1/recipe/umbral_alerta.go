package recipe

// SetUmbralAlertaRecipe define los umbrales de alerta configurables por empresa.
// Los valores deben ser coherentes (voltajeMin < voltajeMax); la validación
// semántica se realiza en el controller.
type SetUmbralAlertaRecipe struct {
	VoltajeMin   float64 `json:"voltajeMin" binding:"required,gt=0"`
	VoltajeMax   float64 `json:"voltajeMax" binding:"required,gt=0"`
	CorrienteMax float64 `json:"corrienteMax" binding:"required,gt=0"`
	ConsumoMax   float64 `json:"consumoMax" binding:"required,gt=0"`
}
