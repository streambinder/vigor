FROM golang:1.26.1-alpine3.22 AS builder
RUN apk add --no-cache gcc=14.2.0-r6 musl-dev=1.2.5-r12
WORKDIR /build
COPY backend/go.mod backend/go.sum ./backend/
COPY cockpit/go.mod cockpit/go.sum ./cockpit/
WORKDIR /build/cockpit
RUN go mod download
COPY backend/ /build/backend/
COPY cockpit/ .
# hadolint ignore=DL3062
RUN go run github.com/a-h/templ/cmd/templ@v0.3.977 generate ./view/ \
    && CGO_ENABLED=1 go build -o cockpit main.go

FROM alpine:3.22

LABEL org.opencontainers.image.source=https://github.com/streambinder/vigor

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /home/appuser
COPY --from=builder /build/cockpit/cockpit .
COPY --from=builder /build/cockpit/static/ ./static/
RUN chown -R appuser:appgroup /home/appuser
USER appuser

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

EXPOSE 8080
CMD ["./cockpit"]
