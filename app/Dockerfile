# hadolint ignore=DL3006
FROM ghcr.io/adrianjagielak/flutter:3.47.0 AS builder

WORKDIR /app

# copy dependencies first for better layer caching
COPY pubspec.yaml pubspec.lock ./
RUN flutter config --enable-web && flutter pub get

# copy source code after dependencies are cached
COPY . .
# note: --web-renderer is deprecated in Flutter 3.22+, canvaskit is now the default
RUN dart run build_runner build --delete-conflicting-outputs \
    && flutter build web --release

FROM nginx:1.31.3-alpine

COPY --from=builder /app/build/web /usr/share/nginx/html
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY docker/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh && \
chown -R nginx:nginx /usr/share/nginx/html
EXPOSE 80
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1
# entrypoint runs as root to bind port 80 + chown writable dirs, then nginx
# master spawns worker processes as the unprivileged `nginx` user (NGINX default)
# checkov:skip=CKV_DOCKER_3:nginx official image drops privileges in master/worker model
ENTRYPOINT ["/docker-entrypoint.sh"]

LABEL org.opencontainers.image.source=https://github.com/streambinder/vigor
