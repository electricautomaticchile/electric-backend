package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UsuarioEmpresaEntity struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	EmpresaID         primitive.ObjectID `bson:"empresaId" json:"empresaId"`
	Nombre            string             `bson:"nombre" json:"nombre"`
	Email             string             `bson:"email" json:"email"`
	Password          string             `bson:"password" json:"password"`
	Role              string             `bson:"role" json:"role"`
	Telefono          string             `bson:"telefono,omitempty" json:"telefono,omitempty"`
	Cargo             string             `bson:"cargo,omitempty" json:"cargo,omitempty"`
	Activo            bool               `bson:"activo" json:"activo"`
	PasswordTemporal  bool               `bson:"passwordTemporal" json:"passwordTemporal"`
	UltimoAcceso      time.Time          `bson:"ultimoAcceso,omitempty" json:"ultimoAcceso,omitempty"`
	FechaCreacion     time.Time          `bson:"fechaCreacion" json:"fechaCreacion"`
	FechaActualizacion time.Time         `bson:"fechaActualizacion" json:"fechaActualizacion"`
}
