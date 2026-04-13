package resilience

import "time"

// Global circuit breakers for external services
var (
	CBSMS   = NewCircuitBreaker("SNS-SMS", WithMaxFailures(3), WithResetTimeout(60*time.Second))
	CBEmail = NewCircuitBreaker("SES-Email", WithMaxFailures(5), WithResetTimeout(30*time.Second))
	CBMongo = NewCircuitBreaker("MongoDB", WithMaxFailures(10), WithResetTimeout(15*time.Second))
)

var SMSRetryConfig = RetryConfig{
	MaxAttempts: 3,
	BaseDelay:   1 * time.Second,
	MaxDelay:    10 * time.Second,
	Multiplier:  2.0,
}

var EmailRetryConfig = RetryConfig{
	MaxAttempts: 3,
	BaseDelay:   500 * time.Millisecond,
	MaxDelay:    5 * time.Second,
	Multiplier:  2.0,
}

// SendSMSWithResilience wraps SMS sending with circuit breaker + retry.
func SendSMSWithResilience(fn func() error) error {
	return ExecuteWithRetryAndCB(CBSMS, SMSRetryConfig, fn)
}

// SendEmailWithResilience wraps email sending with circuit breaker + retry.
func SendEmailWithResilience(fn func() error) error {
	return ExecuteWithRetryAndCB(CBEmail, EmailRetryConfig, fn)
}
