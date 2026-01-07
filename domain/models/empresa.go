package models

import "time"

type ContactoPrincipal struct {
	Nombre   string `json:"nombre" binding:"required"`
	Cargo    string `json:"cargo" binding:"required"`
	Telefono string `json:"telefono" binding:"required"`
	Correo   string `json:"correo" binding:"required,email"`
}

type EmpresaModel struct {
	ID                 string             `json:"_id" bson:"_id"`
	NombreEmpresa      string             `json:"nombreEmpresa" binding:"required"`
	RazonSocial        string             `json:"razonSocial" binding:"required"`
	Rut                string             `json:"rut" binding:"required"`
	Correo             string             `json:"correo" binding:"required,email"`
	Telefono           string             `json:"telefono" binding:"required"`
	Direccion          string             `json:"direccion" binding:"required"`
	Ciudad             string             `json:"ciudad" binding:"required"`
	Region             string             `json:"region" binding:"required"`
	ImagenPerfil       string             `json:"imagenPerfil,omitempty"`
	ContactoPrincipal  ContactoPrincipal  `json:"contactoPrincipal" binding:"required"`
	NumeroCliente      string             `json:"numeroCliente"`
	Password           string             `json:"-"` // No exponer en JSON
	PasswordTemporal   bool               `json:"passwordTemporal"`
	Role               string             `json:"role,omitempty"`
	TipoUsuario        string             `json:"tipoUsuario,omitempty"`
	Estado             string             `json:"estado"` // "activo" | "suspendido" | "inactivo"
	FechaCreacion      time.Time          `json:"fechaCreacion"`
	FechaActualizacion *time.Time         `json:"fechaActualizacion,omitempty"`
	UltimoAcceso       *time.Time         `json:"ultimoAcceso,omitempty"`
	FechaActivacion    *time.Time         `json:"fechaActivacion,omitempty"`
	FechaSuspension    *time.Time         `json:"fechaSuspension,omitempty"`
	MotivoSuspension   string             `json:"motivoSuspension,omitempty"`
	UsuarioID          string             `json:"usuarioId,omitempty"`
}
