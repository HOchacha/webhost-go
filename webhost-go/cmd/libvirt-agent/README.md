# 디스크 및 파일 관리 구조 문서
```
/var/lib/libvirt/
├── images/                # VM 디스크 파일 (.qcow2 등)
│   ├── vm-001.qcow2
│   └── vm-002.qcow2
├── templates/             # 템플릿 이미지 (원본, readonly)
│   └── ubuntu-22.04-template.qcow2
├── cloud-init/            # cloud-init ISO 등 초기화에 필요한 부수 파일
│   ├── vm-001-seed.iso
│   └── vm-002-seed.iso
└── iso/                   # OS 설치용 ISO (선택)
    └── ubuntu-22.04.iso
```

