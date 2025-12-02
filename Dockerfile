FROM golang:1.25-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git build-base

COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Сборка main.go в src/cmd/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/app ./src/cmd

FROM alpine:latest

WORKDIR /app
RUN apk add --no-cache ca-certificates

# Копируем бинарник
COPY --from=builder /app/app .

EXPOSE 8080
CMD ["./app"]
