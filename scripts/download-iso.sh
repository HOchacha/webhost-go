#!/bin/bash

# Ubuntu 버전 및 아키텍처 설정
UBUNTU_VERSION="22.04"
ARCH="amd64"
IMAGE_NAME="ubuntu-${UBUNTU_VERSION}-server-cloudimg-${ARCH}.img"
DOWNLOAD_URL="https://cloud-images.ubuntu.com/releases/${UBUNTU_VERSION}/release/${IMAGE_NAME}"

# 저장 경로 설정
TARGET_DIR="/var/lib/libvirt/images"
TARGET_PATH="${TARGET_DIR}/${IMAGE_NAME}"

# 디렉토리 생성
mkdir -p "$TARGET_DIR"

# 다운로드
echo "Downloading Ubuntu ${UBUNTU_VERSION} cloud image..."
wget -O "$TARGET_PATH" "$DOWNLOAD_URL"

echo "Image downloaded to: $TARGET_PATH"