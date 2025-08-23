docker build -t app .
docker run -d --name db -e POSTGRES_PASSWORD=pass -e POSTGRES_USER=user -e POSTGRES_DB=mydb -p 5432:5432 postgres:15
docker run -d --name app --link postgres -p 8080:8080 -e DATABASE_URL="postgres://user:pass@db:5432/mydb?sslmode=disable" app

docker compose up -d
psql "postgres://user:pass@localhost:5432/db?sslmode=disable" -f schema.sql

export DATABASE_URL="postgres://user:pass@localhost:5432/db?sslmode=disable"
export ACCESS_SECRET="super-secret-access-please-change"
export REFRESH_SECRET="super-secret-refresh-please-change"
go mod tidy
go run .

curl -i -X POST localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"P@ssw0rd!"}'

TOKENS=$(curl -s -X POST localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"P@ssw0rd!"}')
echo "$TOKENS" | jq
ACCESS=$(echo "$TOKENS" | jq -r .access_token)
REFRESH=$(echo "$TOKENS" | jq -r .refresh_token)

curl -s localhost:8080/me -H "Authorization: Bearer $ACCESS" | jq

NEW=$(curl -s -X POST localhost:8080/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH\"}")
echo "$NEW" | jq
ACCESS=$(echo "$NEW" | jq -r .access_token)
REFRESH=$(echo "$NEW" | jq -r .refresh_token)

curl -s -X POST localhost:8080/logout \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH\"}" | jq
