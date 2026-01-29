# =========================
# Build stage
# =========================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Dependências
COPY go.mod go.sum ./
RUN go mod download

# Código
COPY . .

# Build do binário
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o itau-bff ./cmd/api

# =========================
# Runtime stage
# =========================
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/itau-bff /app/itau-bff

EXPOSE 8080

ENTRYPOINT ["/app/itau-bff"]
