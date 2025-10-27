#!/bin/bash

set -euo pipefail
shopt -s expand_aliases

alias curl="curl -s -H \"Content-Type: application/json\""
login='{"email":"user@ema.il","password":"pass"}'

curl -X POST http://localhost:8080/register -d "${login}" | jq || true
TOKENS=$(curl -X POST http://localhost:8080/login -d '{"email":"user@ema.il","password":"pass"}')
echo "$TOKENS" | jq
ACCESS=$(echo "$TOKENS" | jq -r .access_token)
REFRESH=$(echo "$TOKENS" | jq -r .refresh_token)
alias curl="curl -s -H \"Content-Type:application/json\" -H \"Authorization: Bearer \${ACCESS}\""

curl -X POST http://localhost:8080/user/update \
	-d '{"birthdate":"01/01/2000","language":"it","height":180,"weight":60,"data":{"goals":["hypertrophy","hyperlordosis compensation"]}}' | jq
curl http://localhost:8080/user | jq

NEW=$(curl -X POST http://localhost:8080/refresh -d '{"refresh_token":"'"$REFRESH"'"}')
echo "$NEW" | jq
ACCESS=$(echo "$NEW" | jq -r .access_token)
REFRESH=$(echo "$NEW" | jq -r .refresh_token)
alias curl="curl -s -H \"Content-Type:application/json\" -H \"Authorization: Bearer \${ACCESS}\""

curl -X POST http://localhost:8080/gym -d '{"name":"Basement","equipment":["barbell","bench","dumbbels","elastic bands","pull-up bar","dip station","rings","ab wheel"]}' | jq || true
curl http://localhost:8080/gym/basement | jq

curl -X POST http://localhost:8080/training -d '{"gym":"basement","duration":30}' | jq

# curl -X POST http://localhost:8080/unregister | jq
