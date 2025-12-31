package services

import (
	"context"
	"electric-backend/domain/ports"
	"electric-backend/types"
	"strings"
)

type BaseService struct {
	powerRepo ports.PortPower
}

func NewBaseService(powerRepo ports.PortPower) *BaseService {
	return &BaseService{
		powerRepo: powerRepo,
	}
}

func (s *BaseService) GetPowersFromContext(ctx context.Context) []string {
	powers, ok := ctx.Value(types.ContextKeyPowers).([]string)
	if !ok {
		return []string{}
	}
	return powers
}

func (s *BaseService) GetUserFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(types.ContextKeyUserID).(string)
	if !ok {
		return ""
	}
	return userID
}

func (s *BaseService) HasPower(ctx context.Context, power string) bool {
	powers := s.GetPowersFromContext(ctx)
	for _, p := range powers {
		if p == power {
			return true
		}
	}
	return false
}

func (s *BaseService) HasAnyPower(ctx context.Context, requiredPowers []string) bool {
	powers := s.GetPowersFromContext(ctx)
	for _, required := range requiredPowers {
		for _, p := range powers {
			if p == required {
				return true
			}
		}
	}
	return false
}

func (s *BaseService) HasPowerPrefix(ctx context.Context, prefix string) bool {
	powers := s.GetPowersFromContext(ctx)
	for _, p := range powers {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

func (s *BaseService) EveryPower(ctx context.Context, requiredPowers ...string) error {
	powers := s.GetPowersFromContext(ctx)
	powerMap := make(map[string]bool)
	for _, p := range powers {
		powerMap[p] = true
	}

	for _, required := range requiredPowers {
		if !powerMap[required] {
			return types.ThrowPower("No tienes permisos suficientes")
		}
	}

	return nil
}

func (s *BaseService) SomePower(ctx context.Context, requiredPowers ...string) error {
	if s.HasAnyPower(ctx, requiredPowers) {
		return nil
	}
	return types.ThrowPower("No tienes permisos suficientes")
}
