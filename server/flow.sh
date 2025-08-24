curl -s -H "Content-Type: application/json" -X POST http://localhost:8080/register -d '{"email":"user@ema.il","password":"pass"}'

TOKENS=$(curl -s -H "Content-Type: application/json" -X POST http://localhost:8080/login -d '{"email":"user@ema.il","password":"pass"}')
echo "$TOKENS" | jq
ACCESS=$(echo "$TOKENS" | jq -r .access_token)
REFRESH=$(echo "$TOKENS" | jq -r .refresh_token)

curl -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS" http://localhost:8080/profile | jq

NEW=$(curl -H "Content-Type: application/json" -X POST http://localhost:8080/refresh -d '{"refresh_token":"'$REFRESH'"}')
echo "$NEW" | jq
ACCESS=$(echo "$NEW" | jq -r .access_token)
REFRESH=$(echo "$NEW" | jq -r .refresh_token)

curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS" http://localhost:8080/logout -d '{"refresh_token":"'$REFRESH'"}' | jq
