package recipe

type CrearAlertaRecipe struct {
	EmpresaID      string `json:"empresaId" binding:"required"`
	DispositivoID  string `json:"dispositivoId,omitempty"`
	Tipo           string `json:"tipo" binding:"required"`
	Severidad      string `json:"severidad" binding:"required"`
	Mensaje        string `json:"mensaje" binding:"required"`
}

type ResolverAlertaRecipe struct {
	Resolucion string `json:"resolucion" binding:"required"`
}
