echo "📁 Creating additional config directories..."
sudo mkdir -p ${NGINX_PREFIX}/conf/stream.d
sudo mkdir -p ${NGINX_PREFIX}/conf/sites-available/locations

echo "🧩 Ensuring nginx.conf includes stream.d and locations..."

NGINX_CONF="${NGINX_PREFIX}/conf/nginx.conf"

# Backup the original nginx.conf
sudo cp ${NGINX_CONF} ${NGINX_CONF}.bak

# Add include directives if not already present
if ! grep -q "include ${NGINX_PREFIX}/conf/stream.d/\*.conf;" "${NGINX_CONF}"; then
  sudo sed -i "/^stream {/a \    include ${NGINX_PREFIX}/conf/stream.d/*.conf;" "${NGINX_CONF}" || \
    echo -e "\nstream {\n    include ${NGINX_PREFIX}/conf/stream.d/*.conf;\n}" | sudo tee -a "${NGINX_CONF}"
fi

if ! grep -q "include ${NGINX_PREFIX}/conf/sites-available/locations/\*.conf;" "${NGINX_CONF}"; then
  sudo sed -i "/^http {/a \    include ${NGINX_PREFIX}/conf/sites-available/locations/*.conf;" "${NGINX_CONF}" || \
    echo -e "\nhttp {\n    include ${NGINX_PREFIX}/conf/sites-available/locations/*.conf;\n}" | sudo tee -a "${NGINX_CONF}"
fi

echo "🔁 Reloading ${NGINX_SERVICE_NAME} to apply config changes..."
sudo systemctl reload ${NGINX_SERVICE_NAME}

echo "✅ Directories stream.d and locations created and included in nginx.conf"

echo "📁 Creating sites-available and sites-enabled directories..."
sudo mkdir -p ${NGINX_PREFIX}/conf/sites-available
sudo mkdir -p ${NGINX_PREFIX}/conf/sites-enabled

echo "🧩 Ensuring nginx.conf includes sites-enabled configs..."
NGINX_CONF="${NGINX_PREFIX}/conf/nginx.conf"

if ! grep -q "include ${NGINX_PREFIX}/conf/sites-enabled/\*.conf;" "${NGINX_CONF}"; then
  sudo sed -i "/^http {/a \    include ${NGINX_PREFIX}/conf/sites-enabled/*.conf;" "${NGINX_CONF}" || \
    echo -e "\nhttp {\n    include ${NGINX_PREFIX}/conf/sites-enabled/*.conf;\n}" | sudo tee -a "${NGINX_CONF}"
fi

echo "🔗 Linking webhost.conf..."
if [ -f "${NGINX_PREFIX}/conf/sites-available/webhost.conf" ]; then
  sudo ln -sf ${NGINX_PREFIX}/conf/sites-available/webhost.conf ${NGINX_PREFIX}/conf/sites-enabled/webhost.conf
else
  echo "⚠️  Warning: webhost.conf does not exist in sites-available. Please create it manually."
fi

echo "🔁 Reloading ${NGINX_SERVICE_NAME} to apply config..."
sudo systemctl reload ${NGINX_SERVICE_NAME}

echo "✅ webhost.conf linked and activated"