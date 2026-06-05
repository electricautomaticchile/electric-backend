package controllers

import (
	"testing"
	"time"
)

func TestSanitizeLeadExtraRemovesUnsafeKeysAndValues(t *testing.T) {
	extra := sanitizeLeadExtra(map[string]interface{}{
		"$where": "bad",
		"role":   "<b>Operaciones</b>",
		"nested": map[string]interface{}{
			"meterCount": "<script>1000</script>",
			"$bad":       "drop",
		},
		"unsupported": struct{}{},
	})

	if _, ok := extra["$where"]; ok {
		t.Fatal("expected mongo operator key to be removed")
	}
	if _, ok := extra["unsupported"]; ok {
		t.Fatal("expected unsupported values to be removed")
	}
	if extra["role"] != "Operaciones" {
		t.Fatalf("expected sanitized role, got %v", extra["role"])
	}
	nested, ok := extra["nested"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected sanitized nested map, got %T", extra["nested"])
	}
	if _, ok := nested["$bad"]; ok {
		t.Fatal("expected nested mongo operator key to be removed")
	}
}

func TestParseLeadDateSupportsDateAndRFC3339(t *testing.T) {
	dateOnly, err := parseLeadDate("2026-06-03", true)
	if err != nil {
		t.Fatalf("parse date-only: %v", err)
	}
	if dateOnly.Format(time.RFC3339Nano) != "2026-06-03T23:59:59.999999999Z" {
		t.Fatalf("expected end-of-day UTC, got %s", dateOnly.Format(time.RFC3339Nano))
	}

	withTime, err := parseLeadDate("2026-06-03T10:30:00Z", false)
	if err != nil {
		t.Fatalf("parse rfc3339: %v", err)
	}
	if withTime.Format(time.RFC3339) != "2026-06-03T10:30:00Z" {
		t.Fatalf("expected rfc3339 value, got %s", withTime.Format(time.RFC3339))
	}
}
