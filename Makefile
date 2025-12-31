.PHONY: run build test clean install dev

# Variables
BINARY_NAME=electric-backend
MAIN_PATH=main.go

# Instalar dependencias
install:
	go mod download
	go mod tidy

# Ejecutar en desarrollo
dev:
	go run $(MAIN_PATH)

# Compilar
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Ejecutar binario compilado
run: build
	./$(BINARY_NAME)

# Tests
test:
	go test -v ./...

# Tests con coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Limpiar archivos generados
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Formatear código
fmt:
	go fmt ./...

# Linter
lint:
	golangci-lint run

# Hot reload (requiere air: go install github.com/cosmtrek/air@latest)
watch:
	air
