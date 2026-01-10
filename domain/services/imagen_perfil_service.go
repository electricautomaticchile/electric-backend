package services

import (
	"context"
	"electric-backend/domain/ports"
	"electric-backend/infrastructure/aws"
	"fmt"
	"mime/multipart"
)

type ImagenPerfilService struct {
	s3Service   *aws.S3Service
	clienteRepo ports.PortCliente
	empresaRepo ports.PortEmpresa
}

func NewImagenPerfilService(
	clienteRepo ports.PortCliente,
	empresaRepo ports.PortEmpresa,
	s3Service *aws.S3Service,
) *ImagenPerfilService {
	return &ImagenPerfilService{
		s3Service:   s3Service,
		clienteRepo: clienteRepo,
		empresaRepo: empresaRepo,
	}
}

func (s *ImagenPerfilService) SubirImagenPerfil(file multipart.File, header *multipart.FileHeader, tipoUsuario string, userID string) (string, error) {
	if s.s3Service == nil {
		return "", fmt.Errorf("servicio S3 no configurado")
	}

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

		if cliente.ImagenPerfil != "" && s.s3Service != nil {
			if err := s.s3Service.EliminarImagen(cliente.ImagenPerfil); err != nil {
				return fmt.Errorf("error eliminando imagen anterior: %w", err)
			}
		}

		cliente.ImagenPerfil = imageURL
		if err := s.clienteRepo.Update(ctx, cliente); err != nil {
			return fmt.Errorf("error actualizando cliente: %w", err)
		}

	case "empresa":
		empresa, err := s.empresaRepo.FindByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("empresa no encontrada: %w", err)
		}

		if empresa.ImagenPerfil != "" && s.s3Service != nil {
			if err := s.s3Service.EliminarImagen(empresa.ImagenPerfil); err != nil {
				return fmt.Errorf("error eliminando imagen anterior: %w", err)
			}
		}

		empresa.ImagenPerfil = imageURL
		if err := s.empresaRepo.Update(ctx, empresa); err != nil {
			return fmt.Errorf("error actualizando empresa: %w", err)
		}

	default:
		return fmt.Errorf("tipo de usuario inválido: %s", tipoUsuario)
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
		return "", fmt.Errorf("tipo de usuario inválido: %s", tipoUsuario)
	}
}

func (s *ImagenPerfilService) EliminarImagenPerfil(tipoUsuario string, userID string) error {
	ctx := context.Background()

	imageURL, err := s.ObtenerImagenPerfil(tipoUsuario, userID)
	if err != nil {
		return err
	}

	if imageURL != "" && s.s3Service != nil {
		if err := s.s3Service.EliminarImagen(imageURL); err != nil {
			return fmt.Errorf("error eliminando imagen de S3: %w", err)
		}
	}

	switch tipoUsuario {
	case "cliente":
		cliente, err := s.clienteRepo.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		cliente.ImagenPerfil = ""
		return s.clienteRepo.Update(ctx, cliente)

	case "empresa":
		empresa, err := s.empresaRepo.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		empresa.ImagenPerfil = ""
		return s.empresaRepo.Update(ctx, empresa)

	default:
		return fmt.Errorf("tipo de usuario inválido: %s", tipoUsuario)
	}
}
