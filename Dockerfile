# ==============================================================================
# Dockerfile Multi-Stage para Nodo IPv7
# Imagen ultra-ligera (< 25 MB) basada en Alpine Linux para producción
# ==============================================================================

# STAGE 1: Builder
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Instalar dependencias del sistema mínimas
RUN apk add --no-cache git ca-certificates

# Cachear módulos Go
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente completo
COPY . .

# Compilar binario estático optimizado para producción
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/ipv7-node ./cmd/node

# STAGE 2: Runtime Minimalista
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

# Crear usuario sin privilegios por seguridad
RUN addgroup -S ipv7 && adduser -S -G ipv7 ipv7
USER ipv7
WORKDIR /home/ipv7

# Copiar binario compilado desde el builder
COPY --from=builder /bin/ipv7-node /usr/local/bin/ipv7-node

# Puertos expuestos:
# 7001/udp - Transporte P2P Malla (QUIC / UDP)
# 7001/tcp - Transporte Fallback Relay
# 8080/tcp - Dashboard Web, OpenAPI, Prometheus & MCP
EXPOSE 7001/udp 7001/tcp 8080/tcp

# Punto de entrada predeterminado
ENTRYPOINT ["/usr/local/bin/ipv7-node"]
CMD ["-port", "7001", "-ui", "8080", "-open=false"]
