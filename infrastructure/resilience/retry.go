package resilience

import (
	"fmt"
	"log"
	"math"
	"time"
)

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	BaseDelay:   500 * time.Millisecond,
	MaxDelay:    10 * time.Second,
	Multiplier:  2.0,
}

// RetryWithBackoff executes fn with exponential backoff.
// Returns the last error if all attempts fail.
func RetryWithBackoff(name string, cfg RetryConfig, fn func() error) error {
	var lastErr error
	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if attempt == cfg.MaxAttempts {
			break
		}

		delay := time.Duration(float64(cfg.BaseDelay) * math.Pow(cfg.Multiplier, float64(attempt-1)))
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}

		log.Printf("⚠️ [%s] intento %d/%d falló: %v — reintentando en %v",
			name, attempt, cfg.MaxAttempts, lastErr, delay)
		time.Sleep(delay)
	}

	return fmt.Errorf("[%s] falló después de %d intentos: %w", name, cfg.MaxAttempts, lastErr)
}

// ExecuteWithRetryAndCB combines circuit breaker + retry.
func ExecuteWithRetryAndCB(cb *CircuitBreaker, cfg RetryConfig, fn func() error) error {
	return cb.Execute(func() error {
		return RetryWithBackoff(cb.Name(), cfg, fn)
	})
}
