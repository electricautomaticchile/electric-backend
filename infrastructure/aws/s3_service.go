package aws

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"electric-backend/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Service struct {
	client *s3.Client
	config *config.Config
}

func NewS3Service(cfg *config.Config) (*S3Service, error) {
	// Usa el IAM Role del App Runner automáticamente (sin credenciales estáticas).
	// En desarrollo local, el SDK busca credenciales en ~/.aws/credentials o env vars AWS_*.
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)

	return &S3Service{
		client: client,
		config: cfg,
	}, nil
}

func (s *S3Service) SubirImagenPerfil(file multipart.File, header *multipart.FileHeader, tipoUsuario string, userID string) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}

	if !allowedExts[ext] {
		return "", fmt.Errorf("tipo de archivo no permitido: %s", ext)
	}

	if header.Size > 5*1024*1024 {
		return "", fmt.Errorf("archivo demasiado grande, máximo 5MB")
	}

	fileName := fmt.Sprintf("%s/%s/%s%s", tipoUsuario, userID, uuid.New().String(), ext)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("error leyendo archivo: %w", err)
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.config.S3BucketImages),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("error subiendo archivo a S3: %w", err)
	}

	imageURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		s.config.S3BucketImages,
		s.config.AWSRegion,
		fileName,
	)

	return imageURL, nil
}

func (s *S3Service) EliminarImagen(imageURL string) error {
	if imageURL == "" {
		return nil
	}

	key := s.extractKeyFromURL(imageURL)
	if key == "" {
		return fmt.Errorf("URL inválida")
	}

	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.S3BucketImages),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("error eliminando archivo de S3: %w", err)
	}

	return nil
}

func (s *S3Service) GenerarURLFirmada(key string, duracion time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.config.S3BucketImages),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duracion
	})
	if err != nil {
		return "", fmt.Errorf("error generando URL firmada: %w", err)
	}

	return request.URL, nil
}

func (s *S3Service) extractKeyFromURL(url string) string {
	parts := strings.Split(url, ".amazonaws.com/")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
