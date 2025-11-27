FROM debian:latest AS builder

ARG FLUTTER_VERSION=stable
ENV PATH="/flutter/bin:${PATH}"
WORKDIR /app
COPY . .

RUN apt-get update && apt-get install -y \
    curl \
    git \
    wget \
    unzip \
    xz-utils \
    zip \
    libglu1-mesa \
    && git clone --depth 1 --branch ${FLUTTER_VERSION} https://github.com/flutter/flutter.git /flutter \
    && flutter config --enable-web \
    && flutter precache --web \
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
