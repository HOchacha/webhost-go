#!/bin/bash
set -e

NGINX_PREFIX="/usr/local/nginx"
NGINX_CONF="${NGINX_PREFIX}/conf/nginx.conf"
NGINX_SERVICE_NAME="nginx-custom"
WEBHOST_CONF="${NGINX_PREFIX}/conf/sites-available/webhost.conf"
LOCATIONS_DIR="${NGINX_PREFIX}/conf/sites-available/locations"

echo "📁 Creating required NGINX config and log directories..."
sudo mkdir -p ${NGINX_PREFIX}/conf/stream.d
sudo mkdir -p ${LOCATIONS_DIR}
sudo mkdir -p ${NGINX_PREFIX}/conf/sites-available
sudo mkdir -p ${NGINX_PREFIX}/conf/sites-enabled
sudo mkdir -p /var/log/nginx
sudo touch /var/log/nginx/access.log /var/log/nginx/error.log

echo "🗑️ Removing existing nginx.conf if it exists..."
sudo rm -f "${NGINX_CONF}"

echo "📝 Writing fresh nginx.conf..."
sudo tee "${NGINX_CONF}" > /dev/null <<EOF
worker_processes  1;

events {
    worker_connections  1024;
}

http {
    include       mime.types;
    default_type  application/octet-stream;

    sendfile        on;
    keepalive_timeout  65;

    access_log /var/log/nginx/access.log;
    error_log  /var/log/nginx/error.log;

    gzip on;

    include ${NGINX_PREFIX}/conf/sites-enabled/*;
}

stream {
    include ${NGINX_PREFIX}/conf/stream.d/*.conf;
}
EOF

echo "📝 Writing webhost.conf..."
sudo tee "${WEBHOST_CONF}" > /dev/null <<EOF
server {
    listen 80;
    server_name _;


    error_log /var/log/nginx/error.log;
    access_log /var/log/nginx/access.log;

    include ${LOCATIONS_DIR}/*.conf;
}
EOF

echo "🔗 Linking webhost.conf to sites-enabled..."
sudo ln -sf "${WEBHOST_CONF}" "${NGINX_PREFIX}/conf/sites-enabled/webhost.conf"

echo "🔁 Reloading ${NGINX_SERVICE_NAME} to apply config..."
sudo systemctl reload ${NGINX_SERVICE_NAME}

echo "✅ NGINX configuration fully set up and running!"