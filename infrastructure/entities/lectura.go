package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LecturaEntity struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Timestamp     time.Time          `bson:"timestamp"`
	DispositivoID string             `bson:"dispositivoId"`
	ClienteID     string             `bson:"clienteId,omitempty"`
	EmpresaID     string             `bson:"empresaId,omitempty"`
	Voltaje       float64            `bson:"voltaje"`
	Corriente     float64            `bson:"corriente"`
	Potencia      float64            `bson:"potencia"`
	Energia       float64            `bson:"energia"`
	Costo         float64            `bson:"costo"`
}

func (LecturaEntity) CollectionName() string {
	return "lecturas"
}

type LecturaAgregadaEntity struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Timestamp     time.Time          `bson:"timestamp"`
	DispositivoID string             `bson:"dispositivoId"`
	ClienteID     string             `bson:"clienteId,omitempty"`
	EmpresaID     string             `bson:"empresaId,omitempty"`
	Periodo       string             `bson:"periodo"`
	VoltajeMin    float64            `bson:"voltajeMin"`
	VoltajeMax    float64            `bson:"voltajeMax"`
	VoltajeAvg    float64            `bson:"voltajeAvg"`
	CorrienteMin  float64            `bson:"corrienteMin"`
	CorrienteMax  float64            `bson:"corrienteMax"`
	CorrienteAvg  float64            `bson:"corrienteAvg"`
	PotenciaMin   float64            `bson:"potenciaMin"`
	PotenciaMax   float64            `bson:"potenciaMax"`
	PotenciaAvg   float64            `bson:"potenciaAvg"`
	EnergiaInicio float64            `bson:"energiaInicio"`
	EnergiaFin    float64            `bson:"energiaFin"`
	CostoInicio   float64            `bson:"costoInicio"`
	CostoFin      float64            `bson:"costoFin"`
	NumLecturas   int                `bson:"numLecturas"`
}

func (LecturaAgregadaEntity) CollectionName() string {
	return "lecturas_agregadas"
}
