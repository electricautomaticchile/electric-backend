package recipe

type CrearLeadRecipe struct {
	Type           string                 `json:"type" binding:"required,oneof=investor distributor"`
	Name           string                 `json:"name" binding:"required,min=2,max=120"`
	Email          string                 `json:"email" binding:"required,email,max=255"`
	Organization   string                 `json:"organization,omitempty" binding:"omitempty,max=200"`
	Message        string                 `json:"message,omitempty" binding:"omitempty,max=2000"`
	Extra          map[string]interface{} `json:"extra,omitempty"`
	TurnstileToken string                 `json:"turnstileToken,omitempty" binding:"omitempty,max=2048"`
}

type ActualizarEstadoLeadRecipe struct {
	Status string `json:"status" binding:"required,oneof=new contacted qualified discarded"`
}
