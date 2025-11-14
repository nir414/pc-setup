# Windows 개발 환경 설정 자동화

개발 환경을 빠르게 구축하고 설정을 안전하게 동기화할 수 있는 자동화 도구 모음입니다.

## 🚀 빠른 시작

```powershell
# 1. 저장소 클론
git clone https://github.com/nir414/pc-setup.git
cd pc-setup

# 2. 앱 자동 설치
./install.ps1

# 3. 설정 동기화 (백업 권장)
./sync.ps1 -Backup
```

**미리보기:** `./sync.ps1 -DryRun` (실제 변경 없이 확인)

**미리보기:** `./sync.ps1 -DryRun` (실제 변경 없이 확인)

---

## 📂 설정 동기화 상세

### 동기화 대상

| 구분 | 저장소 경로 | → | 시스템 경로 |
|------|-------------|---|-------------|
| 앱 설정 | `AppData/*` | → | `%APPDATA%/*` |
| 사용자 설정 | `USERPROFILE/*` | → | `%USERPROFILE%/*` |

### 백업 제어 (`.syncconfig`)

설정 파일에서 백업 동작을 세밀하게 제어할 수 있습니다:

**제외 패턴** (AppData - 블랙리스트 방식)
```ini
# 전체 백업하되, 아래 패턴만 제외
AppData/Notepad++/backup
AppData/*/cache
*.log
```

**포함 패턴** (USERPROFILE - 화이트리스트 방식)
```ini
# 지정된 경로만 선택적으로 백업 (+ 시작)
+USERPROFILE/Documents/PowerToys
+USERPROFILE/.vscode
```

> **성능 최적화:** USERPROFILE은 대용량 폴더(Documents 등)가 많으므로 포함 패턴 방식을 사용해 필요한 것만 백업합니다.

### 주요 옵션

```powershell
./sync.ps1 -Backup              # 백업 + 동기화 (권장)
./sync.ps1 -DryRun              # 미리보기 (WhatIf)
./sync.ps1 -AppDataOnly         # AppData만 동기화
./sync.ps1 -UserprofileOnly     # USERPROFILE만 동기화
./sync.ps1 -BackupPath "D:\Backup"  # 백업 경로 지정
./sync.ps1 -Force               # 확인 프롬프트 생략
```

### 백업 구조

```
C:\Users\<User>\AppData\backup_20251114_150751\
├── AppData\          # AppData 백업 (제외 패턴 적용)
│   ├── CopyQ\
│   ├── Notepad++\    # backup 폴더는 제외됨
│   └── ...
└── USERPROFILE\      # USERPROFILE 백업 (포함 패턴만)
    └── Documents\
        └── PowerToys\
```

---

## 🗂 시스템 설정 체크리스트

Windows 기본 설정 최적화 (개발 환경 편의성 향상):

| 항목 | 설정 내용 | 경로 |
|------|-----------|------|
| 파일 탐색기 | 확장명/숨김파일 표시, 시작 위치 '이 PC' | 보기 → 옵션 |
| 시작 메뉴 | 설정, 사용자 폴더 아이콘 표시 | 설정 → 개인 설정 → 시작 → 폴더 |
| 마우스 | 포인터 정확도 향상 끄기 | 설정 → Bluetooth 및 장치 → 마우스 |
| 개발자 기능 | 작업 종료(End Task) 메뉴 활성화 | 설정 → 시스템 → 개발자용 |

**팁:** 폴더 보기 설정 후 `보기 → 옵션 → 보기 탭 → 모든 폴더에 적용`으로 일괄 적용

---

## 📦 winget 명령어 참고

### 기본 명령어

```powershell
winget source update                      # 소스 갱신
winget search <키워드>                    # 패키지 검색
winget show --id <ID>                     # 상세 정보
winget install --id <ID>                  # 설치
winget upgrade --all                      # 전체 업데이트
winget list                               # 설치된 패키지
winget uninstall --id <ID>                # 제거
```

### 환경 백업/복원

```powershell
winget export --output .\apps.json        # 설치 목록 내보내기
winget import --import-file .\apps.json   # 다른 PC에서 일괄 설치
```

### 자주 사용하는 예시

```powershell
# VSCode Insiders 설치
winget install --id Microsoft.VisualStudioCode.Insiders `
  --accept-package-agreements --accept-source-agreements

# Git 사용자 스코프 설치
winget install --id Git.Git --scope user `
  --accept-package-agreements --accept-source-agreements
```

### 트러블슈팅

```powershell
# 소스 리셋
winget source reset --force
winget source update

# 관리자 권한 필요 시
# PowerShell을 관리자로 실행 후 재시도
```

---

## 📁 저장소 구조

```text
pc-setup/
├── install.ps1          # winget 기반 앱 자동 설치
├── sync.ps1             # 설정 동기화 (백업 지원)
├── .syncconfig          # 백업 제어 규칙 (제외/포함 패턴)
├── AppData/             # %APPDATA% 설정 파일
│   ├── CopyQ/
│   ├── Everything/
│   ├── FileZilla/
│   ├── Notepad++/
│   └── ...
├── USERPROFILE/         # %USERPROFILE% 설정 파일
│   └── Documents/
│       └── PowerToys/
└── README.md            # 이 문서
```

---

## 💡 주의사항

### 백업 권장

- **첫 동기화 전 반드시 `-Backup` 옵션 사용**
- 백업 폴더: `C:\Users\<User>\AppData\backup_YYYYMMDD_HHMMSS\`
- 실행 중인 앱은 종료 후 동기화

### 민감 정보 제외

이 저장소에 포함하지 말 것:

- `.ssh/` 개인키
- `.gitconfig`의 이메일/이름 (전역 설정)
- API 키, 토큰, 비밀번호
- 라이선스 파일

대안: 암호화된 별도 저장소 또는 클라우드 동기화 사용

---

## 🔗 다음 단계

### VSCode 설정 동기화

Settings Sync 활성화로 확장 및 설정 자동 동기화

### Git 전역 설정

```powershell
git config --global user.name "Your Name"
git config --global user.email "your@email.com"
```

### PowerToys 설정

`USERPROFILE/Documents/PowerToys/Backup/`에 백업 파일 포함됨

### 추가 자동화

필요 시 `bootstrap.ps1` 작성으로 초기 프로비저닝 확장

---

## 📝 라이선스

개인 사용 목적. 필요에 따라 자유롭게 수정하세요.
