FROM golang:1.25.6-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o auth-service ./cmd/main.go

# STAGE 2
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/auth-service .

CMD ["./auth-service"]