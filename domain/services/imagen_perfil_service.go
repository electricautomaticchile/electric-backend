package services

import (
	"context"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/aws"
	"fmt"
	"mime/multipart"
)

type ImagenPerfilService struct {
	clienteRepo ports.PortCliente
	empresaRepo ports.PortEmpresa
	s3Service   *aws.S3Service
}

func NewImagenPerfilService(
	clienteRepo ports.PortCliente,
	empresaRepo ports.PortEmpresa,
	s3Service *aws.S3Service,
) *ImagenPerfilService {
	return &ImagenPerfilService{
		clienteRepo: clienteRepo,
		empresaRepo: empresaRepo,
		s3Service:   s3Service,
	}
}

func (s *ImagenPerfilService) SubirImagenPerfil(file multipart.File, header *multipart.FileHeader, tipoUsuario string, userID string) (string, error) {
	imageURL, err := s.s3Service.SubirImagenPerfil(file, header, tipoUsuario, userID)
	if err != nil {
		return "", err
	}

	return imageURL, nil
}

func (s *ImagenPerfilService) ActualizarImagenPerfil(imageURL string, tipoUsuario string, userID string) error {
	ctx := context.Background()

	switch tipoUsuario {
	case "cliente":
		cliente, err := s.clienteRepo.FindByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("cliente no encontrado: %w", err)
		}

		if cliente.ImagenPerfil != "" {
			_ = s.s3Service.EliminarImagen(cliente.ImagenPerfil)
		}

		cliente.ImagenPerfil = imageURL
		if err := s.clienteRepo.Update(ctx, userID, cliente); err != nil {
			return fmt.Errorf("error actualizando cliente: %w", err)
		}

	case "empresa":
		empresa, err := s.empresaRepo.FindByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("empresa no encontrada: %w", err)
		}

		if empresa.ImagenPerfil != "" {
			_ = s.s3Service.EliminarImagen(empresa.ImagenPerfil)
		}

		empresa.ImagenPerfil = imageURL
		if err := s.empresaRepo.Update(ctx, userID, empresa); err != nil {
			return fmt.Errorf("error actualizando empresa: %w", err)
		}

	default:
		return fmt.Errorf("tipo de usuario no válido: %s", tipoUsuario)
	}

	return nil
}

func (s *ImagenPerfilService) ObtenerImagenPerfil(tipoUsuario string, userID string) (string, error) {
	ctx := context.Background()

	switch tipoUsuario {
	case "cliente":
		cliente, err := s.clienteRepo.FindByID(ctx, userID)
		if err != nil {
			return "", fmt.Errorf("cliente no encontrado: %w", err)
		}
		return cliente.ImagenPerfil, nil

	case "empresa":
		empresa, err := s.empresaRepo.FindByID(ctx, userID)
		if err != nil {
			return "", fmt.Errorf("empresa no encontrada: %w", err)
		}
		return empresa.ImagenPerfil, nil

	default:
		return "", fmt.Errorf("tipo de usuario no válido: %s", tipoUsuario)
	}
}

func (s *ImagenPerfilService) EliminarImagenPerfil(tipoUsuario string, userID string) error {
	ctx := context.Background()

	switch tipoUsuario {
	case "cliente":
		cliente, err := s.clienteRepo.FindByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("cliente no encontrado: %w", err)
		}

		if cliente.ImagenPerfil != "" {
			if err := s.s3Service.EliminarImagen(cliente.ImagenPerfil); err != nil {
				return fmt.Errorf("error eliminando imagen de S3: %w", err)
			}
		}

		cliente.ImagenPerfil = ""
		if err := s.clienteRepo.Update(ctx, userID, cliente); err != nil {
			return fmt.Errorf("error actualizando cliente: %w", err)
		}

	case "empresa":
		empresa, err := s.empresaRepo.FindByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("empresa no encontrada: %w", err)
		}

		if empresa.ImagenPerfil != "" {
			if err := s.s3Service.EliminarImagen(empresa.ImagenPerfil); err != nil {
				return fmt.Errorf("error eliminando imagen de S3: %w", err)
			}
		}

		empresa.ImagenPerfil = ""
		if err := s.empresaRepo.Update(ctx, userID, empresa); err != nil {
			return fmt.Errorf("error actualizando empresa: %w", err)
		}

	default:
		return fmt.Errorf("tipo de usuario no válido: %s", tipoUsuario)
	}

	return nil
}
