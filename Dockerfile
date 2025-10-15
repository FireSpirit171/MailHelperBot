FROM golang:1.23-alpine

# Устанавливаем нужные пакеты
RUN apk add --no-cache git bash netcat-openbsd

# Рабочая директория
WORKDIR /app

# Копируем модули и ставим зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем бинарник, чтобы не запускать go run в runtime
RUN go build -o main ./cmd/main.go

# Команда запуска
CMD ["./main"]
