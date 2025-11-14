FROM golang:1.25-alpine AS bootstrap-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY model/ ./model/
COPY main.go ./
RUN CGO_ENABLED=1 go build -o bootstrap main.go

FROM postgres:16-alpine AS data-builder
RUN apk add --no-cache curl libgcc file libc6-compat
WORKDIR /build
COPY --from=bootstrap-builder /build/bootstrap /usr/local/bin/bootstrap
COPY features features
RUN su-exec postgres initdb -D /tmp/pgdata && \
    su-exec postgres pg_ctl -D /tmp/pgdata -o "-k /tmp" -l /tmp/postgres.log start && \
    su-exec postgres pg_isready -t 30 -h /tmp && \
    su-exec postgres psql -h /tmp -U postgres -c "CREATE DATABASE db;" && \
    su-exec postgres psql -h /tmp -U postgres -c "CREATE USER \"user\" WITH PASSWORD 'password';" && \
    su-exec postgres psql -h /tmp -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE db TO \"user\";" && \
    su-exec postgres psql -h /tmp -U postgres -d db -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" && \
    su-exec postgres psql -h /tmp -U postgres -d db -c "GRANT ALL ON SCHEMA public TO \"user\";" && \
    su-exec postgres sh -c 'DATABASE_URL="postgres://user:password@/db?host=/tmp" bootstrap /build/features' && \
    su-exec postgres pg_dump -h /tmp -U postgres -d db --clean --if-exists >/build/init.sql && \
    su-exec postgres pg_ctl -D /tmp/pgdata stop

FROM postgres:16-alpine
ENV POSTGRES_DB=db
ENV POSTGRES_USER=user
COPY --from=data-builder /build/init.sql /docker-entrypoint-initdb.d/
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD pg_isready -U user -d db || exit 1
EXPOSE 5432
USER postgres
