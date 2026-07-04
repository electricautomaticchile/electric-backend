package push

import (
	"context"
	"log"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FirebaseService implementa PushService usando Firebase Admin SDK.
type FirebaseService struct {
	client    *messaging.Client
	tokenRepo TokenEliminador
}

// NewFirebaseService inicializa el cliente de Firebase Cloud Messaging.
// Las credenciales se leen desde la variable de entorno FIREBASE_SERVICE_ACCOUNT
// (JSON completo de la service account) o, si no está presente, desde
// GOOGLE_APPLICATION_CREDENTIALS (ruta a un archivo JSON).
func NewFirebaseService(tokenRepo TokenEliminador) (*FirebaseService, error) {
	ctx := context.Background()

	var opts []option.ClientOption
	if raw := strings.TrimSpace(os.Getenv("FIREBASE_SERVICE_ACCOUNT")); raw != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(raw)))
	} else if path := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")); path != "" {
		opts = append(opts, option.WithCredentialsFile(path))
	}

	app, err := firebase.NewApp(ctx, nil, opts...)
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	log.Println("push: Firebase Cloud Messaging inicializado")
	return &FirebaseService{client: client, tokenRepo: tokenRepo}, nil
}

// EnviarPush envía la notificación a todos los tokens y limpia los que Firebase
// reporte como no registrados (registration-token-not-registered).
func (s *FirebaseService) EnviarPush(tokens []string, titulo, cuerpo string, data map[string]string) error {
	if len(tokens) == 0 {
		return nil
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: titulo,
			Body:  cuerpo,
		},
		Data: data,
	}

	ctx := context.Background()
	resp, err := s.client.SendEachForMulticast(ctx, message)
	if err != nil {
		return err
	}

	if resp.FailureCount > 0 {
		for i, r := range resp.Responses {
			if r.Success {
				continue
			}
			// Token inválido/no registrado: se elimina para no reintentar.
			if messaging.IsUnregistered(r.Error) || messaging.IsInvalidArgument(r.Error) {
				if s.tokenRepo != nil && i < len(tokens) {
					if delErr := s.tokenRepo.DeleteByToken(ctx, tokens[i]); delErr != nil {
						log.Printf("push: no se pudo eliminar token inválido: %v", delErr)
					}
				}
			} else {
				log.Printf("push: fallo enviando a token: %v", r.Error)
			}
		}
	}

	return nil
}
