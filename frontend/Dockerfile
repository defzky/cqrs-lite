FROM nginx:alpine

# Install curl for healthcheck
RUN apk add --no-cache curl

# Copy static files
COPY static/ /usr/share/nginx/html/

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
    CMD curl -fsS http://localhost/ || exit 1
