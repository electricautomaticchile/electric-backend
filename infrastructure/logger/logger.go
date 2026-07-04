package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Log es el logger estructurado global de la aplicación.
var Log zerolog.Logger

// Init configura zerolog:
//   - En producción: salida JSON (ideal para CloudWatch/Render logs, filtrable
//     por campos como empresaId, clienteId, dispositivoId).
//   - En desarrollo: salida legible en consola con colores.
func Init() {
	zerolog.TimeFieldFormat = time.RFC3339

	if isProduction() {
		// JSON estructurado a stdout.
		Log = zerolog.New(os.Stdout).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
		Log = zerolog.New(output).With().Timestamp().Caller().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Permitir override del nivel por env var LOG_LEVEL (debug, info, warn, error).
	if lvl, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL")); err == nil && os.Getenv("LOG_LEVEL") != "" {
		zerolog.SetGlobalLevel(lvl)
	}
}

func isProduction() bool {
	return os.Getenv("NODE_ENV") == "production" || os.Getenv("ENV") == "production"
}

func Info() *zerolog.Event  { return Log.Info() }
func Debug() *zerolog.Event { return Log.Debug() }
func Error() *zerolog.Event { return Log.Error() }
func Warn() *zerolog.Event  { return Log.Warn() }
func Fatal() *zerolog.Event { return Log.Fatal() }
