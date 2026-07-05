package models

import "time"

type FeatureFlagModel struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Descripcion string    `json:"descripcion,omitempty"`
	Enabled     bool      `json:"enabled"`
	EmpresaIDs  []string  `json:"empresaIds"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
