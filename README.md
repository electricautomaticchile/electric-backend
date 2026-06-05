# Electric Backend

API principal de ElectricAutomaticChile. Expone autenticacion, clientes,
empresas, dispositivos IoT, lecturas, boletas, reportes, leads de la landing y
servicios administrativos.

## Requisitos

- Go 1.24 o superior.
- MongoDB disponible.
- Redis disponible para produccion.
- `k6` solo si se ejecutan pruebas de carga.
- Credenciales externas opcionales segun modulo: AWS S3, Resend, SNS,
  Cloudflare Turnstile.

## Configuracion

Crear `.env` desde `.env.example` y completar valores reales:

```bash
cp .env.example .env
```

Variables principales:

```env
PORT=4000
MONGODB_URI=mongodb://localhost:27017/electricautomaticchile
MONGODB_DATABASE=electricautomaticchile
MONGODB_MAX_POOL_SIZE=800
MONGODB_MIN_POOL_SIZE=20
JWT_SECRET=clave-larga
NODE_ENV=development
REDIS_HOST=localhost
REDIS_PORT=6379
CORS_ORIGINS=http://localhost:3000,http://localhost:8080
AUTH_COOKIE_DOMAIN=
TURNSTILE_SECRET_KEY=
```

Para pruebas de carga usar una base separada, por ejemplo:

```bash
MONGODB_DATABASE=electricautomaticchile_loadtest go run .
```

No ejecutar k6 contra la base productiva.

## Modo Alta Carga

El backend usa colas en memoria con escritura por lotes para absorber picos de
telemetria IoT y leads publicos sin bloquear el request HTTP en cada escritura
MongoDB.

Variables IoT:

```env
IOT_INGEST_ASYNC=true
IOT_INGEST_QUEUE_SIZE=100000
IOT_INGEST_BATCH_SIZE=1000
IOT_INGEST_WORKERS=4
IOT_INGEST_FLUSH_INTERVAL_MS=100
IOT_INGEST_WRITE_TIMEOUT_MS=15000
IOT_INGEST_MAX_RETRIES=3
IOT_TOKEN_CACHE_TTL_MS=300000
```

Variables leads:

```env
LEAD_INGEST_ASYNC=true
LEAD_INGEST_QUEUE_SIZE=50000
LEAD_INGEST_BATCH_SIZE=500
LEAD_INGEST_WORKERS=4
LEAD_INGEST_FLUSH_INTERVAL_MS=100
LEAD_INGEST_WRITE_TIMEOUT_MS=15000
LEAD_INGEST_MAX_RETRIES=3
```

`/health` expone `iot_ingestor` y `lead_ingestor` con `enqueued`, `flushed`,
`queued`, `dropped` y `failed`. En produccion esos valores deben estar en
monitoreo; `queued` debe volver a cero despues de picos y `dropped/failed`
deben mantenerse en cero.

Nota operativa: estas colas son locales al proceso. Para alta disponibilidad
multiinstancia y tolerancia a caidas, mover la cola a Redis Streams, Kafka,
NATS, SQS u otro broker durable.

## Desarrollo

```bash
go mod download
go run .
```

El healthcheck queda disponible en:

```text
GET http://localhost:4000/health
```

Comandos utiles:

```bash
go test ./...
go test ./load-tests/seed-devices
go build -o electric-backend main.go
```

## Produccion

- Usar `NODE_ENV=production` para activar Gin release mode.
- Redis es obligatorio en produccion; si no conecta, el backend termina.
- Configurar `CORS_ORIGINS` con dominios reales.
- Si frontend y API viven en subdominios, usar
  `AUTH_COOKIE_DOMAIN=.electricautomaticchile.com`.
- Mantener `JWT_SECRET` largo y rotado fuera del repositorio.
- Configurar backup y monitoreo de MongoDB antes de operar con clientes reales.
- Revisar logs, latencia p95, errores 5xx, conexiones a MongoDB, uso de Redis y
  metricas `*_ingestor`.

## Endpoints operativos

- `POST /api/auth/login` y `POST /api/auth/login/empresa`.
- `POST /api/leads` para formularios publicos de la landing.
- `GET /api/leads` y `PUT /api/leads/:id/status` para administracion.
- `POST /api/iot/lectura` para lecturas IoT autenticadas por API key.
- `GET /api/dispositivos`, `GET /api/clientes`, `GET /api/dashboard/*` para
  operacion del panel.

## Pruebas de carga

Las pruebas k6 estan en `load-tests/`. El flujo seguro es:

```bash
MONGODB_DATABASE=electricautomaticchile_loadtest DEVICE_COUNT=10000 go run ./load-tests/seed-devices
MONGODB_DATABASE=electricautomaticchile_loadtest go run .
BASE_URL=http://localhost:4000 k6 run load-tests/k6/landing-leads.js
```

Resultados y umbrales se documentan en `load-tests/README.md`.
