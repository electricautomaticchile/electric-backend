# Imagen para el Render Cron Job de backup de MongoDB.
# Build context = este directorio (tools/mongo-backup-cron).
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/mongo-backup-cron .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/mongo-backup-cron /usr/local/bin/mongo-backup-cron
# Render ejecuta el contenedor una vez por cada corrida del cron.
ENTRYPOINT ["/usr/local/bin/mongo-backup-cron"]
