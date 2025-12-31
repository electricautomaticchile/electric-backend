package models

import "time"

type ConfiguracionModel struct {
	ID                 string      `json:"id"`
	Clave              string      `json:"clave"`
	Valor              interface{} `json:"valor"`
	Categoria          string      `json:"categoria"`
	FechaActualizacion time.Time   `json:"fechaActualizacion"`
}
