package recipe

type SetFeatureFlagRecipe struct {
	Key         string   `json:"key" binding:"required,max=100"`
	Descripcion string   `json:"descripcion,omitempty" binding:"max=300"`
	Enabled     bool     `json:"enabled"`
	EmpresaIDs  []string `json:"empresaIds,omitempty"`
}
