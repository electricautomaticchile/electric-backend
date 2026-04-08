package dto

import "time"

// DtoCliente es la respuesta pública de un cliente (sin password ni datos sensibles).
type DtoCliente struct {
	ID              string     `json:"id"`
	Nombre          string     `json:"nombre"`
	Correo          string     `json:"correo"`
	Telefono        string     `json:"telefono,omitempty"`
	NumeroCliente   string     `json:"numeroCliente"`
	Rut             string     `json:"rut"`
	Direccion       string     `json:"direccion,omitempty"`
	Ciudad          string     `json:"ciudad,omitempty"`
	Estado          string     `json:"estado"`
	Activo          bool       `json:"activo"`
	FechaCreacion   *time.Time `json:"fechaCreacion,omitempty"`
	UltimoAcceso    *time.Time `json:"ultimoAcceso,omitempty"`
}

// DtoLoginResponse es la respuesta del login (sin exponer el hash del password).
type DtoLoginResponse struct {
	Token                  string      `json:"token"`
	RefreshToken           string      `json:"refreshToken"`
	User                   *DtoCliente `json:"user"`
	RequiereCambioPassword bool        `json:"requiereCambioPassword"`
}
