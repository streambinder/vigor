#!/bin/bash

set -euo pipefail

curl -s -H "Content-Type: application/json" -X POST http://localhost:8080/register -d '{"email":"user@ema.il","password":"pass"}' | jq

TOKENS=$(curl -s -H "Content-Type: application/json" -X POST http://localhost:8080/login -d '{"email":"user@ema.il","password":"pass"}')
echo "$TOKENS" | jq
ACCESS=$(echo "$TOKENS" | jq -r .access_token)
REFRESH=$(echo "$TOKENS" | jq -r .refresh_token)

curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS"  -X POST http://localhost:8080/user/update \
    -d '{"birthdate":"01/01/2000","language":"it","height":180,"weight":60,"data":{"goals":["hypertrophy","hyperlordosis compensation"]}}' | jq
curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS" http://localhost:8080/user | jq

NEW=$(curl -s -H "Content-Type: application/json" -X POST http://localhost:8080/refresh -d '{"refresh_token":"'$REFRESH'"}')
echo "$NEW" | jq
ACCESS=$(echo "$NEW" | jq -r .access_token)
REFRESH=$(echo "$NEW" | jq -r .refresh_token)

curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS" http://localhost:8080/training | jq

# curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS" http://localhost:8080/unregister | jq
