FROM golang:1.25.0-alpine3.22 AS builder
WORKDIR /app
COPY knowledge/ ./knowledge/
COPY server/go.mod server/go.sum ./server/
WORKDIR /app/server
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

FROM alpine:3.22

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Install wget for healthcheck
RUN apk add --no-cache wget

WORKDIR /home/appuser
COPY --from=builder /app/server/server .

# Change ownership
RUN chown -R appuser:appgroup /home/appuser

# Switch to non-root user
USER appuser

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

EXPOSE 8080
CMD ["./server"]
