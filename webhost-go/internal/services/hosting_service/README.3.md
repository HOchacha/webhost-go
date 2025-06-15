네, 맞습니다! `meta-data`에서는 **네트워크 구성도 가능합니다.**

---

## ✅ 핵심 정리

* `meta-data`는 **cloud-init이 초기 부팅 시 네트워크를 설정할 수 있도록 지시하는 곳**입니다.
* Ubuntu에서는 보통 \*\*Netplan 포맷(version: 2)\*\*을 사용해 `meta-data`에 네트워크 설정을 포함시킬 수 있습니다.

---

## ✅ DHCP 네트워크 예시

```yaml
instance-id: my-vm-001
local-hostname: my-vm
network:
  version: 2
  ethernets:
    eth0:
      dhcp4: true
```

> 위 설정은 `eth0` 인터페이스에 대해 DHCP를 통해 IPv4 주소를 받도록 지정합니다.

---

## ✅ Static IP 설정 예시

```yaml
instance-id: my-vm-002
local-hostname: my-static-vm
network:
  version: 2
  ethernets:
    eth0:
      addresses:
        - 192.168.122.100/24
      gateway4: 192.168.122.1
      nameservers:
        addresses:
          - 8.8.8.8
          - 1.1.1.1
```

이 설정은 다음을 의미합니다:

* VM의 eth0에 고정 IP 192.168.122.100/24 설정
* 게이트웨이는 192.168.122.1
* DNS는 구글과 Cloudflare

---

## ✅ 주의할 점

1. **Netplan을 쓰는 Ubuntu 버전** (>= 17.10)에서는 반드시 `network: version: 2` 형식이어야 합니다.
2. cloud-init이 이를 읽어서 `/etc/netplan/50-cloud-init.yaml`을 만들어줍니다.
3. `meta-data`는 ISO 파일로 VM에 삽입해야 cloud-init이 인식합니다.

---

## ✅ ISO 생성 명령 예시

```bash
cloud-localds my-vm.iso user-data meta-data
```

이 ISO를 libvirt VM에 **CD-ROM 드라이브로 삽입**하면 cloud-init이 자동 적용됩니다.

---

원한다면 `meta-data`, `user-data`, 그리고 VM 생성 스크립트를 모두 자동화해줄 수도 있어요.
