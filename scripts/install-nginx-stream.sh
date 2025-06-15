#!/bin/bash

# ===============================
# 🔧 Build Nginx with stream module and register as systemd service
# ===============================
set -e

NGINX_VERSION="1.24.0"
NGINX_PREFIX="/usr/local/nginx"
NGINX_SERVICE_NAME="nginx-custom"
NGINX_SYMLINK="/usr/bin/nginx"

echo "🚀 Installing build dependencies..."
sudo apt update
sudo apt install -y build-essential libpcre3 libpcre3-dev zlib1g-dev libssl-dev curl

echo "📥 Downloading Nginx source..."
curl -O http://nginx.org/download/nginx-${NGINX_VERSION}.tar.gz
tar -xzf nginx-${NGINX_VERSION}.tar.gz
cd nginx-${NGINX_VERSION}

echo "⚙️ Configuring Nginx with --with-stream..."
./configure --prefix=${NGINX_PREFIX} --with-stream --with-http_ssl_module

echo "🔨 Building and installing Nginx..."
make -j$(nproc)
sudo make install

echo "🔗 Creating symlink: ${NGINX_SYMLINK} → ${NGINX_PREFIX}/sbin/nginx"
sudo ln -sf ${NGINX_PREFIX}/sbin/nginx ${NGINX_SYMLINK}

echo "📝 Creating systemd service: ${NGINX_SERVICE_NAME}"
SERVICE_FILE="/etc/systemd/system/${NGINX_SERVICE_NAME}.service"

sudo mkdir -p /var/log/nginx
sudo chown -R root:root /var/log/nginx

sudo tee ${SERVICE_FILE} > /dev/null <<EOF
[Unit]
Description=Custom Nginx
After=network.target

[Service]
ExecStart=${NGINX_PREFIX}/sbin/nginx
ExecReload=${NGINX_PREFIX}/sbin/nginx -s reload
ExecStop=${NGINX_PREFIX}/sbin/nginx -s quit
Restart=always

[Install]
WantedBy=multi-user.target
EOF

echo "🔄 Reloading systemd daemon and enabling service..."
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable ${NGINX_SERVICE_NAME}
sudo systemctl start ${NGINX_SERVICE_NAME}

echo "✅ Nginx custom build with stream module is installed and running as ${NGINX_SERVICE_NAME}"

