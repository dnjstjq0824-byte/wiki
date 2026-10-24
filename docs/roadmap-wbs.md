# Treewiki 100-Step Master Development Roadmap & WBS

## PHASE 1: 전체 기능 및 아키텍처 분석 & 설계 (01~10) [COMPLETED]
- [x] **01. Treewiki 전체 기능 인벤토리 명세** (`docs/treewiki-feature-inventory.csv`)
- [x] **02. 위키 엔진 기능 동등성 매트릭스 분석** (`docs/feature-parity-matrix.csv`)
- [x] **03. 역할 기반 권한 제어(RBAC/ABAC) 매트릭스 설계** (`docs/rbac-matrix.yaml`)
- [x] **04. 전체 화면 및 라우팅 레지스트리 작성** (`docs/screen-registry.yaml`)
- [x] **05. 이벤트 기반 아키텍처 및 감사 메시지 카탈로그 정의** (`docs/event-catalog.yaml`)
- [x] **06. 편집 정책, 저작권(CC BY-SA 4.0) 및 규제 준수 규격 수립** (`docs/policy-spec.md`)
- [x] **07. 개인정보 보호, 익명 IP 마스킹 및 오버사이트 규격** (`docs/privacy-data-protection.md`)
- [x] **08. 위키 엔진 위협 모델링(STRIDE) 및 CSP v3 보안 설계** (`docs/threat-model.md`)
- [x] **09. 가독성 중심 타이포그래피 & 디자인 시스템 설계** (`docs/design-system.md`)
- [x] **10. 100단계 마스터 개발 로드맵 WBS 수립** (`docs/roadmap-wbs.md`)

---

## PHASE 2: 모노레포 구축 및 기초 인프라 (11~20)
- [ ] **11. 전체 Monorepo 환경 구축** (pnpm workspace, Go module, TurboRepo, TypeScript 베이스)
- [ ] **12. PostgreSQL 16 마스터 스키마 DDL 작성** (25개 핵심 테이블, 외래키, 인덱스, 트리거)
- [ ] **13. Docker Compose 오케스트레이션 환경 구축** (PostgreSQL 16, Redis 7.2, MinIO, Meilisearch)
- [ ] **14. Go 백엔드 API 게이트웨이 및 미들웨어 스켈레톤** (Chi Router, CORS, Security Headers, Request ID)
- [ ] **15. Protobuf / gRPC 계약 규격 정의** (wiki.proto, revision.proto, topictree.proto)
- [ ] **16. Nuxt 3 / Vue 3 웹 클라이언트 스캐폴딩** (Tailwind CSS, Pinia, Lucide Icons, Dark/Light 테마)
- [ ] **17. 공유 패키지(Shared Packages) 설정** (`packages/types`, `packages/ui`)
- [ ] **18. 환경 변수 검증 및 설정 로더 구축** (Viper, .env.example)
- [ ] **19. CI/CD 파이프라인 워크플로 작성** (.github/workflows/ci.yml: 린트, 타입체크, 유닛 테스트)
- [ ] **20. 시스템 헬스체크 및 의존성 핑 엔드포인트 구현** (`/healthz`, `/ready`)

---

## PHASE 3: 계정, 프로필 및 RBAC/보안 코어 (21~30)
- [ ] **21. 사용자 회원가입 및 이메일 인증 플로우 구현**
- [ ] **22. JWT 토큰 발급 및 리프레시 토큰 로테이션 (HttpOnly 쿠키)**
- [ ] **23. 비밀번호 해싱 (Argon2id) 및 보안 자격증명 저장소**
- [ ] **24. 다단계 인증 (MFA / TOTP) 등록 및 검증 엔드포인트**
- [ ] **25. 9단계 역할 기반 권한 검증 미들웨어 구현 (RBAC)**
- [ ] **26. 익명 사용자(IP) 식별 및 자동 마스킹 파이프라인 (IPv4/IPv6)**
- [ ] **27. 자동승인(AutoConfirmed) 자격 판정 배치 및 실시간 확인 엔진**
- [ ] **28. 사용자 공개 프로필 및 활동 배지/기여 카운터**
- [ ] **29. 개인정보 수정 및 비밀번호 변경 워크플로**
- [ ] **30. 계정 비활성화 및 GDPR/PIPA 익명화 스크러빙 핸들러**

---

## PHASE 4: 위키 코어 문서 및 네임스페이스 엔진 (31~40)
- [ ] **31. 10대 네임스페이스(Main, Talk, User, Project, Template, Category 등) 매핑 라우터**
- [ ] **32. 표제어(Title) 정규화 및 슬러그화 (Unicode NFKC, 공백 언더스코어 치환)**
- [ ] **33. 문서 생성 및 기본 메타데이터 저장 핸들러**
- [ ] **34. 문서 수정 및 낙관적 동시성 락(Optimistic Concurrency ETag) 검증**
- [ ] **35. 자동/수동 리다이렉트 엔진 (#REDIRECT 처리 및 순환 참조 방지)**
- [ ] **36. 역링크(WhatLinksHere) 인덱싱 및 조회 API**
- [ ] **37. 다중 부모 카테고리 계층(DAG) 탐색 및 문서 목록 렌더링**
- [ ] **38. 문서 보호(Page Protection) 레벨 설정 및 접근 통제**
- [ ] **39. 문서 이동(Rename/Move) 및 기존 표제어 리다이렉트 자동 생성**
- [ ] **40. 문서 삭제(Soft Delete) 및 복구(Undelete) 핸들러**

---

## PHASE 5: 위키 문법/마크다운 하이브리드 파서 & 서식 (41~50)
- [ ] **41. 렉서(Lexer) 및 토크나이저 구축 (Wikitext 태그 및 Markdown 마크업 인식)**
- [ ] **42. 내부 링크(`[[제목|표시문구]]`) 및 외부 링크 변환기**
- [ ] **43. 자동 목차(TOC) 추출기 (H1~H6 헤딩 파싱 및 중첩 트리 구성)**
- [ ] **44. 인라인 각주(`<ref>...</ref>`) 및 하단 참고문헌 바인딩 엔진**
- [ ] **45. 인포박스(Infobox) 템플릿 매크로 파서 및 레이아웃 컴포넌트**
- [ ] **46. LaTeX 수학 수식 블록 파서 (KaTeX 클라이언트/서버 렌더링)**
- [ ] **47. 표(Table) 서식 파서 및 정렬/필터 지원**
- [ ] **48. 접기/펼치기(Collapsible Box) 및 알림(Callout) 블록 구현**
- [ ] **49. 서버사이드 HTML 살균기 (Bluemonday DOMPurify 통합)**
- [ ] **50. 분할 화면 실시간 편집기 (Split-Pane Live Markdown/Wiki Editor)**

---

## PHASE 6: 리비전 관리, Myers Diff, 롤백 & 충돌 병합 (51~60)
- [ ] **51. 불변 리비전 레저(Immutable Revision Ledger) 저장 모델**
- [ ] **52. Myers Diff 알고리즘 기반 줄/단어 단위 토큰화 비교 엔진**
- [ ] **53. 나란히 보기(Side-by-Side) 및 한 줄 보기(Inline) Diff 뷰어 컴포넌트**
- [ ] **54. 원클릭 롤백(Fast Rollback) 엔진 (악의적 연속 편집 즉시 되돌리기)**
- [ ] **55. 수동 리비전 복원(Revert to Revision) 및 감사 로그 기록**
- [ ] **56. 3-Way 텍스트 병합 알고리즘 및 겹치지 않는 변경 자동 병합**
- [ ] **57. 편집 충돌(Edit Conflict) 시각적 해결 UI 컴포넌트**
- [ ] **58. 편집 요약(Edit Summary) 및 자동 태깅(봇, 사소한 편집, 롤백)**
- [ ] **59. 리비전 텍스트 압축 엔진 (zstandard 히스토리 아카이빙)**
- [ ] **60. 문서 히스토리 타임라인 및 기여자 통계 집계**

---

## PHASE 7: 계층형 토픽 트리 및 인터랙티브 지식 그래프 (61~70)
- [ ] **62. 토픽 트리 DAG 데이터 구조 및 방향성 엣지 스키마**
- [ ] **62. 계층형 토픽 탐색기 재귀 CTE 쿼리 최적화**
- [ ] **63. D3.js 기반 인터랙티브 방사형/트리 뷰어 컴포넌트**
- [ ] **64. 포스 다이렉티드(Force-Directed) 지식 네트워크 그래프 렌더러**
- [ ] **65. 문서와 토픽 노드 간 시맨틱 바인딩 (선수 개념, 심화 개념, 연관 분야)**
- [ ] **66. 큐레이션된 학습 로드맵(Curated Learning Path) 네비게이션**
- [ ] **67. 토픽 노드 추가, 편집 및 드래그 앤 드롭 계층 재배치**
- [ ] **68. 순환 참조(Cycle Detection) 방지 알고리즘**
- [ ] **69. 토픽별 난이도 및 완성도 메트릭 집계**
- [ ] **70. 지식 트리 SVG/PNG 내보내기 기능**

---

## PHASE 8: 실시간 토론(Talk), 합의 투표(RFC) & 알림 (71~80)
- [ ] **71. 계층형 스레드 토론(Threaded Talk Page) 데이터 모델 및 API**
- [ ] **72. 토론 상태 머신 (Open, Under Discussion, RFC, Closed, Resolved)**
- [ ] **73. 커뮤니티 합의 투표(RFC Polling) 컴포넌트 (찬성/반대/기권, 정족수 검증)**
- [ ] **74. 토론 댓글 서식 파서 (서명 자동 추가 `~~~~`, 멘션 처리)**
- [ ] **75. 사용자 멘션(@username) 알림 인프라**
- [ ] **76. Server-Sent Events (SSE) 기반 실시간 토론 알림 스트림**
- [ ] **77. 주시문서(Watchlist) 등록 및 변경 발생 시 인앱/웹푸시 알림**
- [ ] **78. 토론 스레드 고정(Pin) 및 잠금(Lock) 관리자 기능**
- [ ] **79. 토론 아카이빙(Talk Archive) 자동 분할 처리기**
- [ ] **80. 분쟁 해결 위원회(Arbitration) 중재 전용 토론 케이스 관리**

---

## PHASE 9: 전문 검색, 자동완성 & 미디어 파이프라인 (81~90)
- [ ] **81. PostgreSQL `tsvector` + 한국어 형태소/영문 형태소 렉시컬 검색 인덱스**
- [ ] **82. `pg_trgm` 기반 표제어 및 내용 유사도 검색 (오타 교정 및 근사 검색)**
- [ ] **83. 디바운스 실시간 표제어 자동완성 모달 (Ctrl+K 퀵 서치)**
- [ ] **84. 네임스페이스, 카테고리, 날짜 범위 패싯(Faceted) 검색 필터**
- [ ] **85. Meilisearch 동기화 파이프라인 (밀리초 단위 고속 검색)**
- [ ] **86. MinIO / S3 미디어 업로드 파이프라인 및 멀티파트 업로더**
- [ ] **87. 이미지 EXIF 메타데이터 자동 제거 및 지오로케이션 스크러빙**
- [ ] **88. 썸네일 동적 생성기 (WebP/AVIF 온디맨드 리사이징)**
- [ ] **89. SVG 보안 살균기 (DOMPurify 기반 악성 스크립트 박멸)**
- [ ] **90. 파일 라이선스(CC BY-SA, Public Domain, 공정이용) 검증 모듈**

---

## PHASE 10: 거버넌스, 안티 반달리즘, 배포 & E2E (91~100)
- [ ] **91. 반달리즘 방지 정규식 필터(Abuse Filter) 엔진 및 실시간 차단**
- [ ] **92. IP 및 CIDR 서브넷 차단, 프록시/VPN 탐지 미들웨어**
- [ ] **93. Special:RecentChanges 실시간 SSE 브로드캐스트 스트림**
- [ ] **94. 법적 권리침해 신고센터 및 30일 임시조치(게시중단) 자동화**
- [ ] **95. 오버사이트(Oversight) 영구 은폐/말소 감사 로그 파이프라인**
- [ ] **96. XML / JSON 위키 덤프 전체 내보내기 & 복원 배치 러너**
- [ ] **97. Vitest 단위 테스트 및 Go API 통합 테스트 스위트 (100% Pass 달성)**
- [ ] **98. Playwright 브라우저 E2E 테스트 스위트 (문서작성-Diff-토픽트리 플로우)**
- [ ] **99. Docker 멀티스테이지 프로덕션 이미지 빌드 및 Nginx 리버스 프록시**
- [ ] **100. 시스템 가동 검증 및 공식 v1.0.0 핸드오버 문서화**
