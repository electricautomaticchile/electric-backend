package models

import "time"

type ClienteModel struct {
	ID                string     `json:"_id" bson:"_id"`
	Nombre            string     `json:"nombre" binding:"required"`
	Correo            string     `json:"correo" binding:"required,email"`
	Telefono          string     `json:"telefono,omitempty"`
	Direccion         string     `json:"direccion,omitempty"`
	Ciudad            string     `json:"ciudad,omitempty"`
	Latitud           float64    `json:"latitud,omitempty"`
	Longitud          float64    `json:"longitud,omitempty"`
	Rut               string     `json:"rut,omitempty"`
	TipoCliente       string     `json:"tipoCliente,omitempty"`
	Empresa           string     `json:"empresa,omitempty"`
	NumeroCliente     string     `json:"numeroCliente,omitempty"`
	Activo            bool       `json:"activo"`
	Password          string     `json:"-"`
	PasswordTemporal  string     `json:"passwordTemporal,omitempty"`
	Role              string     `json:"role,omitempty"`
	TipoUsuario       string     `json:"tipoUsuario,omitempty"`
	FechaRegistro     time.Time  `json:"fechaRegistro,omitempty"`
	FechaActivacion   *time.Time `json:"fechaActivacion,omitempty"`
	UltimoAcceso      *time.Time `json:"ultimoAcceso,omitempty"`
	EmpresaID         string     `json:"empresaId,omitempty"`
	UsuarioID         string     `json:"usuarioId,omitempty"`
}
