# Stage 1: Builder - Download ExerciseDB data and generate SQL
FROM node:20-slim AS builder

# Install required tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends curl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /build

# Download and extract the latest ExerciseDB API data
RUN curl -L https://github.com/ExerciseDB/exercisedb-api/archive/refs/heads/main.tar.gz -o exercisedb-api.tar.gz && \
    tar -xzf exercisedb-api.tar.gz && \
    mv exercisedb-api-main exercisedb-api && \
    rm exercisedb-api.tar.gz

# Copy SQL generation script
COPY generate-sql.js /build/

# Generate SQL initialization script from JSON data
RUN node generate-sql.js /build/exercisedb-api/src/data > /build/exercisedb-init.sql

# Stage 2: Final PostgreSQL image with seeded data
FROM postgres:16-alpine

# Set environment variables for PostgreSQL
ENV POSTGRES_DB=db
ENV POSTGRES_USER=user

# Copy the generated SQL file from builder
COPY --from=builder /build/exercisedb-init.sql /docker-entrypoint-initdb.d/

# Use postgres user (default for postgres image, but making it explicit)
USER postgres

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD pg_isready -U user -d db || exit 1

# Expose PostgreSQL port
EXPOSE 5432

# The postgres image automatically runs scripts in /docker-entrypoint-initdb.d/
# when the container starts for the first time
