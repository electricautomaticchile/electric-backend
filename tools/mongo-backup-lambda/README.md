# Backup de MongoDB Atlas a S3 (AWS Lambda)

Lambda en Go que respalda todas las colecciones de MongoDB Atlas a S3 (gzip, JSON
extendido). El mismo binario sirve dos modos según la env `MODE`:
- **backup** (default): dump de todas las colecciones a S3. Corre diario.
- **report**: resumen semanal de los backups de los últimos 7 días, enviado por
  email vía SES.

Módulo Go **aislado** (`go.mod` propio) para no afectar el build del backend. El
workspace raíz usa `go.work`, así que hay que compilar con `GOWORK=off`.

## Estado del despliegue (jul 2026 · us-east-1) — ACTIVO

| Recurso | Nombre / valor |
| --- | --- |
| Lambda backup | `mongo-backup` (provided.al2023, arm64) |
| Lambda reporte | `mongo-backup-report` (MODE=report) |
| Bucket S3 | `electricautomaticchile-mongo-backups` (público bloqueado, lifecycle 30 días) |
| Schedule backup | `mongo-backup-diario` — `cron(0 7 * * ? *)` (07:00 UTC) |
| Schedule reporte | `mongo-backup-report-semanal` — `cron(0 8 ? * MON *)` (lunes 08:00 UTC) |
| Secreto | Secrets Manager `electricautomaticchile/mongodb_uri` |
| Alertas de fallo | CloudWatch alarmas `mongo-backup-errores` y `mongo-backup-report-errores` → SNS `mongo-backup-alerts` → email |
| Rol Lambdas | `mongo-backup-lambda-role` (logs, s3:Put/ListBucket, secretsmanager:GetSecretValue, ses:SendEmail) |
| Rol scheduler | `mongo-backup-scheduler-role` (invoca `mongo-backup*`) |

Permisos del deploy: grupo IAM `backup-ops` con la política `backup-scheduler-secrets`.

## Notificaciones por email

- **Fallos (esencial):** las alarmas de CloudWatch sobre la métrica `Errors` de
  cada Lambda publican en el topic SNS `mongo-backup-alerts`, suscrito a
  `pipeaalzamora@gmail.com`. **La suscripción SNS requiere confirmación**: hay que
  hacer clic en el enlace del correo "AWS Notification - Subscription Confirmation".
- **Resumen semanal:** la Lambda `mongo-backup-report` envía cada lunes un correo
  vía SES con nº de backups, último backup y tamaño total de la semana.
- No se envía correo por cada backup exitoso (evita fatiga de alertas).

## Variables de entorno

- `mongo-backup`: `S3_BUCKET`, `SECRET_ID` (ARN del secreto), `MONGO_DB`.
- `mongo-backup-report`: `MODE=report`, `S3_BUCKET`, `REPORT_TO`, `REPORT_FROM`.

## Recompilar / redeployar

```bash
cd tools/mongo-backup-lambda
GOWORK=off GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap main.go
zip backup-lambda.zip bootstrap
aws s3 cp backup-lambda.zip s3://electricautomaticchile-mongo-backups/deploy/backup-lambda.zip
aws lambda update-function-code --function-name mongo-backup       --s3-bucket electricautomaticchile-mongo-backups --s3-key deploy/backup-lambda.zip
aws lambda update-function-code --function-name mongo-backup-report --s3-bucket electricautomaticchile-mongo-backups --s3-key deploy/backup-lambda.zip
```

## Backup manual / restauración

```bash
aws lambda invoke --function-name mongo-backup --region us-east-1 /dev/stdout
# Restaurar:
aws s3 cp --recursive s3://electricautomaticchile-mongo-backups/backups/<YYYY-MM-DD>/<HHMMSS>/ ./restore/
gunzip -k ./restore/clientes.json.gz
mongoimport --uri "<MONGODB_URI>" --db electricautomaticchile --collection clientes --file ./restore/clientes.json
```

> Restauración ya verificada (jul 2026): 3/3 docs de `clientes` restaurados en una
> BD de prueba con `_id` reconstruido como ObjectID. Fidelidad de tipos OK.
