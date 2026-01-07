package recipe

type CrearAlertaRecipe struct {
	EmpresaID      string `json:"empresaId" binding:"required"`
	DispositivoID  string `json:"dispositivoId,omitempty"`
	Tipo           string `json:"tipo" binding:"required,oneof=consumo_alto voltaje_anormal desconexion falla_dispositivo mantenimiento"`
	Severidad      string `json:"severidad" binding:"required,oneof=baja media alta critica"`
	Mensaje        string `json:"mensaje" binding:"required,max=500"`
}

type ResolverAlertaRecipe struct {
	Resolucion string `json:"resolucion" binding:"required,max=1000"`
}
