package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TarifaAplicada struct {
	Distribuidora string  `bson:"distribuidora" json:"distribuidora"`
	Tarifa        string  `bson:"tarifa" json:"tarifa"`
	PrecioKwh     float64 `bson:"precioKwh" json:"precioKwh"`
	CargoFijo     float64 `bson:"cargoFijo" json:"cargoFijo"`
}

type NotificacionesEnviadas struct {
	Generacion          bool `bson:"generacion" json:"generacion"`
	Recordatorio5Dias   bool `bson:"recordatorio5dias" json:"recordatorio5dias"`
	PorVencer3Dias      bool `bson:"porVencer3dias" json:"porVencer3dias"`
	Vencida             bool `bson:"vencida" json:"vencida"`
	Advertencia2Boletas bool `bson:"advertencia2boletas" json:"advertencia2boletas"`
	Aviso48h            bool `bson:"aviso48h" json:"aviso48h"`
	Aviso24h            bool `bson:"aviso24h" json:"aviso24h"`
	CorteEjecutado      bool `bson:"corteEjecutado" json:"corteEjecutado"`
}

type BoletaEntity struct {
	ID                     primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	ClienteID              primitive.ObjectID     `bson:"clienteId" json:"clienteId"`
	EmpresaID              primitive.ObjectID     `bson:"empresaId,omitempty" json:"empresaId,omitempty"`
	DispositivoID          primitive.ObjectID     `bson:"dispositivoId,omitempty" json:"dispositivoId,omitempty"`
	Monto                  float64                `bson:"monto" json:"monto"`
	MontoTotal             float64                `bson:"montoTotal" json:"montoTotal"`
	Periodo                string                 `bson:"periodo" json:"periodo"`
	Mes                    int                    `bson:"mes" json:"mes"`
	Anio                   int                    `bson:"anio" json:"anio"`
	ConsumoKwh             float64                `bson:"consumoKwh" json:"consumoKwh"`
	TarifaAplicada         *TarifaAplicada        `bson:"tarifaAplicada,omitempty" json:"tarifaAplicada,omitempty"`
	Estado                 string                 `bson:"estado" json:"estado"`
	FechaCreacion          time.Time              `bson:"fechaCreacion" json:"fechaCreacion"`
	FechaVencimiento       *time.Time             `bson:"fechaVencimiento,omitempty" json:"fechaVencimiento,omitempty"`
	FechaPago              *time.Time             `bson:"fechaPago,omitempty" json:"fechaPago,omitempty"`
	MotivoCorte            string                 `bson:"motivoCorte,omitempty" json:"motivoCorte,omitempty"`
	NotificacionesEnviadas NotificacionesEnviadas `bson:"notificacionesEnviadas" json:"notificacionesEnviadas"`
}

func (BoletaEntity) CollectionName() string {
	return "boletas"
}
