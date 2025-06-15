templates/에 있는 ubuntu-22.04-server-cloudimg-amd64.img는 부팅 디스크 템플릿으로서, 각 VM의 OS 루트 디스크로 복제되어 사용됨.

cloud-init/share/user-data는 모든 VM이 공통으로 사용하는 cloud-init user-data (예: 사용자 계정 설정, SSH 키 추가 등).

instances/<vm-name>/은 VM별 디렉토리로, 다음과 같은 구성으로 진행하면 좋아:

csharp
복사
편집
instances/<vm-name>/
├── disk.qcow2           # base 템플릿에서 복제한 부팅 디스크
├── meta-data            # VM 전용 metadata (hostname, instance-id 등)
└── cloud-init.iso       # 위의 user-data와 meta-data로 만든 ISO