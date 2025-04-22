FROM golang:1.21.0 as preparer
# Стадия сборки
FROM golang:1.21.0 AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Устанавливаем Revel CLI
RUN go install github.com/revel/cmd/revel@latest

# Собираем приложение (если нужно, убери, если используешь revel run)
# RUN revel build -a . -t /app/target

# Финальная стадия
FROM golang:1.21.0

WORKDIR /app

# Устанавливаем Revel CLI
RUN go install github.com/revel/cmd/revel@latest

# Копируем код из builder
COPY --from=builder /app /app

EXPOSE 9000

# Запускаем приложение
CMD ["revel", "run", "-a", "."]