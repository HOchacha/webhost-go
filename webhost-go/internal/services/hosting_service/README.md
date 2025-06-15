🧭 HostingService: 사용자에게 제공할 기능 정리
📌 1. VM 생성
사용자 요청을 받아 Hosting 엔티티 생성

Plan 정보 확인 (DefaultPlans)

디스크 경로, 도메인 이름 생성

libvirt-agent에 VM 생성 요청

nginx-agent에 프록시 등록 요청

DB에 Hosting 정보 저장

📌 2. VM 상태 변경
시작 / 정지 / 재시작 / 삭제

libvirt-agent에 요청

삭제 시에는 nginx 설정 제거 + 디스크 파일도 제거

📌 3. VM 정보 조회
전체 리스트 조회 (ListHostingForUser)

특정 사용자에 대한 VM 목록 반환

단일 VM 상세 조회 (GetHostingDetail)

ID 또는 도메인 이름으로 단일 VM 정보 확인

📌 4. VM 수정 (옵션)
Plan 변경 (업그레이드, 다운그레이드)

DNS / Hostname 변경 등

⚠️ 단, VM 실행 중 Plan 수정은 제한하거나 중지 후 변경하도록 할 수 있음.