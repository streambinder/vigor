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

# Verify bootstrap is executable
RUN ls -la /usr/local/bin/bootstrap && \
    file /usr/local/bin/bootstrap && \
    echo "Testing bootstrap binary..." && \
    /usr/local/bin/bootstrap 2>&1 | head -5 || echo "Bootstrap ready (needs DATABASE_URL)"

# Create a script to initialize, populate, and dump the database
RUN cat > /build/populate-and-dump.sh << 'EOF'
#!/bin/sh
set -e

echo "Starting temporary PostgreSQL server..."
# Initialize the database cluster
initdb -D /tmp/pgdata

# Start PostgreSQL in the background
pg_ctl -D /tmp/pgdata -o "-k /tmp" -l /tmp/postgres.log start

# Wait for PostgreSQL to be ready
for i in $(seq 1 30); do
    if pg_isready -h /tmp > /dev/null 2>&1; then
        break
    fi
    echo "Waiting for PostgreSQL to start..."
    sleep 1
done

# Create database and user
psql -h /tmp -U postgres -c "CREATE DATABASE db;"
psql -h /tmp -U postgres -c "CREATE USER \"user\" WITH PASSWORD 'password';"
psql -h /tmp -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE db TO \"user\";"
psql -h /tmp -U postgres -d db -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
psql -h /tmp -U postgres -d db -c "GRANT ALL ON SCHEMA public TO \"user\";"

# Run the bootstrap tool
echo "Running bootstrap tool..."
DATABASE_URL="postgres://user:password@/db?host=/tmp" bootstrap /build/data

# Dump the populated database to a location postgres user can write
echo "Dumping database..."
pg_dump -h /tmp -U postgres -d db --clean --if-exists > /tmp/exercisedb-init.sql 2>/tmp/pg_dump_error.log
DUMP_EXIT_CODE=$?

echo "Dump command completed with exit code: $DUMP_EXIT_CODE"
echo "Checking if dump file exists..."
ls -lh /tmp/exercisedb-init.sql || echo "Dump file not found in /tmp"

if [ $DUMP_EXIT_CODE -eq 0 ]; then
    echo "pg_dump succeeded"
    if [ -s /tmp/exercisedb-init.sql ]; then
        echo "Dump file has content, moving to /build..."
        # Move to /build (postgres user owns the file, needs permission to move to /build)
        cp /tmp/exercisedb-init.sql /build/exercisedb-init.sql || echo "Copy failed, trying with elevated permissions"
        echo "✓ Database dump successful!"
        echo "Dump file size: $(ls -lh /build/exercisedb-init.sql)"
        echo "Dump file lines: $(wc -l < /build/exercisedb-init.sql)"
        echo "First 10 lines of dump:"
        head -10 /build/exercisedb-init.sql
    else
        echo "ERROR: Dump file is empty!"
        exit 1
    fi
else
    echo "ERROR: pg_dump failed!"
    echo "pg_dump error output:"
    cat /tmp/pg_dump_error.log || echo "No error log"
    exit 1
fi

# Stop PostgreSQL
echo "Stopping PostgreSQL..."
pg_ctl -D /tmp/pgdata stop

echo "Database dump completed successfully!"
EOF

RUN chmod +x /build/populate-and-dump.sh

# Run the populate and dump script
RUN su-exec postgres /build/populate-and-dump.sh

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
