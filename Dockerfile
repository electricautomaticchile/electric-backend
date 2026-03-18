FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o electric-backend-server .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/electric-backend-server .
COPY --from=builder /app/infrastructure/email/templates ./infrastructure/email/templates

EXPOSE 4000
CMD ["./electric-backend-server"]
