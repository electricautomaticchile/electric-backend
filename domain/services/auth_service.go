package services

import (
	"context"
	"electric-backend/api/v1/recipe"
	"electric-backend/config"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/types"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	empresaRepo       ports.PortEmpresa
	clienteRepo       ports.PortCliente
	recoveryTokenRepo ports.PortRecoveryToken
}

func NewAuthService(empresaRepo ports.PortEmpresa, clienteRepo ports.PortCliente, recoveryTokenRepo ports.PortRecoveryToken) *AuthService {
	return &AuthService{
		empresaRepo:       empresaRepo,
		clienteRepo:       clienteRepo,
		recoveryTokenRepo: recoveryTokenRepo,
	}
}

func (s *AuthService) Login(ctx context.Context, r *recipe.LoginRecipe) (*models.LoginResponseModel, error) {
	numeroCliente := r.NumeroCliente

	if numeroCliente == "" {
		return nil, types.ThrowRecipe("Número de cliente es requerido", "numeroCliente")
	}

	empresa, err := s.empresaRepo.FindByNumeroCliente(ctx, numeroCliente)
	if err == nil && empresa != nil {
		if empresa.Estado != "activo" {
			return nil, types.ThrowAuth("Empresa inactiva")
		}

		err = bcrypt.CompareHashAndPassword([]byte(empresa.Password), []byte(r.Password))
		if err != nil {
			return nil, types.ThrowAuth("Credenciales inválidas")
		}

		s.empresaRepo.UpdateUltimoAcceso(ctx, empresa.ID)

		token, err := s.generateTokenEmpresa(empresa)
		if err != nil {
			return nil, err
		}

		return &models.LoginResponseModel{
			Token:        token,
			RefreshToken: token,
			User:         s.empresaToCliente(empresa),
		}, nil
	}

	cliente, err := s.clienteRepo.FindByNumeroCliente(ctx, numeroCliente)
	if err != nil {
		return nil, types.ThrowAuth("Credenciales inválidas")
	}

	if !cliente.Activo {
		return nil, types.ThrowAuth("Cliente inactivo")
	}

	if cliente.Password == "" {
		return nil, types.ThrowAuth("Cliente sin contraseña configurada")
	}

	err = bcrypt.CompareHashAndPassword([]byte(cliente.Password), []byte(r.Password))
	if err != nil {
		return nil, types.ThrowAuth("Credenciales inválidas")
	}

	s.clienteRepo.UpdateUltimoAcceso(ctx, cliente.ID)

	token, err := s.generateTokenCliente(cliente)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponseModel{
		Token:        token,
		RefreshToken: token,
		User:         cliente,
	}, nil
}

func (s *AuthService) ObtenerPerfil(ctx context.Context, userID string) (*models.ClienteModel, error) {
	// Primero buscar como cliente
	cliente, err := s.clienteRepo.FindByID(ctx, userID)
	if err == nil && cliente != nil {
		return cliente, nil
	}

	// Si no es cliente, buscar como empresa y convertir
	empresa, err := s.empresaRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, types.ThrowData("Usuario no encontrado")
	}

	return s.empresaToCliente(empresa), nil
}

func (s *AuthService) CambiarPassword(ctx context.Context, userID string, r *recipe.CambiarPasswordRecipe) error {
	// Intentar como empresa primero
	empresa, err := s.empresaRepo.FindByID(ctx, userID)
	if err == nil && empresa != nil {
		err = bcrypt.CompareHashAndPassword([]byte(empresa.Password), []byte(r.PasswordActual))
		if err != nil {
			return types.ThrowData("Contraseña actual incorrecta")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.PasswordNuevo), 12)
		if err != nil {
			return err
		}

		return s.empresaRepo.UpdatePassword(ctx, userID, string(hashedPassword))
	}

	// Si no es empresa, intentar como cliente
	cliente, err := s.clienteRepo.FindByID(ctx, userID)
	if err != nil {
		return types.ThrowData("Usuario no encontrado")
	}

	err = bcrypt.CompareHashAndPassword([]byte(cliente.Password), []byte(r.PasswordActual))
	if err != nil {
		return types.ThrowData("Contraseña actual incorrecta")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.PasswordNuevo), 12)
	if err != nil {
		return err
	}

	return s.clienteRepo.UpdatePassword(ctx, userID, string(hashedPassword))
}

func (s *AuthService) SolicitarRecuperacion(ctx context.Context, r *recipe.SolicitarRecuperacionRecipe) error {
	// TODO: Implementar recuperación de contraseña
	return nil
}

func (s *AuthService) RestablecerPassword(ctx context.Context, r *recipe.RestablecerPasswordRecipe) error {
	// TODO: Implementar restablecimiento de contraseña
	return nil
}

func (s *AuthService) RegistrarEmpresa(ctx context.Context, r *recipe.RegistroEmpresaRecipe) (*models.EmpresaModel, error) {
	existente, _ := s.empresaRepo.FindByNumeroCliente(ctx, r.Rut)
	if existente != nil {
		return nil, types.ThrowData("Ya existe una empresa con este RUT")
	}

	passwordTemporal := s.generarPasswordTemporal()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordTemporal), 12)
	if err != nil {
		return nil, err
	}

	numeroCliente := s.generarNumeroCliente()

	empresa := &models.EmpresaModel{
		NombreEmpresa: r.NombreEmpresa,
		RazonSocial:   r.RazonSocial,
		Rut:           r.Rut,
		Correo:        r.Correo,
		Telefono:      r.Telefono,
		Direccion:     r.Direccion,
		Ciudad:        r.Ciudad,
		Region:        r.Region,
		ContactoPrincipal: models.ContactoPrincipal{
			Nombre:   r.ContactoPrincipal.Nombre,
			Cargo:    r.ContactoPrincipal.Cargo,
			Telefono: r.ContactoPrincipal.Telefono,
			Correo:   r.ContactoPrincipal.Correo,
		},
		NumeroCliente:    numeroCliente,
		Password:         string(hashedPassword),
		PasswordTemporal: true,
		Role:             "empresa",
		TipoUsuario:      "empresa",
		Estado:           "activo",
		FechaCreacion:    time.Now(),
	}

	if err := s.empresaRepo.Create(ctx, empresa); err != nil {
		return nil, err
	}

	empresa.Password = passwordTemporal

	return empresa, nil
}

func (s *AuthService) generarNumeroCliente() string {
	rand.Seed(time.Now().UnixNano())
	numero := rand.Intn(9000000) + 1000000
	dv := s.calcularDigitoVerificador(numero)
	return fmt.Sprintf("%d-%s", numero, dv)
}

func (s *AuthService) calcularDigitoVerificador(numero int) string {
	suma := 0
	multiplicador := 2
	numeroStr := fmt.Sprintf("%d", numero)
	
	for i := len(numeroStr) - 1; i >= 0; i-- {
		digito := int(numeroStr[i] - '0')
		suma += digito * multiplicador
		multiplicador++
		if multiplicador > 7 {
			multiplicador = 2
		}
	}
	
	resto := suma % 11
	dv := 11 - resto
	
	if dv == 11 {
		return "0"
	} else if dv == 10 {
		return "K"
	}
	return fmt.Sprintf("%d", dv)
}

func (s *AuthService) generarPasswordTemporal() string {
	return "Temp" + time.Now().Format("0601021504")
}

// Convertir empresa a cliente para respuesta uniforme
func (s *AuthService) empresaToCliente(empresa *models.EmpresaModel) *models.ClienteModel {
	return &models.ClienteModel{
		ID:            empresa.ID,
		Nombre:        empresa.NombreEmpresa,
		Correo:        empresa.Correo,
		NumeroCliente: empresa.NumeroCliente,
		Telefono:      empresa.Telefono,
		Role:          empresa.Role,
		TipoUsuario:   "empresa",
		Activo:        empresa.Estado == "activo",
		Password:      empresa.Password,
		EmpresaID:     empresa.ID,
	}
}

func (s *AuthService) generateTokenEmpresa(empresa *models.EmpresaModel) (string, error) {
	claims := jwt.MapClaims{
		"userId":    empresa.ID,
		"userRole":  empresa.Role,
		"userType":  "empresa",
		"empresaId": empresa.ID,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

func (s *AuthService) generateTokenCliente(cliente *models.ClienteModel) (string, error) {
	claims := jwt.MapClaims{
		"userId":    cliente.ID,
		"userRole":  cliente.Role,
		"userType":  "cliente", // Tipo cliente
		"empresaId": cliente.EmpresaID,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}
