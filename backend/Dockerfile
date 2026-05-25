FROM golang:1.26.3-alpine3.22 AS builder
WORKDIR /app
COPY knowledge/ ./knowledge/
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /app/backend
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o backend main.go

FROM alpine:3.23.4

LABEL org.opencontainers.image.source=https://github.com/streambinder/vigor

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Install wget for healthcheck
RUN apk add --no-cache wget=1.25.0-r1

WORKDIR /home/appuser
COPY --from=builder /app/backend/backend .

# Change ownership
RUN chown -R appuser:appgroup /home/appuser

# Switch to non-root user
USER appuser

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8000/ || exit 1

EXPOSE 8000
CMD ["./backend"]
