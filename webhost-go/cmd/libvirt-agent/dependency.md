# System APT Install
```bash
sudo apt update
sudo apt install -y \
    libvirt-daemon-system \
    libvirt-clients \
    qemu-kvm \
    virtinst \
    bridge-utils \
    genisoimage \
    net-tools

```


# Permission Setting
```bash
# libvirt 그룹에 사용자 추가
sudo usermod -aG libvirt $(whoami)

# 적용 후 재로그인 필요
newgrp libvirt
```