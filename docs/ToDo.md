# HamaShell 개발 로드맵 (Korean)

프로젝트 아키텍처와 현재 코드베이스를 기반으로 한 단계별 개발 계획:

## 🎯 1단계: 소규모 목표 - 기본 CLI 구조
**우선순위: 높음** - 기반 설정
- [x] viper 설정 로딩 수정 및 example.yaml로 테스트
- [x] 기본 start 명령어 골격 구현
- [x] start 명령어에 대한 테스트 
- [x] start 명령어에 대한 리팩토링 
- [ ] 기본 stop 명령어 골격 구현  
- [ ] 기본 status 명령어 골격 구현
- [ ] `pkg/types/config.go`에 설정 타입 생성

## 🔧 2단계: 개선 - 핵심 컴포넌트 
**우선순위: 중간** - 핵심 기능
- [ ] `internal/core/config/loader.go`에 설정 로더 구현
- [ ] `internal/core/config/validator.go`에 설정 검증기 구현
- [ ] `pkg/types/session.go`에 세션 타입 생성
- [ ] `internal/core/session/manager.go`에 기본 세션 관리자 구현

## 🚀 3단계: 중간 목표 - 서비스 레이어
**우선순위: 중간** - 비즈니스 로직 통합
- [ ] `internal/service/config_service.go`에 설정 서비스 구현
- [ ] `internal/service/session_service.go`에 세션 서비스 구현
- [ ] CLI 명령어와 서비스 연결
- [ ] project.stage.service 경로 해석 구현
- [ ] 명령어를 위한 기본 프로세스 실행 추가

## 🌟 4단계: 고급 기능 - 연결 관리
**우선순위: 낮음** - 고급 기능
- [ ] `internal/core/connection/ssh.go`에 SSH 클라이언트 구현
- [ ] 포트 포워딩을 위한 터널 관리자 구현
- [ ] 세션 지속성 및 복구 추가
- [ ] 연결을 위한 상태 모니터링 구현
- [ ] 터미널 멀티플렉서 통합 (tmux/zellij) 추가
- [ ] 대화형 대시보드/TUI 모드 구현

## ✅ 5단계: 최종 검사 및 테스트
**우선순위: 낮음** - 프로덕션 준비
- [ ] 모든 컴포넌트에 대한 포괄적인 단위 테스트 작성
- [ ] CLI 명령어에 대한 통합 테스트 작성
- [ ] 셸 자동완성 스크립트 추가 (bash/zsh/fish)
- [ ] 성능 테스트 및 최적화
- [ ] 크로스 플랫폼 테스트 (Linux/macOS/Windows)
- [ ] 문서화 및 예제
- [ ] 릴리스 준비 및 CI/CD 설정

---

# HamaShell Development Roadmap (English)

Progressive development plan based on project architecture and current codebase:

## 🎯 Phase 1: Small Goals - Basic CLI Structure
**Priority: HIGH** - Foundation setup
- [ ] Fix viper config loading and test with example.yaml
- [ ] Implement basic start command skeleton  
- [ ] Implement basic stop command skeleton
- [ ] Implement basic status command skeleton
- [ ] Create config types in `pkg/types/config.go`

## 🔧 Phase 2: Improvements - Core Components  
**Priority: MEDIUM** - Core functionality
- [ ] Implement config loader in `internal/core/config/loader.go`
- [ ] Implement config validator in `internal/core/config/validator.go`
- [ ] Create session types in `pkg/types/session.go`
- [ ] Implement basic session manager in `internal/core/session/manager.go`

## 🚀 Phase 3: Middle Goals - Service Layer
**Priority: MEDIUM** - Business logic integration
- [ ] Implement config service in `internal/service/config_service.go`
- [ ] Implement session service in `internal/service/session_service.go`
- [ ] Connect CLI commands to services
- [ ] Implement project.stage.service path resolution
- [ ] Add basic process execution for commands

## 🌟 Phase 4: Advanced Features - Connection Management
**Priority: LOW** - Advanced functionality
- [ ] Implement SSH client in `internal/core/connection/ssh.go`
- [ ] Implement tunnel manager for port forwarding
- [ ] Add session persistence and recovery
- [ ] Implement health monitoring for connections
- [ ] Add terminal multiplexer integration (tmux/zellij)
- [ ] Implement interactive dashboard/TUI mode

## ✅ Phase 5: Final Checks and Tests
**Priority: LOW** - Production readiness
- [ ] Write comprehensive unit tests for all components
- [ ] Write integration tests for CLI commands
- [ ] Add shell completion scripts (bash/zsh/fish)
- [ ] Performance testing and optimization
- [ ] Cross-platform testing (Linux/macOS/Windows)
- [ ] Documentation and examples
- [ ] Release preparation and CI/CD setup

This follows the architecture's layered approach: CLI → Service → Core → Infrastructure, building incrementally from basic structure to advanced features.

---

## TODO: Implementation Tasks