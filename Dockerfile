FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o tictactoe ./cmd/main.go
RUN go build -o tgbot ./cmd/bot/main.go
RUN ls -la /app

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/tictactoe .
COPY --from=builder /app/tgbot .
COPY .env .
CMD ["./tictactoe"]