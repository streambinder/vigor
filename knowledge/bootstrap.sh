#!/bin/sh
set -e

echo "Starting temporary PostgreSQL server..."
# Initialize the database cluster
initdb -D /tmp/pgdata

# Start PostgreSQL in the background
pg_ctl -D /tmp/pgdata -o "-k /tmp" -l /tmp/postgres.log start

# Wait for PostgreSQL to be ready
for _ in $(seq 1 30); do
	if pg_isready -h /tmp >/dev/null 2>&1; then
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
pg_dump -h /tmp -U postgres -d db --clean --if-exists >/tmp/init.sql 2>/tmp/pg_dump_error.log
DUMP_EXIT_CODE=$?

echo "Dump command completed with exit code: $DUMP_EXIT_CODE"
echo "Checking if dump file exists..."
ls -lh /tmp/init.sql || echo "Dump file not found in /tmp"

if [ $DUMP_EXIT_CODE -eq 0 ]; then
	echo "pg_dump succeeded"
	if [ -s /tmp/init.sql ]; then
		echo "Dump file has content, moving to /build..."
		cp /tmp/init.sql /build/init.sql || echo "Copy failed, trying with elevated permissions"
		echo "✓ Database dump successful!"
		echo "Dump file size: $(ls -lh /build/init.sql)"
		echo "Dump file lines: $(wc -l </build/init.sql)"
		echo "First 10 lines of dump:"
		head -10 /build/init.sql
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
