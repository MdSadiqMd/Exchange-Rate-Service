FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o exchange-rate-service ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/exchange-rate-service .
COPY --from=builder /app/config.yaml .

EXPOSE 8080
CMD ["./exchange-rate-service"]
