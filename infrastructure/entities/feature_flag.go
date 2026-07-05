package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FeatureFlagEntity representa un flag de funcionalidad para hacer rollout
// gradual durante el piloto, sin necesidad de redeploy (se cambia en la BD).
//
// Semántica:
//   - Enabled=false  → apagado para todos.
//   - Enabled=true y EmpresaIDs vacío → encendido para todos.
//   - Enabled=true y EmpresaIDs con valores → encendido solo para esas empresas.
type FeatureFlagEntity struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Key         string             `bson:"key" json:"key"`
	Descripcion string             `bson:"descripcion,omitempty" json:"descripcion,omitempty"`
	Enabled     bool               `bson:"enabled" json:"enabled"`
	EmpresaIDs  []string           `bson:"empresaIds,omitempty" json:"empresaIds,omitempty"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

func (FeatureFlagEntity) CollectionName() string {
	return "feature_flags"
}
