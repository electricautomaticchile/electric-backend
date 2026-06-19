package config

import (
	"os"
	"testing"
)

func TestLoadConfigReadsRequireRedis(t *testing.T) {
	t.Setenv("ENVIRONMENT", "")
	t.Setenv("JWT_SECRET", "test-secret-test-secret-test-secret")
	t.Setenv("REQUIRE_REDIS", "true")

	cfg := LoadConfig()

	if !cfg.RequireRedis {
		t.Fatal("expected REQUIRE_REDIS=true to enable strict Redis startup")
	}
}

func TestApplyRedisURLParsesRedissURL(t *testing.T) {
	t.Setenv("REDIS_URL", "rediss://:secret@redis.internal:6380/2")
	t.Setenv("REDIS_TLS", "")

	cfg := &Config{
		RedisHost: "localhost",
		RedisPort: "6379",
		RedisDB:   "0",
	}

	applyRedisURL(cfg)

	if cfg.RedisHost != "redis.internal" {
		t.Fatalf("expected RedisHost redis.internal, got %q", cfg.RedisHost)
	}
	if cfg.RedisPort != "6380" {
		t.Fatalf("expected RedisPort 6380, got %q", cfg.RedisPort)
	}
	if cfg.RedisPassword != "secret" {
		t.Fatalf("expected RedisPassword secret, got %q", cfg.RedisPassword)
	}
	if cfg.RedisDB != "2" {
		t.Fatalf("expected RedisDB 2, got %q", cfg.RedisDB)
	}
	if got := os.Getenv("REDIS_TLS"); got != "true" {
		t.Fatalf("expected REDIS_TLS=true for rediss URL, got %q", got)
	}
}
