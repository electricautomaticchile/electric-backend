package server

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"

	"electric-backend/infrastructure/iot"
	"electric-backend/infrastructure/metrics"

	"github.com/gin-gonic/gin"
)

// registerMetricsRoute monta GET /metrics a nivel raíz (fuera de /api) para que
// Prometheus / Grafana Cloud pueda scrapear. Se expone en formato de texto de
// exposición Prometheus construido manualmente (sin dependencias externas).
//
// Auth: por defecto es un endpoint abierto (Prometheus scrapea sin sesión). Si
// la variable de entorno METRICS_TOKEN está seteada, se exige el header
// Authorization: Bearer <METRICS_TOKEN>.
func registerMetricsRoute(router *gin.Engine) {
	router.GET("/metrics", func(c *gin.Context) {
		if token := os.Getenv("METRICS_TOKEN"); token != "" {
			if c.GetHeader("Authorization") != "Bearer "+token {
				c.String(http.StatusUnauthorized, "unauthorized")
				return
			}
		}
		c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		c.String(http.StatusOK, renderMetrics())
	})
}

// renderMetrics construye el cuerpo del endpoint /metrics en formato Prometheus.
func renderMetrics() string {
	var b strings.Builder

	// --- Ingestor IoT ---
	var ingestStats map[string]uint64
	if ingestor := iot.DefaultReadingIngestor(); ingestor != nil {
		ingestStats = ingestor.Stats()
	} else {
		ingestStats = map[string]uint64{}
	}

	writeMetric(&b, "iot_ingest_enqueued_total",
		"Total de lecturas IoT encoladas para persistencia.", "counter",
		ingestStats["enqueued"])
	writeMetric(&b, "iot_ingest_flushed_total",
		"Total de lecturas IoT persistidas (flushed) en la base de datos.", "counter",
		ingestStats["flushed"])
	writeMetric(&b, "iot_ingest_dropped_total",
		"Total de lecturas IoT descartadas por cola llena.", "counter",
		ingestStats["dropped"])
	writeMetric(&b, "iot_ingest_failed_total",
		"Total de lecturas IoT que fallaron al persistir.", "counter",
		ingestStats["failed"])
	writeMetric(&b, "iot_ingest_queued",
		"Lecturas IoT actualmente en cola esperando persistencia.", "gauge",
		ingestStats["queued"])

	// --- Runtime Go ---
	writeMetric(&b, "go_goroutines",
		"Número de goroutines que existen actualmente.", "gauge",
		uint64(runtime.NumGoroutine()))

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	writeMetric(&b, "go_memstats_alloc_bytes",
		"Bytes de heap asignados y todavía en uso.", "gauge",
		mem.Alloc)
	writeMetric(&b, "go_memstats_sys_bytes",
		"Bytes de memoria obtenidos del sistema operativo.", "gauge",
		mem.Sys)

	// --- Peticiones HTTP por código ---
	httpByCode := metrics.HTTPRequestsSnapshot()
	b.WriteString("# HELP http_requests_total Total de peticiones HTTP atendidas, por código de estado.\n")
	b.WriteString("# TYPE http_requests_total counter\n")
	codes := make([]int, 0, len(httpByCode))
	for code := range httpByCode {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	for _, code := range codes {
		fmt.Fprintf(&b, "http_requests_total{code=\"%d\"} %d\n", code, httpByCode[code])
	}

	return b.String()
}

// writeMetric escribe una métrica escalar con sus líneas HELP/TYPE.
func writeMetric(b *strings.Builder, name, help, typ string, value uint64) {
	fmt.Fprintf(b, "# HELP %s %s\n", name, help)
	fmt.Fprintf(b, "# TYPE %s %s\n", name, typ)
	fmt.Fprintf(b, "%s %d\n", name, value)
}
