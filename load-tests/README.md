# Load tests

Pruebas k6 para validar capacidad con miles de clientes/dispositivos contra un
ambiente de staging lo mas parecido posible a produccion.

## Requisitos

- `k6` instalado en la maquina que ejecuta la prueba.
- `BASE_URL` apuntando al backend, por ejemplo `https://api.electricautomaticchile.com`.
- MongoDB y Redis reales, no mocks locales.
- Para pruebas de leads con Turnstile, usar un ambiente de staging sin
  `TURNSTILE_SECRET_KEY` o credenciales de prueba de Cloudflare. Los tokens reales
  son de un solo uso y no sirven para carga sintetica masiva.

## Leads publicos

```bash
BASE_URL=http://localhost:4000 k6 run load-tests/k6/landing-leads.js
```

Variables utiles:

- `START_RATE`: tasa inicial por segundo. Default `10`.
- `PEAK_RATE`: tasa maxima por segundo. Default `1000`.
- `PRE_ALLOCATED_VUS`: VUs preasignados. Default `300`.
- `MAX_VUS`: limite de VUs. Default `2000`.
- `RAMP_25_DURATION`, `RAMP_50_DURATION`, `RAMP_PEAK_DURATION`,
  `HOLD_DURATION`, `RAMP_DOWN_DURATION`: duraciones de cada etapa.
- `IP_SPREAD=false`: desactiva `X-Forwarded-For` aleatorio.
- `TURNSTILE_TOKEN`: token fijo solo para ambientes de prueba.

## Lecturas IoT

Antes de ejecutar, crear o cargar dispositivos existentes en MongoDB. Para un
ambiente de carga aislado puedes generar dispositivos sinteticos:

```bash
MONGODB_DATABASE=electricautomaticchile_loadtest \
DEVICE_COUNT=10000 \
go run ./load-tests/seed-devices
```

El script k6 puede recibir IDs por coma:

```bash
BASE_URL=http://localhost:4000 \
IOT_API_KEY=secret \
DEVICE_IDS=MED-000001,MED-000002,MED-000003 \
k6 run load-tests/k6/iot-readings.js
```

Tambien acepta `START_RATE`, `PEAK_RATE`, `PRE_ALLOCATED_VUS`, `MAX_VUS`,
`RAMP_25_DURATION`, `RAMP_50_DURATION`, `RAMP_PEAK_DURATION`, `HOLD_DURATION`,
`RAMP_DOWN_DURATION` e `IP_SPREAD`.

## Interpretacion minima

No declarar que "soporta miles" solo por pasar build. Para una prueba aceptable:

- `http_req_failed < 1%`.
- `p95 < 500ms` para inserts de leads.
- `p95 < 250ms` para lecturas IoT.
- CPU, memoria, conexiones MongoDB y Redis sin saturacion sostenida.
- Sin 429 excepto si se esta probando explicitamente rate limiting.

## Resultado local 2026-06-04

Ambiente usado:

- Backend Go ejecutado localmente en `:4000`.
- MongoDB Atlas con base aislada `electricautomaticchile_loadtest`.
- 10.000 dispositivos sinteticos creados con `load-tests/seed-devices`.
- Redis local no disponible durante la prueba; en produccion Redis debe estar
  activo.
- Pruebas cortas de descubrimiento, no soak test de varias horas.

Resumen:

| Escenario | Resultado | Lecturas/requests | p95 | Errores HTTP | Nota |
| --- | --- | ---: | ---: | ---: | --- |
| Leads smoke, peak 5/s | Pasa | 97 | 59.57ms | 0% | Sanidad basica |
| Leads peak 50/s | Pasa | 2.117 | 57.93ms | 0% | Estable |
| Leads peak 60/s | Pasa | 2.524 | 297.44ms | 0% | Techo conservador probado |
| Leads peak 75/s | Falla latencia | 3.152 | 650.27ms | 0% | Supera p95<500ms |
| Leads peak 100/s | Falla latencia/capacidad | 3.559 | 4.23s | 0% | 615 iteraciones descartadas |
| IoT peak 25/s | Pasa | 1.060 | 111.2ms | 0% | Estable |
| IoT peak 100/s | Falla latencia/capacidad | 3.342 | 4.72s | 0% | 832 iteraciones descartadas |
| IoT peak 500/s | Falla latencia/capacidad | 7.170 | 7.67s | 0% | 13.954 iteraciones descartadas |

Conclusion operativa: con esta infraestructura no se puede afirmar que soporta
miles de clientes o dispositivos enviando eventos concurrentes. Para leads, el
techo local probado y aceptable queda alrededor de 60 solicitudes/s. Para IoT,
25 lecturas/s pasa, mientras que 100/s y 500/s ya acumulan latencia.

Archivos de salida locales:

- `load-tests/results/landing-leads-smoke.json`
- `load-tests/results/landing-leads-peak50.json`
- `load-tests/results/landing-leads-peak60.json`
- `load-tests/results/landing-leads-peak75.json`
- `load-tests/results/landing-leads-peak100.json`
- `load-tests/results/iot-readings-peak25.json`
- `load-tests/results/iot-readings-peak100.json`
- `load-tests/results/iot-readings-peak500.json`

Antes de subir el objetivo a miles de clientes se requiere repetir estas
pruebas con Redis activo, `NODE_ENV=production`, limites de MongoDB monitoreados,
metricas de CPU/memoria y una prueba sostenida de al menos 30-60 minutos.

## Resultado posterior a optimizacion 2026-06-05

Cambios activos:

- `POST /api/iot/lectura` responde despues de encolar y escribe por lotes en
  MongoDB.
- `POST /api/leads` responde despues de encolar y escribe por lotes en MongoDB.
- Auditoria, compresion y access log se omiten para endpoints de ingesta
  masiva.
- API key IoT global evita roundtrip MongoDB; tokens por dispositivo quedan en
  cache temporal.
- Pool MongoDB y colas quedan configurables por variables de entorno.

Ambiente usado:

- Backend local en `:4000`, `NODE_ENV=development`.
- MongoDB Atlas con base `electricautomaticchile_loadtest`.
- Redis local no disponible.
- 10.000 dispositivos sinteticos.
- Configuracion de prueba:
  `IOT_INGEST_WORKERS=8`, `IOT_INGEST_BATCH_SIZE=1000`,
  `IOT_INGEST_FLUSH_INTERVAL_MS=50`, `IOT_INGEST_MAX_RETRIES=5`,
  `LEAD_INGEST_WORKERS=4`, `LEAD_INGEST_BATCH_SIZE=500`,
  `LEAD_INGEST_FLUSH_INTERVAL_MS=50`, `LEAD_INGEST_MAX_RETRIES=5`,
  `MONGODB_MAX_POOL_SIZE=1000`.

| Escenario | Resultado | Requests | p95 HTTP | Errores HTTP | Persistencia |
| --- | --- | ---: | ---: | ---: | --- |
| IoT peak 500/s | Pasa | 21.124 | 143.19us | 0% | Sin fallos observados |
| IoT peak 1000/s | Pasa | 42.249 | 142.37us | 0% | `flushed=42249`, `failed=0`, `queued=0` |
| IoT peak 3000/s | Pasa como pico | 126.249 | 163.25us | 0% | Cola dreno a `queued=0`, `failed=0` despues del pico |
| Leads peak 100/s | Pasa | 4.174 | 511.11us | 0% | `flushed=4174`, `failed=0`, `queued=0` |
| Leads peak 500/s | Pasa | 21.124 | 353.11us | 0% | `flushed=25298` acumulado, `failed=0`, `queued=0` |

Lectura operativa:

- La API ya absorbe miles de lecturas IoT por segundo en una instancia local.
- A 3000/s MongoDB no fue el cuello de botella del request, pero si necesito
  drenar cola despues del pico; por eso debe monitorearse `queued`.
- La landing queda con margen probado a 500 leads/s sin bloquear al usuario.
- Estos resultados validan picos cortos; falta un soak test de 30-60 minutos
  con Redis activo, `NODE_ENV=production`, replicas reales y metricas de
  infraestructura antes de prometer SLA productivo.

Nuevos artefactos locales:

- `load-tests/results/iot-readings-peak500-after-async.json`
- `load-tests/results/iot-readings-peak1000-after-retry.json`
- `load-tests/results/iot-readings-peak3000-after-retry.json`
- `load-tests/results/landing-leads-peak100-after-async.json`
- `load-tests/results/landing-leads-peak500-after-async.json`
