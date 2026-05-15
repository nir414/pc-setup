# pc-setup 작업 지침

이 저장소는 Windows PC 포맷/신규 셋업을 빠르게 끝내기 위한 개인용 자동화 모음이다.
Claude 가 이 저장소에서 작업할 때 따를 규칙을 정리한다.

## 구성요소

| 파일/디렉토리 | 역할 |
|---|---|
| `scripts/install.ps1` | winget 으로 앱 일괄 설치 |
| `scripts/configure_windows.ps1` | 다크 모드·탐색기 옵션 등 시스템 설정 |
| `scripts/sync.psd1` | 동기화 대상(섹션·폴더·제외 규칙) 정의 |
| `scripts/sync.ps1` | psd1 을 읽어 robocopy 로 백업/복원 실행 |
| `SyncData/<Base>/<RelPath>/` | 실제 백업 데이터 (git 으로 PC 간 동기화) |
| `syncer/`, `syncer.exe` | **deprecated** — 옛 Go 구현, 손대지 말 것 |
| `docs/release/` | 옛 릴리즈 문서 (Go 바이너리용) |

`sync.psd1` 의 구조:

```powershell
Sections = @{
    APPDATA      = @{ Folders = @(...); Excludes = @(...) }
    LOCALAPPDATA = @{ Folders = @(); Excludes = @() }
    USERPROFILE  = @{ Folders = @(...); Excludes = @(...) }
}
```

- `Base` 는 `APPDATA` / `LOCALAPPDATA` / `USERPROFILE` 세 가지만.
- `Folders` 는 각 환경변수 루트 아래 상대 경로.
- `Excludes` 는 폴더 이름 또는 파일 와일드카드 혼용. `sync.ps1::Split-Excludes` 가
  확장자 패턴(`*.log` 등) → robocopy `/XF`, 나머지 → `/XD` 로 자동 분기.

## 수정 체크리스트

`scripts/sync.*` 또는 `SyncData/` 를 손대기 전에:

1. **먼저 읽기**: `scripts/sync.psd1` 과 `scripts/sync.ps1` 둘 다 한 번 읽고 변경이
   기존 의도와 어긋나지 않는지 확인.
2. **psd1 ↔ SyncData 정합성**: `Folders` 에서 항목을 빼면 `SyncData/<Base>/<항목>/`
   디렉토리도 같이 정리. 추가는 반대로.
3. **Excludes 검증**: 새 패턴을 넣었으면 그게 `/XD` 로 갈지 `/XF` 로 갈지 `Split-Excludes`
   로직에 비춰 의도와 일치하는지 확인.
4. **deprecated 영역 금지**: `syncer/`, `syncer.exe`, `docs/release/` 는 새 작업 대상이
   아님. 정리는 사용자 명시 요청 시에만.

## 정책

- **클라우드 의존 금지**: OneDrive·Dropbox 등을 백업 경로로 재도입하지 않는다.
  PC 간 동기화는 `git pull`/`git push` 만으로 끝낸다 (사용자 결정).
- **PowerShell 전용**: 새 도구는 PowerShell 스크립트로. Go·Python 등 별도 빌드가
  필요한 도구는 추가하지 않는다.
- **robocopy 우선**: 파일 복사 로직을 직접 구현하지 말고 robocopy 에 위임한다.
- **⚠️ PC 파일 덮어쓰기 절대 금지**: 이 스크립트는 `Restore` 모드에서 PC 의 파일을
  덮어쓸 수 있다. Claude 는 절대로 사용자 PC 에 `Restore` 모드를 자동 실행하거나,
  `SyncData/` 의 파일을 함부로 수정해서 데이터 손실을 유발하면 안 된다.
  변경 전에 항상 `-WhatIf` 모드(Status)로 미리보기를 해야 한다.
