FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код (включая migrations)
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/app/main.go

# Финальный минимальный образ
FROM alpine:latest

WORKDIR /app

# Копируем собранное приложение
COPY --from=builder /app/main .

# Копируем .env
COPY .env ./

# ←←← КОПИРУЕМ ПАПКУ С МИГРАЦИЯМИ
COPY migrations ./migrations

EXPOSE 8080

CMD ["./main"]