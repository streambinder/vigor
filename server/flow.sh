#!/bin/bash

set -euo pipefail
shopt -s expand_aliases

HOST="${VIGOR_HOST:-http://localhost:8080}"

alias curl="curl -s -H \"Content-Type: application/json\""
login='{"email":"user@ema.il","password":"pass"}'

curl -X POST "${HOST}"/register -d "${login}" | jq || true
TOKENS=$(curl -X POST "${HOST}"/login -d "${login}")
echo "$TOKENS" | jq
ACCESS=$(echo "$TOKENS" | jq -r .access_token)
REFRESH=$(echo "$TOKENS" | jq -r .refresh_token)
alias curl="curl -s -H \"Content-Type:application/json\" -H \"Authorization: Bearer \${ACCESS}\""

curl -X POST "${HOST}"/user/update \
	-d '{"birthdate":"01/01/2000","language":"italiano","height":180,"weight":60,"data":{"goals":["hypertrophy","hyperlordosis compensation"]}}' | jq
curl "${HOST}"/user | jq

NEW=$(curl -X POST "${HOST}"/refresh -d '{"refresh_token":"'"$REFRESH"'"}')
echo "$NEW" | jq
ACCESS=$(echo "$NEW" | jq -r .access_token)
REFRESH=$(echo "$NEW" | jq -r .refresh_token)
alias curl="curl -s -H \"Content-Type:application/json\" -H \"Authorization: Bearer \${ACCESS}\""

curl -X POST "${HOST}"/gym -d '{"name":"Basement","equipment":["barbell","bench","dumbbels","elastic bands","pull-up bar","dip station","rings","ab wheel"]}' | jq || true
curl "${HOST}"/gym/basement | jq

ts="$(date +%s%N)"
curl -X POST "${HOST}"/training -d '{"gym":"basement","duration":30}' | jq
echo "Duration: $((($(date +%s%N) - ts) / 1000000))ms"

# curl -X POST "${HOST}"/unregister | jq
