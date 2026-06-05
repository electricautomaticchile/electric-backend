package services

import (
	"context"
	"crypto/rand"
	"electric-backend/api/v1/recipe"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/email"
	"electric-backend/infrastructure/validation"
	"electric-backend/types"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type UsuarioEmpresaService struct {
	usuarioRepo  ports.PortUsuarioEmpresa
	emailService email.EmailService
}

func NewUsuarioEmpresaService(usuarioRepo ports.PortUsuarioEmpresa, emailService email.EmailService) *UsuarioEmpresaService {
	return &UsuarioEmpresaService{
		usuarioRepo:  usuarioRepo,
		emailService: emailService,
	}
}

func (s *UsuarioEmpresaService) ObtenerTodos(ctx context.Context, empresaID string) ([]*models.UsuarioEmpresaModel, error) {
	usuarios, err := s.usuarioRepo.FindAll(ctx, empresaID)
	if err != nil {
		return []*models.UsuarioEmpresaModel{}, nil
	}

	for _, usuario := range usuarios {
		usuario.Password = ""
	}

	return usuarios, nil
}

func (s *UsuarioEmpresaService) ObtenerPorID(ctx context.Context, id string) (*models.UsuarioEmpresaModel, error) {
	usuario, err := s.usuarioRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	usuario.Password = ""
	return usuario, nil
}

func (s *UsuarioEmpresaService) Crear(ctx context.Context, empresaID string, r *recipe.CrearUsuarioEmpresaRecipe) (*models.UsuarioEmpresaModel, error) {
	existente, _ := s.usuarioRepo.FindByEmail(ctx, validation.SanitizeEmail(r.Email))
	if existente != nil {
		return nil, types.ThrowData("Ya existe un usuario con este email")
	}

	passwordTemporal, err := s.generarPasswordTemporal()
	if err != nil {
		return nil, err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordTemporal), 12)
	if err != nil {
		return nil, err
	}

	telefonoNormalizado := ""
	if r.Telefono != "" {
		telefonoNormalizado = validation.NormalizarTelefono(r.Telefono)
	}

	usuario := &models.UsuarioEmpresaModel{
		EmpresaID:        empresaID,
		Nombre:           validation.SanitizeString(r.Nombre),
		Email:            validation.SanitizeEmail(r.Email),
		Password:         string(hashedPassword),
		Role:             r.Role,
		Telefono:         telefonoNormalizado,
		Cargo:            validation.SanitizeString(r.Cargo),
		PasswordTemporal: true,
	}

	if err := s.usuarioRepo.Create(ctx, usuario); err != nil {
		return nil, err
	}

	go s.enviarCredencialesPorEmail(usuario.Email, usuario.Nombre, usuario.Email, passwordTemporal)

	usuario.Password = ""
	return usuario, nil
}

func (s *UsuarioEmpresaService) Actualizar(ctx context.Context, id string, r *recipe.ActualizarUsuarioEmpresaRecipe) (*models.UsuarioEmpresaModel, error) {
	usuario, err := s.usuarioRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if r.Nombre != "" {
		usuario.Nombre = validation.SanitizeString(r.Nombre)
	}
	if r.Telefono != "" {
		usuario.Telefono = validation.NormalizarTelefono(r.Telefono)
	}
	if r.Cargo != "" {
		usuario.Cargo = validation.SanitizeString(r.Cargo)
	}
	if r.Role != "" {
		usuario.Role = r.Role
	}
	if r.Activo != nil {
		usuario.Activo = *r.Activo
	}

	if err := s.usuarioRepo.Update(ctx, id, usuario); err != nil {
		return nil, err
	}

	usuario.Password = ""
	return usuario, nil
}

func (s *UsuarioEmpresaService) Eliminar(ctx context.Context, id string) error {
	return s.usuarioRepo.Delete(ctx, id)
}

func (s *UsuarioEmpresaService) generarPasswordTemporal() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "Tmp-" + hex.EncodeToString(bytes), nil
}

func (s *UsuarioEmpresaService) enviarCredencialesPorEmail(correo, nombre, email, passwordTemporal string) {
	err := s.emailService.EnviarCredenciales(correo, nombre, email, passwordTemporal)
	if err != nil {
		fmt.Printf("Error enviando credenciales por email: %v\n", err)
	}
}
