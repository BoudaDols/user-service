FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./cmd/main.go

# ── Runtime ──────────────────────────────────────────────────
FROM alpine:3.19

RUN apk upgrade --no-cache

WORKDIR /app

COPY --from=builder /app/user-service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./user-service"]
