FROM golang:1.25-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

# -ldflags="-s -w" - удаляет отладочную информацию и таблицу символов (уменьшает размер на ~25%)
# -trimpath - убирает пути файловой системы из бинарника (для безопасности и воспроизводимости)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o url-shortener ./cmd/api

FROM alpine:latest

RUN apk add --no-cache ca-certificates curl bash \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=builder /app/url-shortener /app/url-shortener
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/db/migrations /app/db/migrations

RUN chown -R app:app /app
USER app

EXPOSE 3000