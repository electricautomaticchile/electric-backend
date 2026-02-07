package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DispositivoEntity struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	NumeroDispositivo   string             `bson:"numeroDispositivo" json:"numeroDispositivo"`
	Nombre              string             `bson:"nombre" json:"nombre"`
	Tipo                string             `bson:"tipo" json:"tipo"`
	ClienteID           primitive.ObjectID `bson:"clienteId" json:"clienteId"`
	EmpresaID           primitive.ObjectID `bson:"empresaId" json:"empresaId"`
	Estado              string             `bson:"estado" json:"estado"`
	Latitud             float64            `bson:"latitud,omitempty" json:"latitud,omitempty"`
	Longitud            float64            `bson:"longitud,omitempty" json:"longitud,omitempty"`
	Direccion           string             `bson:"direccion,omitempty" json:"direccion,omitempty"`
	UltimaLectura       *LecturaDispositivo `bson:"ultimaLectura,omitempty" json:"ultimaLectura,omitempty"`
	Configuracion       map[string]interface{} `bson:"configuracion,omitempty" json:"configuracion,omitempty"`
	Activo              bool               `bson:"activo" json:"activo"`
	FechaCreacion       time.Time          `bson:"fechaCreacion" json:"fechaCreacion"`
	FechaActualizacion  *time.Time         `bson:"fechaActualizacion,omitempty" json:"fechaActualizacion,omitempty"`
}

type LecturaDispositivo struct {
	Voltage       float64   `bson:"voltage" json:"voltage"`
	Current       float64   `bson:"current" json:"current"`
	ActivePower   float64   `bson:"activePower" json:"activePower"`
	Energy        float64   `bson:"energy" json:"energy"`
	Cost          float64   `bson:"cost" json:"cost"`
	ConsumoKWh    float64   `bson:"consumoKwh" json:"consumoKwh"`
	CostoEstimado float64   `bson:"costoEstimado" json:"costoEstimado"`
	Timestamp     time.Time `bson:"timestamp" json:"timestamp"`
}

func (DispositivoEntity) CollectionName() string {
	return "dispositivos"
}
