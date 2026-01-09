package models

import "time"

type UsuarioEmpresaModel struct {
	ID                 string    `json:"id"`
	EmpresaID          string    `json:"empresaId"`
	Nombre             string    `json:"nombre"`
	Email              string    `json:"email"`
	Password           string    `json:"password,omitempty"`
	Role               string    `json:"role"`
	Telefono           string    `json:"telefono,omitempty"`
	Cargo              string    `json:"cargo,omitempty"`
	Activo             bool      `json:"activo"`
	PasswordTemporal   bool      `json:"passwordTemporal"`
	UltimoAcceso       time.Time `json:"ultimoAcceso,omitempty"`
	FechaCreacion      time.Time `json:"fechaCreacion"`
	FechaActualizacion time.Time `json:"fechaActualizacion"`
}

const (
	RoleEmpresaAdmin      = "EMPRESA_ADMIN"
	RoleEmpresaOperador   = "EMPRESA_OPERADOR"
	RoleEmpresaSoporte    = "EMPRESA_SOPORTE"
	RoleEmpresaFinanciero = "EMPRESA_FINANCIERO"
)

var RolesEmpresaValidos = []string{
	RoleEmpresaAdmin,
	RoleEmpresaOperador,
	RoleEmpresaSoporte,
	RoleEmpresaFinanciero,
}

type PermisosRole struct {
	Clientes     PermisosModulo `json:"clientes"`
	Dispositivos PermisosModulo `json:"dispositivos"`
	Alertas      PermisosModulo `json:"alertas"`
	Boletas      PermisosModulo `json:"boletas"`
	Tickets      PermisosModulo `json:"tickets"`
	Reportes     PermisosModulo `json:"reportes"`
	Configuracion PermisosModulo `json:"configuracion"`
	Usuarios     PermisosModulo `json:"usuarios"`
}

type PermisosModulo struct {
	Ver      bool `json:"ver"`
	Crear    bool `json:"crear"`
	Editar   bool `json:"editar"`
	Eliminar bool `json:"eliminar"`
	Exportar bool `json:"exportar"`
}

func GetPermisosRole(role string) PermisosRole {
	switch role {
	case RoleEmpresaAdmin:
		return PermisosRole{
			Clientes:     PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: true, Exportar: true},
			Dispositivos: PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: true, Exportar: true},
			Alertas:      PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: true, Exportar: true},
			Boletas:      PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: true, Exportar: true},
			Tickets:      PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: true, Exportar: true},
			Reportes:     PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: true, Exportar: true},
			Configuracion: PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: true, Exportar: true},
			Usuarios:     PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: true, Exportar: true},
		}
	case RoleEmpresaOperador:
		return PermisosRole{
			Clientes:     PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: false, Exportar: false},
			Dispositivos: PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: false, Exportar: false},
			Alertas:      PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: false, Exportar: false},
			Boletas:      PermisosModulo{Ver: true, Crear: false, Editar: false, Eliminar: false, Exportar: false},
			Tickets:      PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: false, Exportar: false},
			Reportes:     PermisosModulo{Ver: true, Crear: false, Editar: false, Eliminar: false, Exportar: false},
			Configuracion: PermisosModulo{Ver: false, Crear: false, Editar: false, Eliminar: false, Exportar: false},
			Usuarios:     PermisosModulo{Ver: false, Crear: false, Editar: false, Eliminar: false, Exportar: false},
		}
	case RoleEmpresaSoporte:
		return PermisosRole{
			Clientes:     PermisosModulo{Ver: true, Crear: false, Editar: false, Eliminar: false, Exportar: false},
			Dispositivos: PermisosModulo{Ver: true, Crear: false, Editar: false, Eliminar: false, Exportar: false},
			Alertas:      PermisosModulo{Ver: true, Crear: false, Editar: false, Eliminar: false, Exportar: false},
			Boletas:      PermisosModulo{Ver: true, Crear: false, Editar: false, Eliminar: false, Exportar: false},
			Tickets:      PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: false, Exportar: false},
			Reportes:     PermisosModulo{Ver: false, Crear: false, Editar: false, Eliminar: false, Exportar: false},
			Configuracion: PermisosModulo{Ver: false, Crear: false, Editar: false, Eliminar: false, Exportar: false},
			Usuarios:     PermisosModulo{Ver: false, Crear: false, Editar: false, Eliminar: false, Exportar: false},
		}
	case RoleEmpresaFinanciero:
		return PermisosRole{
			Clientes:     PermisosModulo{Ver: true, Crear: false, Editar: false, Eliminar: false, Exportar: true},
			Dispositivos: PermisosModulo{Ver: true, Crear: false, Editar: false, Eliminar: false, Exportar: false},
			Alertas:      PermisosModulo{Ver: false, Crear: false, Editar: false, Eliminar: false, Exportar: false},
			Boletas:      PermisosModulo{Ver: true, Crear: true, Editar: true, Eliminar: false, Exportar: true},
			Tickets:      PermisosModulo{Ver: true, Crear: false, Editar: false, Eliminar: false, Exportar: false},
			Reportes:     PermisosModulo{Ver: true, Crear: false, Editar: false, Eliminar: false, Exportar: true},
			Configuracion: PermisosModulo{Ver: false, Crear: false, Editar: false, Eliminar: false, Exportar: false},
			Usuarios:     PermisosModulo{Ver: false, Crear: false, Editar: false, Eliminar: false, Exportar: false},
		}
	default:
		return PermisosRole{}
	}
}
