# Stage 1: Builder - Clone repo and build PostgreSQL data
FROM node:20-slim AS builder

# Install git and PostgreSQL client tools
RUN apt-get update && \
    apt-get install -y git postgresql-client && \
    rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /build

# Clone the exercises.json repository
RUN git clone https://github.com/wrkout/exercises.json .

# Install dependencies and build PostgreSQL data
RUN npm install && \
    npm run build:psql && \
    npm run build:psql -- exercises && \
    sed -i 's/connect exercises-db/connect db/g' /build/exercises-psql.sql

# Stage 2: Final PostgreSQL image with seeded data
FROM postgres:16-alpine

# Set environment variables for PostgreSQL
ENV POSTGRES_DB=db
ENV POSTGRES_USER=user
ENV POSTGRES_PASSWORD=pass

# Copy the generated SQL file from builder
COPY --from=builder /build/exercises-psql.sql /docker-entrypoint-initdb.d/

# Expose PostgreSQL port
EXPOSE 5432

# The postgres image automatically runs scripts in /docker-entrypoint-initdb.d/
# when the container starts for the first time
