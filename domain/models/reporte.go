package models

import "time"

type ReporteModel struct {
	ID              string                 `json:"_id"`
	Tipo            string                 `json:"tipo"`
	Formato         string                 `json:"formato"`
	FechaGeneracion time.Time              `json:"fechaGeneracion"`
	UsuarioID       string                 `json:"usuarioId"`
	UsuarioTipo     string                 `json:"usuarioTipo"`
	EmpresaID       string                 `json:"empresaId,omitempty"`
	Filtros         map[string]interface{} `json:"filtros"`
	Estadisticas    EstadisticasReporte    `json:"estadisticas"`
	Estado          string                 `json:"estado"`
	Metadatos       MetadatosReporte       `json:"metadatos"`
	Error           *ErrorReporte          `json:"error,omitempty"`
	ExpiresAt       time.Time              `json:"expiresAt"`
	FechaCreacion   time.Time              `json:"fechaCreacion"`
}

type EstadisticasReporte struct {
	TotalRegistros     int `json:"totalRegistros"`
	TamañoArchivo      int `json:"tamañoArchivo"`
	TiempoGeneracion   int `json:"tiempoGeneracion"`
	FiltrosAplicados   int `json:"filtrosAplicados"`
}

type MetadatosReporte struct {
	IPAddress      string `json:"ipAddress,omitempty"`
	UserAgent      string `json:"userAgent,omitempty"`
	NombreArchivo  string `json:"nombreArchivo"`
	TipoMime       string `json:"tipoMime"`
}

type ErrorReporte struct {
	Mensaje   string    `json:"mensaje"`
	Codigo    string    `json:"codigo"`
	Timestamp time.Time `json:"timestamp"`
}
