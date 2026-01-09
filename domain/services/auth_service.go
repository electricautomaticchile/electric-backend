package services

import (
	"context"
	"crypto/rand"
	"electric-backend/api/v1/recipe"
	"electric-backend/config"
	"electric-backend/domain/models"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/email"
	"electric-backend/infrastructure/entities"
	"electric-backend/infrastructure/validation"
	"electric-backend/types"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	empresaRepo       ports.PortEmpresa
	clienteRepo       ports.PortCliente
	usuarioEmpresaRepo ports.PortUsuarioEmpresa
	recoveryTokenRepo ports.PortRecoveryToken
	emailService      *email.ResendService
}

func NewAuthService(empresaRepo ports.PortEmpresa, clienteRepo ports.PortCliente, usuarioEmpresaRepo ports.PortUsuarioEmpresa, recoveryTokenRepo ports.PortRecoveryToken, emailService *email.ResendService) *AuthService {
	return &AuthService{
		empresaRepo:       empresaRepo,
		clienteRepo:       clienteRepo,
		usuarioEmpresaRepo: usuarioEmpresaRepo,
		recoveryTokenRepo: recoveryTokenRepo,
		emailService:      emailService,
	}
}

func (s *AuthService) Login(ctx context.Context, r *recipe.LoginRecipe) (*models.LoginResponseModel, error) {
	numeroCliente := validation.SanitizeNumeroCliente(r.NumeroCliente)

	if numeroCliente == "" {
		return nil, types.ThrowRecipe("Número de cliente es requerido", "numeroCliente")
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
	empresa, err := s.empresaRepo.FindByID(ctx, userID)
	if err == nil && empresa != nil {
		if !empresa.PasswordTemporal && r.PasswordActual != "" {
			err = bcrypt.CompareHashAndPassword([]byte(empresa.Password), []byte(r.PasswordActual))
			if err != nil {
				return types.ThrowData("Contraseña actual incorrecta")
			}
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.PasswordNuevo), 12)
		if err != nil {
			return err
		}

		return s.empresaRepo.UpdatePassword(ctx, userID, string(hashedPassword))
	}

	cliente, err := s.clienteRepo.FindByID(ctx, userID)
	if err != nil {
		return types.ThrowData("Usuario no encontrado")
	}

	esPasswordTemporal := cliente.PasswordTemporal != ""
	
	if !esPasswordTemporal && cliente.Password != "" && r.PasswordActual != "" {
		err = bcrypt.CompareHashAndPassword([]byte(cliente.Password), []byte(r.PasswordActual))
		if err != nil {
			return types.ThrowData("Contraseña actual incorrecta")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.PasswordNuevo), 12)
	if err != nil {
		return err
	}

	return s.clienteRepo.UpdatePassword(ctx, userID, string(hashedPassword))
}

func (s *AuthService) SolicitarRecuperacion(ctx context.Context, r *recipe.SolicitarRecuperacionRecipe) error {
	var usuarioID primitive.ObjectID
	var nombre, correo string

	cliente, err := s.clienteRepo.FindByCorreo(ctx, r.Email)
	if err == nil && cliente != nil {
		usuarioID, _ = primitive.ObjectIDFromHex(cliente.ID)
		nombre = cliente.Nombre
		correo = cliente.Correo
	} else {
		empresa, err := s.empresaRepo.FindByCorreo(ctx, r.Email)
		if err != nil {
			return nil
		}
		usuarioID, _ = primitive.ObjectIDFromHex(empresa.ID)
		nombre = empresa.NombreEmpresa
		correo = empresa.Correo
	}

	token := s.generarTokenRecuperacion()
	
	recoveryToken := &entities.RecoveryTokenEntity{
		UsuarioID:       usuarioID,
		Token:           token,
		FechaExpiracion: time.Now().Add(1 * time.Hour),
		Usado:           false,
	}

	if err := s.recoveryTokenRepo.Create(ctx, recoveryToken); err != nil {
		return err
	}

	go s.emailService.EnviarRecuperacionPassword(correo, nombre, token)

	return nil
}

func (s *AuthService) RestablecerPassword(ctx context.Context, r *recipe.RestablecerPasswordRecipe) error {
	recoveryToken, err := s.recoveryTokenRepo.FindByToken(ctx, r.Token)
	if err != nil {
		return types.ThrowData("Token inválido o expirado")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.Password), 12)
	if err != nil {
		return err
	}

	usuarioID := recoveryToken.UsuarioID.Hex()

	empresa, err := s.empresaRepo.FindByID(ctx, usuarioID)
	if err == nil && empresa != nil {
		if err := s.empresaRepo.UpdatePassword(ctx, usuarioID, string(hashedPassword)); err != nil {
			return err
		}
	} else {
		if err := s.clienteRepo.UpdatePassword(ctx, usuarioID, string(hashedPassword)); err != nil {
			return err
		}
	}

	return s.recoveryTokenRepo.MarkAsUsed(ctx, recoveryToken.ID.Hex())
}

func (s *AuthService) generarTokenRecuperacion() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *AuthService) RegistrarEmpresa(ctx context.Context, r *recipe.RegistroEmpresaRecipe) (*models.EmpresaModel, error) {
	rutNormalizado := validation.NormalizarRUT(r.Rut)
	
	existente, _ := s.empresaRepo.FindByNumeroCliente(ctx, rutNormalizado)
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
		NombreEmpresa: validation.SanitizeString(r.NombreEmpresa),
		RazonSocial:   validation.SanitizeString(r.RazonSocial),
		Rut:           rutNormalizado,
		Correo:        validation.SanitizeEmail(r.Correo),
		Telefono:      validation.NormalizarTelefono(r.Telefono),
		Direccion:     validation.SanitizeString(r.Direccion),
		Ciudad:        validation.SanitizeString(r.Ciudad),
		Region:        validation.SanitizeString(r.Region),
		ContactoPrincipal: models.ContactoPrincipal{
			Nombre:   validation.SanitizeString(r.ContactoPrincipal.Nombre),
			Cargo:    validation.SanitizeString(r.ContactoPrincipal.Cargo),
			Telefono: validation.NormalizarTelefono(r.ContactoPrincipal.Telefono),
			Correo:   validation.SanitizeEmail(r.ContactoPrincipal.Correo),
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

	usuarioAdmin := &models.UsuarioEmpresaModel{
		EmpresaID:        empresa.ID,
		Nombre:           r.ContactoPrincipal.Nombre,
		Email:            validation.SanitizeEmail(r.ContactoPrincipal.Correo),
		Password:         string(hashedPassword),
		Role:             models.RoleEmpresaAdmin,
		Telefono:         validation.NormalizarTelefono(r.ContactoPrincipal.Telefono),
		Cargo:            validation.SanitizeString(r.ContactoPrincipal.Cargo),
		PasswordTemporal: true,
	}

	if err := s.usuarioEmpresaRepo.Create(ctx, usuarioAdmin); err != nil {
		return nil, err
	}

	empresa.Password = passwordTemporal

	return empresa, nil
}

func (s *AuthService) generarNumeroCliente() string {
	numero := time.Now().UnixNano()%9000000 + 1000000
	dv := s.calcularDigitoVerificador(int(numero))
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

func (s *AuthService) LoginEmpresa(ctx context.Context, r *recipe.LoginEmpresaRecipe) (*models.LoginResponseModel, error) {
	email := validation.SanitizeEmail(r.Email)

	if email == "" {
		return nil, types.ThrowRecipe("Email es requerido", "email")
	}

	empresa, err := s.empresaRepo.FindByCorreo(ctx, email)
	if err != nil {
		return nil, types.ThrowAuth("Credenciales inválidas")
	}

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

	permisos := models.GetPermisosRole(empresa.Role)

	return &models.LoginResponseModel{
		Token:        token,
		RefreshToken: token,
		User: &models.ClienteModel{
			ID:        empresa.ID,
			Nombre:    empresa.NombreEmpresa,
			Correo:    empresa.Correo,
			Role:      empresa.Role,
			EmpresaID: empresa.ID,
			Activo:    empresa.Estado == "activo",
			TipoCliente: empresa.TipoUsuario,
		},
		Permisos: &permisos,
	}, nil
}

func (s *AuthService) generateTokenUsuarioEmpresa(usuario *models.UsuarioEmpresaModel) (string, error) {
	claims := jwt.MapClaims{
		"userId":    usuario.ID,
		"userRole":  usuario.Role,
		"userType":  "usuario_empresa",
		"empresaId": usuario.EmpresaID,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}
