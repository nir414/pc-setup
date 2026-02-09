# Release v1.0.0

## 개요
Windows PC 설정 파일을 자동으로 백업하는 도구의 첫 번째 정식 릴리즈입니다.

## 주요 기능

### 백업 기능
- `syncer.exe backup`: 시스템 설정 파일을 SyncData 폴더로 백업
- 안전한 백업: 시스템 파일을 덮어쓰지 않고 백업만 수행
- 스냅샷 기능으로 변경사항 추적
- 원자적 파일 쓰기로 안전성 보장

### 상태 확인
- `syncer.exe status`: 백업이 필요한 파일 목록 확인
- 새로 추가된 파일, 수정된 파일, 삭제된 파일 표시

### 설정 관리
- TOML 기반 설정 파일 (`sync.toml`)
- 폴더별 백업 대상 지정
- 제외 규칙 지원 (캐시, 로그 파일 등)

### 지원 환경
- Windows 10/11
- %APPDATA%, %LOCALAPPDATA%, %USERPROFILE% 지원

## 사용 방법

1. `syncer.exe version` - 버전 확인
2. `syncer.exe status` - 백업 필요 파일 확인
3. `syncer.exe backup` - 백업 실행
4. `syncer.exe help` - 도움말 확인

## 기술 스택
- Go 1.22
- github.com/pelletier/go-toml/v2

## 설치 방법
1. `syncer.exe` 파일을 원하는 위치에 다운로드
2. `sync.toml` 설정 파일 준비
3. 명령 프롬프트 또는 PowerShell에서 실행

## 알려진 제한사항
- Windows 전용 (Linux/macOS 미지원)
- 백업만 지원 (복원 기능 없음)

## 라이선스
MIT License
