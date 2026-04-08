package services

import (
	"context"
	"electric-backend/types"
)

// PowerChecker provides permission checking for services.
// Embed this in any service that needs permission verification.
type PowerChecker struct{}

// Power constants
const (
	PowerClienteVer      = "clientes:ver"
	PowerClienteCrear    = "clientes:crear"
	PowerClienteEditar   = "clientes:editar"
	PowerClienteEliminar = "clientes:eliminar"
	PowerDispositivoVer      = "dispositivos:ver"
	PowerDispositivoCrear    = "dispositivos:crear"
	PowerDispositivoEditar   = "dispositivos:editar"
	PowerDispositivoEliminar = "dispositivos:eliminar"
	PowerBoletaVer           = "boletas:ver"
	PowerBoletaCrear         = "boletas:crear"
	PowerReporteVer          = "reportes:ver"
	PowerConfigVer           = "configuracion:ver"
	PowerConfigEditar        = "configuracion:editar"
	PowerUsuarioVer          = "usuarios:ver"
	PowerUsuarioCrear        = "usuarios:crear"
	PowerUsuarioEditar       = "usuarios:editar"
	PowerUsuarioEliminar     = "usuarios:eliminar"
)

// EveryPower checks if the user in context has ALL specified powers.
func (pc PowerChecker) EveryPower(ctx context.Context, powers ...string) bool {
	userPowers, ok := ctx.Value(types.ContextKeyPowers).([]string)
	if !ok || len(userPowers) == 0 {
		return false
	}
	powerSet := make(map[string]bool, len(userPowers))
	for _, p := range userPowers {
		powerSet[p] = true
	}
	for _, required := range powers {
		if !powerSet[required] {
			return false
		}
	}
	return true
}

// SomePower checks if the user in context has AT LEAST ONE of the specified powers.
func (pc PowerChecker) SomePower(ctx context.Context, powers ...string) bool {
	userPowers, ok := ctx.Value(types.ContextKeyPowers).([]string)
	if !ok || len(userPowers) == 0 {
		return false
	}
	powerSet := make(map[string]bool, len(userPowers))
	for _, p := range userPowers {
		powerSet[p] = true
	}
	for _, required := range powers {
		if powerSet[required] {
			return true
		}
	}
	return false
}
