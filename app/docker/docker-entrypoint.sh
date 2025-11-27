#!/bin/sh
set -e

# Generate runtime configuration
cat > /usr/share/nginx/html/config.js <<EOF
window.ENV = {
  API_URL: "${API_URL:-http://localhost:8080}"
};
EOF

echo "Runtime configuration generated:"
cat /usr/share/nginx/html/config.js

# Ensure proper ownership (nginx is the default user in nginx:alpine)
chown nginx:nginx /usr/share/nginx/html/config.js

# Start nginx as root (it will drop privileges to nginx user automatically)
exec nginx -g "daemon off;"
