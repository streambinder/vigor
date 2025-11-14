FROM golang:1.25-alpine AS bootstrap-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY model/ ./model/
COPY main.go ./
RUN CGO_ENABLED=0 go build -o bootstrap main.go

FROM pgvector/pgvector:0.8.1-pg18 AS data-builder
USER postgres
WORKDIR /build
COPY --from=bootstrap-builder /build/bootstrap /usr/local/bin/bootstrap
COPY features features
RUN initdb -D /tmp/pgdata && \
  pg_ctl -D /tmp/pgdata -o "-k /tmp" start && \
  pg_isready -t 30 -h /tmp && \
  psql -h /tmp -U postgres -c "CREATE DATABASE db;" && \
  psql -h /tmp -U postgres -c "CREATE USER \"user\" WITH PASSWORD 'password';" && \
  psql -h /tmp -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE db TO \"user\";" && \
  psql -h /tmp -U postgres -d db -c "CREATE EXTENSION IF NOT EXISTS vector;" && \
  psql -h /tmp -U postgres -d db -c "GRANT ALL ON SCHEMA public TO \"user\";" && \
  sh -c 'DATABASE_URL="postgres://user:password@/db?host=/tmp" bootstrap' && \
  pg_dump -h /tmp -U postgres -d db --clean --if-exists >/build/init.sql && \
  pg_ctl -D /tmp/pgdata stop

FROM pgvector/pgvector:pg16
ENV POSTGRES_DB=db
ENV POSTGRES_USER=user
COPY --from=data-builder /build/init.sql /docker-entrypoint-initdb.d/
RUN echo "CREATE EXTENSION IF NOT EXISTS vector;" > /docker-entrypoint-initdb.d/00-init-pgvector.sql
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD pg_isready -U user -d db || exit 1
EXPOSE 5432
USER postgres
