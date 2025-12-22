FROM golang:1.24-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go run github.com/a-h/templ/cmd/templ@latest generate ./view/
RUN CGO_ENABLED=1 go build -o cockpit main.go

FROM alpine:3.22

LABEL org.opencontainers.image.source=https://github.com/streambinder/vigor

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /home/appuser
COPY --from=builder /build/cockpit .
RUN chown -R appuser:appgroup /home/appuser
USER appuser

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

EXPOSE 8080
CMD ["./cockpit"]
