# Multi-stage production build for Treewiki
FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /treewiki-gateway ./services/gateway/cmd

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /treewiki-gateway /app/treewiki-gateway
COPY infrastructure/nginx.conf /etc/nginx/nginx.conf 2>/dev/null || true
EXPOSE 8080
ENTRYPOINT ["/app/treewiki-gateway"]