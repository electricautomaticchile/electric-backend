package services

import (
	"context"
	"electric-backend/domain/ports"
	"fmt"
	"mime/multipart"
)

type profileImageStorage interface {
	SubirImagenPerfil(file multipart.File, header *multipart.FileHeader, tipoUsuario string, userID string) (string, error)
	EliminarImagen(imageURL string) error
}

type ImagenPerfilService struct {
	storage     profileImageStorage
	clienteRepo ports.PortCliente
	empresaRepo ports.PortEmpresa
}

func NewImagenPerfilService(
	clienteRepo ports.PortCliente,
	empresaRepo ports.PortEmpresa,
	storage profileImageStorage,
) *ImagenPerfilService {
	return &ImagenPerfilService{
		storage:     storage,
		clienteRepo: clienteRepo,
		empresaRepo: empresaRepo,
	}
}

func (s *ImagenPerfilService) SubirImagenPerfil(file multipart.File, header *multipart.FileHeader, tipoUsuario string, userID string) (string, error) {
	if s.storage == nil {
		return "", fmt.Errorf("servicio de imágenes no configurado")
	}

	imageURL, err := s.storage.SubirImagenPerfil(file, header, tipoUsuario, userID)
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

		if cliente.ImagenPerfil != "" && s.storage != nil {
			if err := s.storage.EliminarImagen(cliente.ImagenPerfil); err != nil {
				return fmt.Errorf("error eliminando imagen anterior: %w", err)
			}
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

		if empresa.ContactoPrincipal.ImagenPerfil != "" && s.storage != nil {
			if err := s.storage.EliminarImagen(empresa.ContactoPrincipal.ImagenPerfil); err != nil {
				return fmt.Errorf("error eliminando imagen anterior: %w", err)
			}
		}

		empresa.ContactoPrincipal.ImagenPerfil = imageURL
		if err := s.empresaRepo.Update(ctx, userID, empresa); err != nil {
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
		return empresa.ContactoPrincipal.ImagenPerfil, nil

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

	if imageURL != "" && s.storage != nil {
		if err := s.storage.EliminarImagen(imageURL); err != nil {
			return fmt.Errorf("error eliminando imagen: %w", err)
		}
	}

	switch tipoUsuario {
	case "cliente":
		cliente, err := s.clienteRepo.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		cliente.ImagenPerfil = ""
		return s.clienteRepo.Update(ctx, userID, cliente)

	case "empresa":
		empresa, err := s.empresaRepo.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		empresa.ContactoPrincipal.ImagenPerfil = ""
		return s.empresaRepo.Update(ctx, userID, empresa)

	default:
		return fmt.Errorf("tipo de usuario inválido: %s", tipoUsuario)
	}
}
