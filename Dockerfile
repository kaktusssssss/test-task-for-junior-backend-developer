FROM golang:1.23-alpine AS builder

# Используем зеркало USTC для загрузки зависимостей Go
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /out/taskservice ./cmd/api

# --- Финальный образ ---
FROM alpine:3.20

# Используем то же зеркало USTC для установки ca-certificates
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/taskservice /app/taskservice

EXPOSE 8080

CMD ["/app/taskservice"]