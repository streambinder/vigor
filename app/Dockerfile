FROM ghcr.io/cirruslabs/flutter:stable AS builder

WORKDIR /app
COPY . .
RUN flutter config --enable-web \
    && flutter pub get \
    && flutter pub run build_runner build --delete-conflicting-outputs \
    && flutter build web --release

FROM nginx:1.27-alpine

COPY --from=builder /app/build/web /usr/share/nginx/html
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY docker/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh && \
chown -R nginx:nginx /usr/share/nginx/html
EXPOSE 80
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1
ENTRYPOINT ["/docker-entrypoint.sh"]

LABEL org.opencontainers.image.source=https://github.com/streambinder/vigor
