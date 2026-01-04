package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReporteEntity struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Tipo            string                 `bson:"tipo" json:"tipo"`
	Formato         string                 `bson:"formato" json:"formato"`
	FechaGeneracion time.Time              `bson:"fechaGeneracion" json:"fechaGeneracion"`
	UsuarioID       primitive.ObjectID     `bson:"usuarioId" json:"usuarioId"`
	UsuarioTipo     string                 `bson:"usuarioTipo" json:"usuarioTipo"`
	EmpresaID       primitive.ObjectID     `bson:"empresaId,omitempty" json:"empresaId,omitempty"`
	Filtros         map[string]interface{} `bson:"filtros" json:"filtros"`
	Estadisticas    EstadisticasReporteEntity `bson:"estadisticas" json:"estadisticas"`
	Estado          string                 `bson:"estado" json:"estado"`
	Metadatos       MetadatosReporteEntity `bson:"metadatos" json:"metadatos"`
	Error           *ErrorReporteEntity    `bson:"error,omitempty" json:"error,omitempty"`
	ExpiresAt       time.Time              `bson:"expiresAt" json:"expiresAt"`
	FechaCreacion   time.Time              `bson:"fechaCreacion" json:"fechaCreacion"`
}

type EstadisticasReporteEntity struct {
	TotalRegistros   int `bson:"totalRegistros" json:"totalRegistros"`
	TamañoArchivo    int `bson:"tamañoArchivo" json:"tamañoArchivo"`
	TiempoGeneracion int `bson:"tiempoGeneracion" json:"tiempoGeneracion"`
	FiltrosAplicados int `bson:"filtrosAplicados" json:"filtrosAplicados"`
}

type MetadatosReporteEntity struct {
	IPAddress     string `bson:"ipAddress,omitempty" json:"ipAddress,omitempty"`
	UserAgent     string `bson:"userAgent,omitempty" json:"userAgent,omitempty"`
	NombreArchivo string `bson:"nombreArchivo" json:"nombreArchivo"`
	TipoMime      string `bson:"tipoMime" json:"tipoMime"`
}

type ErrorReporteEntity struct {
	Mensaje   string    `bson:"mensaje" json:"mensaje"`
	Codigo    string    `bson:"codigo" json:"codigo"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

func (ReporteEntity) CollectionName() string {
	return "reportes"
}
