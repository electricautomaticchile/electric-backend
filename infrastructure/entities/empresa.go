package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContactoPrincipalEntity struct {
	Nombre       string `bson:"nombre" json:"nombre"`
	Cargo        string `bson:"cargo" json:"cargo"`
	Telefono     string `bson:"telefono" json:"telefono"`
	Correo       string `bson:"correo" json:"correo"`
	ImagenPerfil string `bson:"imagenPerfil,omitempty" json:"imagenPerfil,omitempty"`
}

type EmpresaEntity struct {
	ID                 primitive.ObjectID      `bson:"_id,omitempty" json:"id"`
	NombreEmpresa      string                  `bson:"nombreEmpresa" json:"nombreEmpresa"`
	RazonSocial        string                  `bson:"razonSocial,omitempty" json:"razonSocial,omitempty"`
	Correo             string                  `bson:"correo" json:"correo"`
	NumeroCliente      string                  `bson:"numeroCliente" json:"numeroCliente"`
	Password           string                  `bson:"password" json:"-"`
	PasswordTemporal   bool                    `bson:"passwordTemporal,omitempty" json:"passwordTemporal,omitempty"`
	Telefono           string                  `bson:"telefono,omitempty" json:"telefono,omitempty"`
	Direccion          string                  `bson:"direccion,omitempty" json:"direccion,omitempty"`
	Ciudad             string                  `bson:"ciudad,omitempty" json:"ciudad,omitempty"`
	Region             string                  `bson:"region,omitempty" json:"region,omitempty"`
	RUT                string                  `bson:"rut,omitempty" json:"rut,omitempty"`
	ContactoPrincipal  ContactoPrincipalEntity `bson:"contactoPrincipal,omitempty" json:"contactoPrincipal,omitempty"`
	Role               string                  `bson:"role" json:"role"`
	TipoUsuario        string                  `bson:"tipoUsuario" json:"tipoUsuario"`
	Estado             string                  `bson:"estado" json:"estado"`
	UsuarioID          primitive.ObjectID      `bson:"usuarioId,omitempty" json:"usuarioId,omitempty"`
	Activo             bool                    `bson:"activo" json:"activo"`
	FechaCreacion      interface{}             `bson:"fechaCreacion,omitempty" json:"fechaCreacion,omitempty"`
	FechaActualizacion interface{}             `bson:"fechaActualizacion,omitempty" json:"fechaActualizacion,omitempty"`
	UltimoAcceso       interface{}             `bson:"ultimoAcceso,omitempty" json:"ultimoAcceso,omitempty"`
	FechaActivacion    interface{}             `bson:"fechaActivacion,omitempty" json:"fechaActivacion,omitempty"`
	FechaSuspension    interface{}             `bson:"fechaSuspension,omitempty" json:"fechaSuspension,omitempty"`
	MotivoSuspension   string                  `bson:"motivoSuspension,omitempty" json:"motivoSuspension,omitempty"`
}

func (EmpresaEntity) CollectionName() string {
	return "empresas"
}
