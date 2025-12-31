package models

import "time"

type CotizacionModel struct {
	ID                 string     `json:"_id" bson:"_id"`
	Numero             string     `json:"numero,omitempty"`
	Nombre             string     `json:"nombre" binding:"required"`
	Email              string     `json:"email" binding:"required,email"`
	Empresa            string     `json:"empresa,omitempty"`
	Telefono           string     `json:"telefono,omitempty"`
	Servicio           string     `json:"servicio" binding:"required"` // Ver valores abajo
	Plazo              string     `json:"plazo,omitempty"` // "urgente" | "pronto" | "normal" | "planificacion"
	Mensaje            string     `json:"mensaje" binding:"required"`
	Estado             string     `json:"estado"` // Ver valores abajo
	Prioridad          string     `json:"prioridad"` // "baja" | "media" | "alta" | "critica"
	FechaCreacion      time.Time  `json:"fechaCreacion"`
	FechaActualizacion *time.Time `json:"fechaActualizacion,omitempty"`
}

// Valores de Servicio:
// - "cotizacion_reposicion"
// - "cotizacion_monitoreo"
// - "cotizacion_mantenimiento"
// - "cotizacion_completa"

// Valores de Estado:
// - "pendiente"
// - "en_revision"
// - "cotizando"
// - "cotizada"
// - "aprobada"
// - "rechazada"
// - "convertida_cliente"
