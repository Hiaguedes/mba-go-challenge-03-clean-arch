FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/ordersystem ./cmd/ordersystem

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/ordersystem ./ordersystem
COPY --from=builder /app/cmd/ordersystem/.env ./cmd/ordersystem/.env

EXPOSE 8000 50051 8080

ENTRYPOINT ["./ordersystem"]
