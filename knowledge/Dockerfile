# docker build -t ghcr.io/streambinder/vigor/knowledge:latest -f knowledge/Dockerfile --network host .

FROM golang:1.25-alpine AS bootstrap-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /build/knowledge
COPY knowledge/go.mod knowledge/go.sum ./
COPY server/go.mod server/go.sum ../server/
RUN go mod download
COPY server ../server/
COPY knowledge/main.go ./
RUN CGO_ENABLED=0 go build -o bootstrap main.go

FROM pgvector/pgvector:0.8.1-pg18 AS data-builder
USER postgres
WORKDIR /build
COPY --from=bootstrap-builder /build/knowledge/bootstrap /usr/local/bin/bootstrap
COPY knowledge/features features
RUN initdb -D /tmp/pgdata && \
  pg_ctl -D /tmp/pgdata -o "-k /tmp" start && \
  pg_isready -t 30 -h /tmp && \
  psql -h /tmp -U postgres -c "CREATE DATABASE db;" && \
  psql -h /tmp -U postgres -c "CREATE USER \"user\" WITH PASSWORD 'password';" && \
  psql -h /tmp -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE db TO \"user\";" && \
  psql -h /tmp -U postgres -d db -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" && \
  psql -h /tmp -U postgres -d db -c "CREATE EXTENSION IF NOT EXISTS vector;" && \
  psql -h /tmp -U postgres -d db -c "GRANT ALL ON SCHEMA public TO \"user\";" && \
  sh -c "DATABASE_URL='postgres://user:password@/db?host=/tmp' LLAMACPP_EMBEDDING_TIERS=http://host.docker.internal:8001 bootstrap" && \
  pg_dump -h /tmp -U postgres -d db --clean --if-exists >/build/init.sql && \
  pg_ctl -D /tmp/pgdata stop

FROM pgvector/pgvector:0.8.1-pg18
ENV POSTGRES_DB=db
ENV POSTGRES_USER=user
COPY --from=data-builder /build/init.sql /docker-entrypoint-initdb.d/
RUN echo "CREATE EXTENSION IF NOT EXISTS vector;" > /docker-entrypoint-initdb.d/00-init-pgvector.sql
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD pg_isready -U user -d db || exit 1
EXPOSE 5432
USER postgres
