# pc-setup

Windows PC 설정 파일을 자동으로 백업하는 도구입니다.

[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/nir414/pc-setup/releases)

## 현재 구성

- `syncer.exe` : 실행 파일 (Windows 전용)
- `SyncData/` : 백업된 설정 파일들
- `sync.toml` : 백업 대상 폴더 및 제외 규칙 정의
- `syncer/` : 백업 도구 소스 코드 (Go)

## 빠른 시작

릴리즈 페이지에서 `syncer.exe`를 다운로드하고 바로 사용하세요!

```powershell
# 버전 확인
.\syncer.exe version

# 도움말 보기
.\syncer.exe help
```

## 사용법

### 0. 실행 파일 다운로드 (권장)

[릴리즈 페이지](https://github.com/nir414/pc-setup/releases)에서 최신 `syncer.exe`를 다운로드하세요.

### 1. 프로그램 빌드 (선택사항)

```powershell
cd syncer
go build -o ../syncer.exe ./cmd/syncer
cd ..
```

### 2. 명령어

**버전 확인:**
```powershell
.\syncer.exe version
```

**백업 전 상태 확인:**
```powershell
.\syncer.exe status
```
어떤 파일들이 변경되었는지 확인합니다.

**백업 실행:**
```powershell
.\syncer.exe backup
```
시스템 파일을 SyncData 폴더로 백업합니다.

**도움말:**
```powershell
.\syncer.exe help
```

## 작동 방식

### 백업 (backup)

1. `sync.toml`에 정의된 폴더 구조를 읽습니다
2. **SyncData 폴더**에서 백업할 파일 목록을 파악합니다 (템플릿)
3. 시스템(%APPDATA% 등)에서 해당 파일을 찾습니다
4. 변경된 파일을 **SyncData\backup** 폴더로 복사합니다
5. 현재 상태를 스냅샷으로 저장합니다 (`.syncer/state.json`)

**경로 구조:**
- `SyncData\APPDATA` → 백업할 파일 목록 (템플릿)
- `%APPDATA%` → 시스템 원본 파일
- `SyncData\backup\APPDATA` → 실제 백업 저장소

**예시:**
- `SyncData\APPDATA\Notepad++\config.xml` 파일이 있으면 (템플릿)
- `%APPDATA%\Notepad++\config.xml`에서 최신 버전을 찾아
- `SyncData\backup\APPDATA\Notepad++\config.xml`로 복사합니다

### 상태 확인 (status)

백업이 필요한 파일 목록을 보여줍니다:
- **새로 추가됨**: 시스템에만 존재하는 파일
- **수정됨**: 시스템에서 변경된 파일
- **삭제됨**: 시스템에서 삭제된 파일
- **SyncData에만 존재**: 시스템에는 없는 파일

## 안전 기능

- ✅ 백업만 지원 (시스템 파일 덮어쓰기 없음)
- ✅ 스냅샷 기능으로 변경사항 추적
- ✅ 제외 규칙 지원 (캐시, 로그 파일 등 제외)
- ✅ 원자적 파일 쓰기 (임시 파일 사용)

## 설정 파일 (sync.toml)

```toml
[SyncData.APPDATA]
folders = [
    "Notepad++/",
    "FileZilla/"
]
excludes = [
    "*/cache/",
    "*.log"
]
```

- `folders`: 백업할 폴더 목록
- `excludes`: 제외할 파일 패턴

