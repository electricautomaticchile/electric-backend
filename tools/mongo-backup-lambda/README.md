# Backup diario de MongoDB Atlas a S3 (AWS Lambda)

Lambda en Go que respalda todas las colecciones de MongoDB Atlas a S3, comprimidas
en gzip (JSON extendido). Se ejecuta una vez al día vía EventBridge Scheduler.

Módulo Go **aislado** (tiene su propio `go.mod`) para no afectar el build del
backend. El workspace raíz usa `go.work`, así que hay que compilar con `GOWORK=off`.

## Estado del despliegue (jul 2026 · us-east-1) — ACTIVO

Desplegado, **probado** (backup real de 16 colecciones / 1502 docs a S3) y
**automatizado**:

| Recurso | Nombre |
| --- | --- |
| Lambda | `mongo-backup` (provided.al2023, arm64, 120s, 256MB) |
| Bucket S3 | `electricautomaticchile-mongo-backups` (público bloqueado, lifecycle 30 días en `backups/`) |
| Schedule | `mongo-backup-diario` — `cron(0 7 * * ? *)` (07:00 UTC), ENABLED |
| Secreto | Secrets Manager `electricautomaticchile/mongodb_uri` |
| Rol ejecución | `mongo-backup-lambda-role` (logs + s3:PutObject + secretsmanager:GetSecretValue) |
| Rol scheduler | `mongo-backup-scheduler-role` (lambda:InvokeFunction) |
| Env de la Lambda | `S3_BUCKET`, `SECRET_ID` (ARN del secreto), `MONGO_DB=electricautomaticchile` |

Permisos: se creó el grupo IAM `backup-ops` con la política gestionada
`backup-scheduler-secrets` (scheduler + secrets + passrole) y se agregó
`electric-cli` al grupo (el usuario ya tenía sus 10 políticas al máximo).

## Seguridad del secreto

La `MONGODB_URI` vive en **AWS Secrets Manager**. La Lambda la lee vía la env
`SECRET_ID` con `secretsmanager:GetSecretValue`. No hay credenciales en claro en
variables de entorno ni en el código. Orden de resolución soportado por el
código: `SECRET_ID` → `MONGODB_URI` (env) → `SSM_PARAM_NAME`.

## Recompilar y redeployar el código

```bash
cd tools/mongo-backup-lambda
GOWORK=off GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap main.go
zip backup-lambda.zip bootstrap
aws s3 cp backup-lambda.zip s3://electricautomaticchile-mongo-backups/deploy/backup-lambda.zip
aws lambda update-function-code --function-name mongo-backup \
  --s3-bucket electricautomaticchile-mongo-backups --s3-key deploy/backup-lambda.zip
```

## Ejecutar un backup manual

```bash
aws lambda invoke --function-name mongo-backup --region us-east-1 /dev/stdout
```

## Restaurar un backup

```bash
# 1. Descargar el prefijo del día deseado
aws s3 cp --recursive s3://electricautomaticchile-mongo-backups/backups/2026-07-05/155426/ ./restore/
# 2. Por cada colección (JSON extendido, una línea por doc, gzip):
gunzip -k ./restore/clientes.json.gz
mongoimport --uri "<MONGODB_URI>" --db electricautomaticchile \
  --collection clientes --file ./restore/clientes.json
```

> Pendiente recomendado: probar la restauración a una BD de prueba al menos una vez.
