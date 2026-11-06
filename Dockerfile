# Multi-stage production build for the Treewiki API gateway (Go).
# Build:   docker build -t treewiki-gateway -f Dockerfile .
# Run:     docker run --rm -p 8080:8080 --env-file .env treewiki-gateway
FROM golang:1.27-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY pkg/ pkg/
COPY services/ services/
COPY database/ database/
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/treewiki-gateway ./services/gateway/cmd

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata wget \
  && adduser -D -H -u 10001 treewiki
WORKDIR /app
COPY --from=builder /out/treewiki-gateway /app/treewiki-gateway
COPY --from=builder /src/database/migrations /app/database/migrations
RUN mkdir -p /app/uploads && chown -R treewiki:treewiki /app
USER treewiki
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/app/treewiki-gateway"]
