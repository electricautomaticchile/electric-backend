package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ClienteEntity struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Nombre             string             `bson:"nombre" json:"nombre"`
	Correo             string             `bson:"correo" json:"correo"`
	NumeroCliente      string             `bson:"numeroCliente" json:"numeroCliente"`
	Password           string             `bson:"password,omitempty" json:"-"`
	PasswordTemporal   bool               `bson:"passwordTemporal,omitempty" json:"passwordTemporal,omitempty"`
	Telefono           string             `bson:"telefono,omitempty" json:"telefono,omitempty"`
	Direccion          string             `bson:"direccion,omitempty" json:"direccion,omitempty"`
	Ciudad             string             `bson:"ciudad,omitempty" json:"ciudad,omitempty"`
	Latitud            float64            `bson:"latitud,omitempty" json:"latitud,omitempty"`
	Longitud           float64            `bson:"longitud,omitempty" json:"longitud,omitempty"`
	Rut                string             `bson:"rut,omitempty" json:"rut,omitempty"`
	ImagenPerfil       string             `bson:"imagenPerfil,omitempty" json:"imagenPerfil,omitempty"`
	TipoCliente        string             `bson:"tipoCliente,omitempty" json:"tipoCliente,omitempty"`
	Empresa            string             `bson:"empresa,omitempty" json:"empresa,omitempty"`
	Role               string             `bson:"role,omitempty" json:"role,omitempty"`
	TipoUsuario        string             `bson:"tipoUsuario,omitempty" json:"tipoUsuario,omitempty"`
	EmpresaID          primitive.ObjectID `bson:"empresaId,omitempty" json:"empresaId,omitempty"`
	UsuarioID          primitive.ObjectID `bson:"usuarioId,omitempty" json:"usuarioId,omitempty"`
	Activo             bool               `bson:"activo" json:"activo"`
	NotificacionesSMS  bool               `bson:"notificacionesSms,omitempty" json:"notificacionesSms,omitempty"`
	EsTerceraEdad      bool               `bson:"esTerceraEdad,omitempty" json:"esTerceraEdad,omitempty"`
	FechaCreacion      interface{}        `bson:"fechaCreacion,omitempty" json:"fechaCreacion,omitempty"`
	FechaActualizacion interface{}        `bson:"fechaActualizacion,omitempty" json:"fechaActualizacion,omitempty"`
	UltimoAcceso       interface{}        `bson:"ultimoAcceso,omitempty" json:"ultimoAcceso,omitempty"`
}

func (ClienteEntity) CollectionName() string {
	return "clientes"
}
