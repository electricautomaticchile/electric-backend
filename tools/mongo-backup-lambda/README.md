# Backup diario de MongoDB Atlas a S3 (AWS Lambda)

Lambda en Go que respalda todas las colecciones de MongoDB Atlas a S3, comprimidas
en gzip (JSON extendido). Pensada para ejecutarse una vez al día vía EventBridge.

Módulo Go **aislado** (tiene su propio `go.mod`) para no afectar el build del
backend. El workspace raíz usa `go.work`, así que hay que compilar con `GOWORK=off`.

## Estado del despliegue (jul 2026 · us-east-1)

Ya desplegado y **probado** (backup real de 16 colecciones / 1502 docs a S3):

| Recurso | Nombre |
| --- | --- |
| Lambda | `mongo-backup` (provided.al2023, arm64, 120s, 256MB) |
| Bucket S3 | `electricautomaticchile-mongo-backups` (público bloqueado, lifecycle 30 días en `backups/`) |
| Rol de ejecución | `mongo-backup-lambda-role` (logs + s3:PutObject + ssm:GetParameter + kms:Decrypt) |
| Rol para el scheduler | `mongo-backup-scheduler-role` (lambda:InvokeFunction) — listo para usar |
| Variables de entorno | `S3_BUCKET`, `MONGODB_URI`, `MONGO_DB=electricautomaticchile` |

### ⏳ Pendiente: el disparador diario (permiso faltante)

La API key usada (`electric-cli`) **no tiene** permisos de `events:*` ni
`scheduler:*`, así que el cron no se pudo crear por código. Falta UN paso manual:

**Opción A — Consola (1 min):** EventBridge → Schedules → Create schedule →
cron `0 7 * * ? *` (07:00 UTC), Target = Lambda `mongo-backup`, usar el rol
`mongo-backup-scheduler-role`.

**Opción B — dar permisos y terminar por código:** agregar a `electric-cli`
`scheduler:CreateSchedule` + `iam:PassRole` (sobre `mongo-backup-scheduler-role`)
y avisar para crear el schedule automáticamente.

## Seguridad del secreto

Hoy la `MONGODB_URI` va como variable de entorno de la Lambda (cifrada en reposo
por KMS de AWS, visible para quien tenga acceso de lectura a la config de la
Lambda). Cuando `electric-cli` tenga permisos de SSM/Secrets Manager, migrar el
secreto ahí: el código ya lo soporta vía `SSM_PARAM_NAME` (borrar entonces la env
`MONGODB_URI`).

## Recompilar y redeployar el código

```bash
cd tools/mongo-backup-lambda
GOWORK=off GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap main.go
zip backup-lambda.zip bootstrap
# subir el zip a s3://electricautomaticchile-mongo-backups/deploy/backup-lambda.zip
aws lambda update-function-code --function-name mongo-backup \
  --s3-bucket electricautomaticchile-mongo-backups --s3-key deploy/backup-lambda.zip
```

## Restaurar un backup

```bash
# 1. Descargar el prefijo del día deseado
aws s3 cp --recursive s3://electricautomaticchile-mongo-backups/backups/2026-07-05/154618/ ./restore/
# 2. Por cada coleccion (JSON extendido, una linea por doc, gzip):
gunzip -k ./restore/clientes.json.gz
mongoimport --uri "<MONGODB_URI>" --db electricautomaticchile \
  --collection clientes --file ./restore/clientes.json
```

> Probar la restauración a una BD de prueba al menos una vez (criterio de aceptación).
