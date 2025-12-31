package types

// ContextKey es el tipo para las claves del contexto
type ContextKey string

const (
	// ContextKeyUserID es la clave para el ID del usuario en el contexto
	ContextKeyUserID ContextKey = "userId"
	
	// ContextKeyUserRole es la clave para el rol del usuario
	ContextKeyUserRole ContextKey = "userRole"
	
	// ContextKeyUserType es la clave para el tipo de usuario
	ContextKeyUserType ContextKey = "userType"
	
	// ContextKeyPowers es la clave para los permisos del usuario
	ContextKeyPowers ContextKey = "powers"
	
	// ContextKeyEmpresaID es la clave para el ID de la empresa
	ContextKeyEmpresaID ContextKey = "empresaId"
)
