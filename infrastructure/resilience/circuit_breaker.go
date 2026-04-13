package resilience

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type State int

const (
	StateClosed   State = iota // Normal — requests pass through
	StateOpen                  // Failing — requests blocked
	StateHalfOpen              // Testing — limited requests allowed
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

type CircuitBreaker struct {
	name          string
	mu            sync.RWMutex
	state         State
	failures      int
	successes     int
	maxFailures   int
	resetTimeout  time.Duration
	halfOpenMax   int
	lastFailure   time.Time
	onStateChange func(name string, from, to State)
}

type CBOption func(*CircuitBreaker)

func WithMaxFailures(n int) CBOption   { return func(cb *CircuitBreaker) { cb.maxFailures = n } }
func WithResetTimeout(d time.Duration) CBOption { return func(cb *CircuitBreaker) { cb.resetTimeout = d } }
func WithHalfOpenMax(n int) CBOption   { return func(cb *CircuitBreaker) { cb.halfOpenMax = n } }

func WithStateChangeCallback(fn func(string, State, State)) CBOption {
	return func(cb *CircuitBreaker) { cb.onStateChange = fn }
}

func NewCircuitBreaker(name string, opts ...CBOption) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:         name,
		state:        StateClosed,
		maxFailures:  5,
		resetTimeout: 30 * time.Second,
		halfOpenMax:  2,
	}
	for _, opt := range opts {
		opt(cb)
	}
	return cb
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allowRequest() {
		return fmt.Errorf("circuit breaker [%s] is open", cb.name)
	}

	err := fn()
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			// Transition to half-open (need write lock)
			cb.mu.RUnlock()
			cb.mu.Lock()
			if cb.state == StateOpen {
				cb.transition(StateHalfOpen)
			}
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		return cb.successes < cb.halfOpenMax
	}
	return true
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.state == StateHalfOpen {
		cb.transition(StateOpen)
		return
	}

	if cb.failures >= cb.maxFailures {
		cb.transition(StateOpen)
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.successes++
		if cb.successes >= cb.halfOpenMax {
			cb.transition(StateClosed)
		}
		return
	}

	cb.failures = 0
}

func (cb *CircuitBreaker) transition(to State) {
	from := cb.state
	cb.state = to
	cb.failures = 0
	cb.successes = 0
	log.Printf("⚡ Circuit breaker [%s]: %s → %s", cb.name, from, to)
	if cb.onStateChange != nil {
		go cb.onStateChange(cb.name, from, to)
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) Name() string { return cb.name }
