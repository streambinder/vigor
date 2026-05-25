# docker build -t ghcr.io/streambinder/vigor/knowledge:latest -f knowledge/Dockerfile \
#   --secret id=openrouter_api_key,env=OPENROUTER_API_KEY \
#   --build-arg OPENROUTER_EMBEDDING_MODELS=google/gemini-embedding-001 .

FROM golang:1.26.1-alpine3.22 AS bootstrap-builder
RUN apk add --no-cache gcc=14.2.0-r6 musl-dev=1.2.5-r12
WORKDIR /build/knowledge
COPY knowledge/go.mod knowledge/go.sum ./
COPY backend/go.mod backend/go.sum ../backend/
RUN go mod download
COPY backend ../backend/
COPY knowledge/main.go ./
RUN CGO_ENABLED=0 go build -o bootstrap main.go

FROM pgvector/pgvector:0.8.2-pg18 AS data-builder
ARG OPENROUTER_EMBEDDING_MODELS
USER root
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates=20230311+deb12u1 && rm -rf /var/lib/apt/lists/*
USER postgres
WORKDIR /build
COPY --from=bootstrap-builder /build/knowledge/bootstrap /usr/local/bin/bootstrap
COPY knowledge/features features
RUN --mount=type=secret,id=openrouter_api_key,uid=999 initdb -D /tmp/pgdata && \
  pg_ctl -D /tmp/pgdata -o "-k /tmp" start && \
  pg_isready -t 30 -h /tmp && \
  psql -h /tmp -U postgres -c "CREATE DATABASE db;" && \
  psql -h /tmp -U postgres -c "CREATE USER \"user\" WITH PASSWORD 'password';" && \
  psql -h /tmp -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE db TO \"user\";" && \
  psql -h /tmp -U postgres -d db -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" && \
  psql -h /tmp -U postgres -d db -c "CREATE EXTENSION IF NOT EXISTS vector;" && \
  psql -h /tmp -U postgres -d db -c "GRANT ALL ON SCHEMA public TO \"user\";" && \
  # checkov:skip=CKV_SECRET_4:Temporary build-time credentials for local database bootstrap
  sh -c "DATABASE_URL='postgres://user:password@/db?host=/tmp' OPENROUTER_API_KEY=$(cat /run/secrets/openrouter_api_key) OPENROUTER_EMBEDDING_MODELS='$OPENROUTER_EMBEDDING_MODELS' bootstrap" && \
  pg_dump -h /tmp -U postgres -d db --clean --if-exists >/build/init.sql && \
  pg_ctl -D /tmp/pgdata stop

FROM pgvector/pgvector:0.8.2-pg18

LABEL org.opencontainers.image.source=https://github.com/streambinder/vigor

ENV POSTGRES_DB=db
ENV POSTGRES_USER=user
COPY --from=data-builder /build/init.sql /docker-entrypoint-initdb.d/
RUN echo "CREATE EXTENSION IF NOT EXISTS vector;" > /docker-entrypoint-initdb.d/00-init-pgvector.sql
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD pg_isready -U user -d db || exit 1
EXPOSE 5432
USER postgres
