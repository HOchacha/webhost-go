요구사항 ID	명칭	충족 여부	설명
SFR-WHT-001	VM 생성	❌ 부분만 충족	libvirt를 통한 VM 생성 자체는 아직 미구현, 주석만 있음 (// → 추후 s.libvirtClient.CreateVM(...))
SFR-WHT-003	프록시 구성 (nginx-agent)	✅	Register(agent nginx.AgentInfo)로 nginx-agent 호출 구현 완료
SFR-WHT-004	생성 정보 DB 기록	✅	repo.Create(hosting)를 통해 DB에 생성 정보 저장
SFR-WHM-006	VM 상태 조회	❌	관련 메서드 없음
SFR-WHM-005	VM 중지	❌	관련 메서드 없음
SFR-WHM-003	VM 삭제	❌	관련 메서드 없음
SFR-WHM-004	프록시 제거 (nginx-agent)	❌	프록시 제거용 Unregister 메서드 없음
SFR-WHM-005	전체 호스팅 목록 조회	✅	ListHostings()로 전체 호스팅 조회 가능
SFR-WHT-002	cloud-init ISO 구성	❌	해당 기능은 포함되지 않음
SFR-ACC-004	사용자 정보 조회	❌	이건 user_service 쪽 범위임