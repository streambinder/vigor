# Stage 1: Build the bootstrap tool
FROM golang:1.25-alpine AS bootstrap-builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# Copy go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY model/ ./model/
COPY bootstrap/ ./bootstrap/

# Build the bootstrap binary with CGO enabled for PostgreSQL driver
# Note: Using dynamic linking since gorm/pgx requires it
WORKDIR /build/bootstrap
# hadolint ignore=DL4006
RUN CGO_ENABLED=1 go build -o /tmp/bootstrap . && \
    chmod +x /tmp/bootstrap && \
    file /tmp/bootstrap && \
    echo "Testing binary..." && \
    /tmp/bootstrap 2>&1 | head -5 || echo "Binary requires DATABASE_URL"

# Stage 2: Download ExerciseDB data and populate database
FROM postgres:16-alpine AS data-builder

# Install required tools and runtime libraries for Go binary
RUN apk add --no-cache curl libgcc file libc6-compat

# Set working directory
WORKDIR /build

# Download and extract the latest ExerciseDB API data
RUN curl -L https://github.com/ExerciseDB/exercisedb-api/archive/refs/heads/main.tar.gz -o exercisedb-api.tar.gz && \
    tar -xzf exercisedb-api.tar.gz && \
    mv exercisedb-api-main/src/data ./data && \
    rm -rf exercisedb-api-main exercisedb-api.tar.gz && \
    chmod 777 /build

# Copy the bootstrap tool from bootstrap-builder (already executable)
COPY --from=bootstrap-builder /tmp/bootstrap /usr/local/bin/bootstrap
COPY --chmod=777 bootstrap.sh /usr/local/bin/bootstrap.sh

# Verify bootstrap is executable
# hadolint ignore=DL4006
RUN ls -la /usr/local/bin/bootstrap && \
    file /usr/local/bin/bootstrap && \
    echo "Testing bootstrap binary..." && \
    /usr/local/bin/bootstrap 2>&1 | head -5 || echo "Bootstrap ready (needs DATABASE_URL)"

# Run the populate and dump script
RUN su-exec postgres bootstrap.sh

# Stage 3: Final PostgreSQL image with seeded data
FROM postgres:16-alpine

# Set environment variables for PostgreSQL
ENV POSTGRES_DB=db
ENV POSTGRES_USER=user

# Copy the generated SQL dump from data-builder
COPY --from=data-builder /build/exercisedb-init.sql /docker-entrypoint-initdb.d/

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD pg_isready -U user -d db || exit 1

# Expose PostgreSQL port
EXPOSE 5432

# Run as postgres user (non-root)
USER postgres

# The postgres image automatically runs scripts in /docker-entrypoint-initdb.d/
# when the container starts for the first time
