package server

import (
	"strings"
	"testing"
)

// TestRenderMetrics verifica que el endpoint /metrics produce texto de
// exposición Prometheus válido con las métricas mínimas esperadas.
func TestRenderMetrics(t *testing.T) {
	out := renderMetrics()

	requeridas := []string{
		"iot_ingest_enqueued_total",
		"iot_ingest_flushed_total",
		"iot_ingest_dropped_total",
		"iot_ingest_failed_total",
		"iot_ingest_queued",
		"go_goroutines",
		"http_requests_total",
	}
	for _, m := range requeridas {
		if !strings.Contains(out, "# TYPE "+m+" ") {
			t.Errorf("falta la declaración # TYPE para %q en /metrics", m)
		}
		if !strings.Contains(out, m) {
			t.Errorf("falta la métrica %q en /metrics", m)
		}
	}

	// Cada bloque debe tener sus líneas HELP/TYPE.
	if !strings.Contains(out, "# HELP go_goroutines") {
		t.Error("falta la línea # HELP para go_goroutines")
	}
}
