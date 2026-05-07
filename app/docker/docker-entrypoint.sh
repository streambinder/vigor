#!/bin/sh
set -e

# generate build version metadata (use env var or timestamp as fallback)
BUILD_VERSION="${BUILD_VERSION:-$(date +%s)}"
BUILD_TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# generate runtime configuration with version metadata
cat >/usr/share/nginx/html/config.js <<EOF
window.ENV = {
  API_URL: "${API_URL:-http://localhost:8080}",
  FRONTEND_URL: "${FRONTEND_URL:-http://localhost:8080}",
  BUILD_VERSION: "${BUILD_VERSION}",
  BUILD_TIMESTAMP: "${BUILD_TIMESTAMP}"
};
EOF

echo "Runtime configuration generated:"
cat /usr/share/nginx/html/config.js

# inject version metadata into index.html as meta tag for debugging
if [ -f /usr/share/nginx/html/index.html ]; then
	sed -i 's|<head>|<head>\
  <meta name="app-version" content="'"${BUILD_VERSION}"'">|' /usr/share/nginx/html/index.html
	echo "Version metadata injected into index.html: ${BUILD_VERSION}"
fi

# ensure proper ownership (nginx is the default user in nginx:alpine)
chown nginx:nginx /usr/share/nginx/html/config.js
chown nginx:nginx /usr/share/nginx/html/index.html 2>/dev/null || true

# extract dns resolver from /etc/resolv.conf for nginx upstream resolution
DNS_RESOLVER=$(awk '/^nameserver/{print $2; exit}' /etc/resolv.conf)
export DNS_RESOLVER

# substitute env vars in nginx config (only these, to preserve nginx's own $variables)
# shellcheck disable=SC2016
envsubst '${API_INTERNAL_URL} ${DNS_RESOLVER}' </etc/nginx/conf.d/default.conf >/etc/nginx/conf.d/default.conf.tmp
mv /etc/nginx/conf.d/default.conf.tmp /etc/nginx/conf.d/default.conf

# start nginx as root (it will drop privileges to nginx user automatically)
exec nginx -g "daemon off;"
