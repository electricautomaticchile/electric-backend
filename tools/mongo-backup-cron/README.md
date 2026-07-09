# Backup de MongoDB como Render Cron Job (alternativa off-AWS)

Backup one-shot de MongoDB a almacenamiento **S3-compatible** (Cloudflare R2 o AWS
S3), pensado para correr como **Render Cron Job**. Ventaja: el acceso a Mongo sale
de IPs de **Render**, así el allowlist de Atlas puede restringirse a Render y no
depende de la IP dinámica de la Lambda de AWS.

Módulo Go aislado (`go.mod` propio). Compilar con `GOWORK=off`.

## Por qué / tradeoffs
- ✅ Resuelve el allowlist de Atlas (todo Mongo pasa por Render).
- ✅ Con **Cloudflare R2** (10 GB gratis, S3-compatible) el backup queda 100% off-AWS.
- ⚠️ Los **Render Cron Jobs tienen un costo pequeño** (no son free como la Lambda de AWS, que es $0 en free tier).
- ⚠️ Requiere crear el bucket (R2 o S3) y cargar las credenciales en Render.

Mientras no se despliegue esto, la **Lambda de AWS sigue haciendo el backup diario** (no se rompe nada). Al activar el cron de Render, se deshabilita la Lambda + su schedule.

## Almacenamiento recomendado: Cloudflare R2
1. Cloudflare → R2 → crear bucket `electricautomaticchile-mongo-backups`.
2. Crear un API Token R2 (Access Key ID + Secret).
3. Endpoint: `https://<account_id>.r2.cloudflarestorage.com`.
4. Configurar lifecycle de 30 días en el bucket (retención).

## Desplegar en Render
Opción A (dashboard): New → Cron Job → conectar el repo → runtime Docker →
Dockerfile `tools/mongo-backup-cron/Dockerfile`, context `tools/mongo-backup-cron` →
schedule `0 7 * * *` (UTC) → cargar las env vars.

Opción B (blueprint): usar `render.yaml` de este directorio como referencia.

### Variables de entorno (en Render, NO en el repo)
| Var | Valor |
| --- | --- |
| `MONGODB_URI` | URI de Atlas (idealmente el usuario de mínimo privilegio rotado) |
| `MONGO_DB` | `electricautomaticchile` |
| `S3_BUCKET` | nombre del bucket |
| `S3_ENDPOINT` | endpoint R2 (o vacío para AWS S3) |
| `S3_REGION` | `auto` (R2) o `us-east-1` (S3) |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` | credenciales del storage (R2 o S3) |

## Migración desde la Lambda de AWS
1. Crear el bucket R2 + credenciales.
2. Crear el Cron Job en Render con las env vars y verificar una corrida (logs).
3. Confirmar objetos en R2.
4. Deshabilitar el schedule de EventBridge `mongo-backup-diario` y la Lambda `mongo-backup` (o dejarlas 1-2 días de solape).
5. Restringir el allowlist de Atlas a IPs de Render (requerimiento A2).

## Probar localmente
```bash
cd tools/mongo-backup-cron
GOWORK=off MONGODB_URI="..." S3_BUCKET="..." S3_ENDPOINT="https://<acct>.r2.cloudflarestorage.com" \
  S3_REGION=auto AWS_ACCESS_KEY_ID="..." AWS_SECRET_ACCESS_KEY="..." GOWORK=off go run .
```
