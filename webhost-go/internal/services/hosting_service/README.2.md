좋습니다. 지금까지 논의한 내용을 기반으로 **Ubuntu Cloud-Init 기반 VM을 Libvirt로 생성하는 전체 절차**를 완벽하게 정리해드릴게요.

---

# ✅ Libvirt 기반 Ubuntu VM 생성 전체 절차 (Cloud-Init + VNC)

---

## ⚙️ 1. 시스템 준비

### 필수 패키지 설치 (Ubuntu 기준)

```bash
sudo apt update
sudo apt install -y qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils virtinst cloud-image-utils genisoimage
sudo systemctl enable --now libvirtd
```

---

## 📦 2. Ubuntu Cloud Image 다운로드

```bash
wget https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img -O ubuntu-22.04-base.qcow2
sudo mv ubuntu-22.04-base.qcow2 /var/lib/libvirt/images/templates/
```

---

## 📝 3. 고정 `user-data` 작성

`~/cloud-init/user-data`:

```yaml
#cloud-config
users:
  - name: ubuntu
    plain_text_passwd: '1234'
    lock_passwd: false
    shell: /bin/bash
    sudo: ["ALL=(ALL) NOPASSWD:ALL"]
chpasswd:
  expire: false
ssh_pwauth: true
```

---

## 🧩 4. VM마다 `meta-data` 생성

예: `~/cloud-init/vm-meta/vm-001-meta-data`:

```yaml
instance-id: vm-001
local-hostname: vm-001
```

---

## 💿 5. ISO 생성

```bash
genisoimage -output /var/lib/libvirt/images/cloud-init/vm-001.iso \
  -volid cidata -joliet -rock \
  ~/cloud-init/user-data ~/cloud-init/vm-meta/vm-001-meta-data
```

---

## 🧱 6. VM 디스크 생성 (템플릿 복사)

```bash
cp /var/lib/libvirt/images/templates/ubuntu-22.04-base.qcow2 /var/lib/libvirt/images/vm-001.qcow2
```

---

## 🧾 7. XML 정의 (예: `/etc/libvirt/qemu/vm-001.xml`)

```xml
<domain type='kvm'>
  <name>vm-001</name>
  <memory unit='MiB'>1024</memory>
  <vcpu>1</vcpu>
  <os>
    <type arch='x86_64'>hvm</type>
    <boot dev='hd'/>
  </os>
  <devices>
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='/var/lib/libvirt/images/vm-001.qcow2'/>
      <target dev='vda' bus='virtio'/>
    </disk>
    <disk type='file' device='cdrom'>
      <driver name='qemu' type='raw'/>
      <source file='/var/lib/libvirt/images/cloud-init/vm-001.iso'/>
      <target dev='hdb' bus='ide'/>
      <readonly/>
    </disk>
    <interface type='network'>
      <source network='default'/>
      <model type='virtio'/>
    </interface>
    <graphics type='vnc' port='-1' autoport='yes' listen='0.0.0.0'/>
  </devices>
</domain>
```

> VNC를 외부에서 접근하고 싶으면 `listen='0.0.0.0'` 로 바꿔야 함

---

## 🚀 8. VM 생성 및 실행

```bash
virsh define /etc/libvirt/qemu/vm-001.xml
virsh start vm-001
```

---

## 🔎 9. VNC 연결

### VNC 포트 확인

```bash
virsh dumpxml vm-001 | grep vnc
```

예시:

```xml
<graphics type='vnc' port='5900' autoport='yes' listen='0.0.0.0'/>
```

→ `localhost:5900` 또는 `서버IP:5900` 으로 VNC 클라이언트에서 접속

---

## 📟 10. Console 접근 (optional)

```bash
virsh console vm-001
```

문제: `cannot find character device <null>`

> 해결: XML에 `serial` 디바이스 추가 필요

```xml
<serial type='pty'>
  <target port='0'/>
</serial>
<console type='pty'>
  <target type='serial' port='0'/>
</console>
```

---

## ✅ cloud-init 적용 확인

1. VNC 또는 Console 접속 후:

```plaintext
ubuntu login: ubuntu
Password: 1234
```

2. 확인 명령어:

```bash
cat /var/lib/cloud/instance/user-data.txt
```

---

## 🧼 정리: 디렉토리 구조 예시

```
/var/lib/libvirt/images/
├── templates/
│   └── ubuntu-22.04-base.qcow2
├── vm-001.qcow2
├── cloud-init/
│   └── vm-001.iso

~/cloud-init/
├── user-data
└── vm-meta/
    └── vm-001-meta-data
```

---

## 📌 결론 요약

| 단계            | 요약                                      |
| ------------- | --------------------------------------- |
| 시스템 준비        | libvirt & cloud-utils 설치                |
| 템플릿 다운로드      | 공식 Ubuntu Cloud Image 사용                |
| cloud-init 구성 | `user-data` 고정, `meta-data` 개별 생성       |
| 디스크 준비        | 템플릿 복사 or 동적 생성                         |
| ISO 생성        | VM마다 1개씩 cloud-init ISO 생성              |
| XML 정의        | 디스크 + cloud-init ISO + VNC/serial 구성 포함 |
| 실행 및 접속       | `virsh` + VNC/Console로 확인 가능            |

---

필요하다면 이 전체 흐름을 자동화하는 **Go 함수**, **쉘 스크립트**, 또는 **API 엔드포인트 설계**도 도와드릴게요.
