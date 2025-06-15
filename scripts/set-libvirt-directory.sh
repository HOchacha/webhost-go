#!/bin/bash

set -e

# 기본 경로
BASE_DIR="/var/lib/libvirt/images"
CLOUD_INIT_SHARE="$BASE_DIR/cloud-init/share"
TEMPLATES_DIR="$BASE_DIR/templates"
USER_DATA="$CLOUD_INIT_SHARE/user-data"
UBUNTU_IMG="$TEMPLATES_DIR/ubuntu-22.04-server-cloudimg-amd64.img"
UBUNTU_IMG_URL="https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img"

echo "🔧 디렉토리 생성..."
mkdir -p "$CLOUD_INIT_SHARE"
mkdir -p "$BASE_DIR/instances"
mkdir -p "$TEMPLATES_DIR"

echo "📄 user-data 작성..."
cat > "$USER_DATA" <<EOF
#cloud-config

package_update: true
package_upgrade: true

packages:
  - nginx

users:
  - name: ubuntu
    plain_text_passwd: 'ubuntu'
    lock_passwd: false
    shell: /bin/bash
    sudo: ["ALL=(ALL) NOPASSWD:ALL"]
chpasswd:
  expire: false
ssh_pwauth: true

runcmd:
  - systemctl enable nginx
  - systemctl start nginx
  - ln -s /var/www/html /home/ubuntu/www
  - chown -R ubuntu:ubuntu /var/www/html
EOF

echo "📥 Ubuntu Cloud Image 다운로드..."
if [ ! -f "$UBUNTU_IMG" ]; then
    curl -L -o "$UBUNTU_IMG" "$UBUNTU_IMG_URL"
else
    echo "✅ 이미지가 이미 존재합니다: $UBUNTU_IMG"
fi

echo "✅ 초기화 완료!"

# 현재 사용자
CURRENT_USER=$(whoami)

echo "🔐 $CURRENT_USER 계정에 libvirt 및 kvm 그룹 권한 추가 중..."

sudo usermod -aG libvirt "$CURRENT_USER"
sudo usermod -aG kvm "$CURRENT_USER"

echo "✅ 권한이 성공적으로 추가되었습니다."

echo -e "\n⚠️ 변경 사항을 적용하려면 로그아웃 후 다시 로그인하거나, 다음 명령어를 실행하세요:\n"
echo "exec su - $CURRENT_USER"