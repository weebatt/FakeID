FROM golang:1.23.8-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /task-service ./cmd/main.go

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /task-service .
COPY --from=builder /app/config/.env.example ./config/.env
COPY --from=builder /app/migrations ./migrations

RUN apk --no-cache add ca-certificates

EXPOSE 8080

CMD ["./task-service"]