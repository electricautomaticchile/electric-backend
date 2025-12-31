# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Instalar dependencias del sistema
RUN apk add --no-cache git

# Copiar go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar binario desde builder
COPY --from=builder /app/main .

EXPOSE 4000

CMD ["./main"]
