#!/bin/bash
set -e

# 변수 설정
UBUNTU_VERSION="22.04"
IMAGE_NAME="ubuntu-${UBUNTU_VERSION}-server-cloudimg-amd64.img"
IMAGE_URL="https://cloud-images.ubuntu.com/releases/${UBUNTU_VERSION}/release/${IMAGE_NAME}"
TEMPLATE_DIR="/var/lib/libvirt/images/templates"

# libvirt 설치 확인 (virsh 명령 확인)
if ! command -v virsh &> /dev/null; then
    echo "[오류] libvirt 또는 virsh 명령이 존재하지 않습니다."
    echo "      libvirt 패키지를 먼저 설치하세요. (예: sudo apt install libvirt-daemon-system libvirt-clients)"
    exit 1
fi

# sudo 확인
if [[ $EUID -ne 0 ]]; then
    echo "[오류] 이 스크립트는 root 권한이 필요합니다. sudo로 실행하세요."
    exit 1
fi

# 디렉토리 생성
mkdir -p "${TEMPLATE_DIR}"

# 이미지가 이미 존재하는지 확인
if [ -f "${TEMPLATE_DIR}/${IMAGE_NAME}" ]; then
    echo "[정보] 이미지가 이미 존재합니다: ${TEMPLATE_DIR}/${IMAGE_NAME}"
else
    echo "[다운로드] 이미지 받는 중: ${IMAGE_URL}"
    curl -o "${TEMPLATE_DIR}/${IMAGE_NAME}" -L "${IMAGE_URL}"
    echo "[완료] 다운로드 완료: ${TEMPLATE_DIR}/${IMAGE_NAME}"
fi

# 권한 확인 및 소유권 설정
chown root:kvm "${TEMPLATE_DIR}/${IMAGE_NAME}"
chmod 644 "${TEMPLATE_DIR}/${IMAGE_NAME}"

echo "[완료] Ubuntu 클라우드 이미지가 성공적으로 준비되었습니다."