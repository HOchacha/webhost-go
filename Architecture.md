1. management_server
   사용자 요청 처리

DB와 연동하여 Hosting 정보 관리

다음 역할 수행:

VM 생성 요청 → libvirt-agent

리버스 프록시 설정 요청 → nginx-agent

게이트웨이 포트 할당 관리 (HTTP/SSH)

2. libvirt-agent (on compute node)
   HTTP API 제공 (예: /api/libvirt/create, /destroy, /status)

실제 libvirt를 사용하여 VM 정의/시작/삭제

필요한 경우 qcow2 디스크 자동 생성

3. nginx-agent (on infra node)
   HTTP API 제공 (/api/nginx/register, /remove)

/etc/nginx/sites-available/, stream.d/에 프록시 설정 추가/삭제

Nginx reload 자동 수행

