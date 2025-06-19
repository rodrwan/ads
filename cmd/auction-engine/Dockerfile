FROM golang:1.24-alpine

WORKDIR /app

# Instalar dependencias necesarias
RUN apk add --no-cache git make

# Instalar CompileDaemon para hot-reload
RUN go install github.com/githubnemo/CompileDaemon@latest

# Instalar sqlc (última versión compatible con Go 1.24)
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest


# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Generar código SQL
RUN make generate


# Exponer el puerto
EXPOSE 8080

# Comando por defecto para desarrollo
ENTRYPOINT CompileDaemon -build="go build -o /app/auction-engine ./cmd/auction-engine/main.go" \
    -command="/app/auction-engine"
