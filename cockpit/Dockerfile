FROM golang:1.24-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /build
COPY backend/go.mod backend/go.sum ./backend/
COPY cockpit/go.mod cockpit/go.sum ./cockpit/
RUN cd cockpit && go mod download
COPY backend/ ./backend/
COPY cockpit/ ./cockpit/
RUN cd cockpit && go run github.com/a-h/templ/cmd/templ@latest generate ./view/
RUN cd cockpit && CGO_ENABLED=1 go build -o cockpit main.go

FROM alpine:3.22

LABEL org.opencontainers.image.source=https://github.com/streambinder/vigor

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /home/appuser
COPY --from=builder /build/cockpit/cockpit .
COPY --from=builder /build/cockpit/static ./static
RUN chown -R appuser:appgroup /home/appuser
USER appuser

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

EXPOSE 8080
CMD ["./cockpit"]
