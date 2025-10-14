# Windows 개발 환경 설정 가이드

## 1. 수동 설정 체크리스트

### 1.1 파일 탐색기

#### 설정 위치

- **보기 설정**: 파일 탐색기 → [보기] 탭 → [옵션] → [폴더 옵션] → [보기]
- **일반 설정**: 파일 탐색기 → [보기] 탭 → [옵션] → [폴더 옵션] → [일반]

#### 변경 항목 - 파일 탐색기

| 항목 | 기본값 | 설정값 | 효과 |
|------|--------|--------|------|
| 파일 확장명 표시 | 숨김 | 표시 | 파일 형식 명확히 확인 |
| 숨김 파일 보기 | 숨김 | 표시 | 시스템 파일 접근 용이 |
| 탐색기 시작 위치 | 빠른 액세스 | 이 PC | 디스크 구조 빠르게 확인 |

#### 체크리스트 - 파일 탐색기

- [x] 파일 확장명 숨기기 해제
- [x] 숨김 파일, 폴더 및 드라이브 표시
- [x] 기본 열기 위치를 '이 PC'로 변경

### 1.2 시작 메뉴 > 폴더

#### 설정 위치 - 시작 메뉴 폴더

설정 > 개인 설정 > 시작 > 폴더

#### 변경 항목 - 시작 메뉴 폴더

| 항목 | 기본값 | 설정값 | 효과 |
|------|--------|--------|------|
| 설정 | 끔 | 켬 | 시작 메뉴에서 설정 바로가기 접근 |
| 개인 폴더 | 끔 | 켬 | 시작 메뉴에서 사용자 폴더 바로가기 접근 |

#### 체크리스트 - 시작 메뉴 폴더

- [x] 설정 폴더 표시
- [x] 개인 폴더 표시

### 1.3 마우스

#### 설정 위치 - 마우스

설정 > Bluetooth 및 장치 > 마우스

#### 변경 항목 - 마우스

| 항목 | 기본값 | 설정값 | 효과 |
|------|--------|--------|------|
| 포인터 정확도 향상 | 켜짐 | 끄기 | 게임/정밀 작업 시 마우스 정밀도 향상 |

#### 체크리스트 - 마우스

- [x] 포인터 정확도 향상 끄기

> 💡 **참고**: 마우스 포인터 속도, 스크롤 설정 등은 개인 선호에 따라 조정하시면 됩니다 (기본값 유지 권장)

### 1.4 시스템 > 개발자용

#### 설정 위치 - 개발자용

설정 > 시스템 > 개발자용

#### 변경 항목 - 개발자용

| 항목 | 기본값 | 설정값 | 효과 |
|------|--------|--------|------|
| 작업 종료 기능 | 끔 | 켬 | 작업 표시줄 우클릭 시 End Task 메뉴 추가 |

#### 체크리스트 - 개발자용

- [x] 작업 종료 기능 활성화

## 2. 주요 변경 사항 요약

### 2.1 핵심 변경사항 (기본값과 다른 항목만)

| 구분 | 기본값 | 변경값 | 효과 |
|------|--------|--------|------|
| 파일 탐색기 시작 | 빠른 액세스 | 이 PC | 드라이브 구조 빠른 확인 |
| 파일 확장명 | 숨김 | 표시 | 파일 형식 명확한 식별 |
| 숨김 파일 | 숨김 | 표시 | 시스템 파일 접근성 |
| 시작 메뉴 설정 폴더 | 끔 | 켬 | 설정 바로가기 접근 |
| 시작 메뉴 개인 폴더 | 끔 | 켬 | 사용자 폴더 바로가기 접근 |
| 포인터 정확도 향상 | 켜짐 | 끄기 | 마우스 정밀도 향상 |
| 작업 종료 기능 | 끔 | 켬 | End Task 메뉴 추가 |

### 2.2 제외된 항목 (Windows 11 기본값 유지)

다음 항목들은 Windows 11에서 기본적으로 활성화되어 있으므로 별도 설정이 불필요합니다:

- **시작 메뉴**: 최근 앱, 자주 사용하는 앱, 추천 파일, 팁/권장사항, 계정 알림 표시
- **마우스**: 기본 단추 설정, 스크롤 방향, 비활성 창 스크롤, 기본 스크롤 줄 수

### 2.3 개발 환경 최적화 포인트

1. **파일 시스템 가시성 향상**: 확장명과 숨김 파일 표시로 개발 파일 관리 용이
2. **프로세스 관리 개선**: End Task 메뉴로 비정상 앱 처리 효율화
3. **탐색 효율성**: '이 PC' 시작으로 프로젝트 디렉터리 접근 단축

## 💡 추가 팁

### 폴더 보기 설정 유지 방법

1. 원하는 폴더 보기 설정 (아이콘 크기, 정렬 방식 등)
2. [보기] → [옵션] → [보기] 탭
3. **모든 폴더에 적용** 클릭
4. 설정이 모든 폴더에 일괄 적용됨

---

## 📦 winget 빠른 매뉴얼

Windows 패키지 매니저 winget으로 앱을 검색하고 설치/업데이트/제거할 수 있습니다. 아래 예시는 PowerShell 기준입니다.

### 1) 기본 준비

```powershell
winget source update   # 패키지 카탈로그 최신화
winget --version       # winget 버전 확인
```

### 2) 검색과 상세 보기

```powershell
winget search mybox            # 키워드로 검색 (예: 네이버 MYBOX)
winget show NAVER.MYBOX        # 정확한 ID로 상세 정보/버전/설치 지침 확인
winget show --id NAVER.MYBOX --versions  # 사용 가능한 버전 목록
```

### 3) 설치(Install)

```powershell
# 표준 설치
winget install --id NAVER.MYBOX --accept-package-agreements --accept-source-agreements

# 특정 버전으로 설치 (카탈로그에 버전이 있을 때)
winget install --id Microsoft.VisualStudioCode --version 1.92.0 \
  --accept-package-agreements --accept-source-agreements

# 채널이 있는 패키지라면 (예: Insiders/Canary/Preview 등)
winget install --id Microsoft.VisualStudioCode --channel insiders \
  --accept-package-agreements --accept-source-agreements

# 사용자 스코프(현재 사용자만) 설치가 필요할 때
winget install --id Git.Git --scope user \
  --accept-package-agreements --accept-source-agreements
```

### 4) 업그레이드(Upgrade)

```powershell
winget upgrade                          # 업데이트 가능한 앱 목록
winget upgrade --all \
  --accept-package-agreements --accept-source-agreements
winget upgrade --id Microsoft.PowerToys \
  --accept-package-agreements --accept-source-agreements
```

### 5) 제거(Uninstall)와 목록(List)

```powershell
winget list                 # 설치된 앱 목록 (winget이 추적 가능한 항목)
winget uninstall --id NAVER.MYBOX
```

### 6) 내 설치 목록 내보내기/가져오기

```powershell
# 현재 설치 목록을 JSON으로 내보내기
winget export --output .\apps.json --include-versions

# 다른 PC에서 같은 환경으로 설치
winget import --import-file .\apps.json \
  --accept-package-agreements --accept-source-agreements
```

### 7) 문제 해결 트러블슈팅

- ID가 모호할 때: `winget show --id <정확한ID>`로 확인 후 사용
- 알파/프리릴리즈가 필요할 때: `--channel` 또는 `--version` 지원 여부를 `winget show`로 확인
- 관리자 권한 필요 설치: PowerShell을 관리자 권한으로 실행 후 재시도
- 소스 오류: `winget source reset --force` 후 `winget source update`
- 네트워크 정책/스토어 로그인 요구: Microsoft Store 로그인 후 재시도

### 8) 이 저장소 스크립트 실행

```powershell
winget source update
./install.ps1
```

권장: PowerShell(Windows Terminal)에서 실행하고, 필요한 경우 관리자 권한으로 실행하세요.
