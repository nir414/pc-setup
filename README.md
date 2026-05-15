# pc-setup

Windows PC 포맷/신규 셋업 시 앱 설치와 설정 동기화를 빠르게 하기 위한 도구.

## 구조

- `scripts\install.ps1` — winget 으로 앱 일괄 설치
- `scripts\configure_windows.ps1` — 시스템 다크 모드 및 탐색기 옵션 적용
- `scripts\sync.ps1` — 앱 설정 폴더를 저장소의 `SyncData\` 와 동기화
- `scripts\sync.psd1` — 동기화 대상 앱·경로·제외 규칙
- `SyncData\` — 앱별 백업 데이터 (git 으로 버전 관리)
- `syncer\` — Go 기반 백업 도구 소스 코드 (deprecated)
- `docs\release\` — 릴리즈 문서

## 동기화 사용법

설정 백업/복원은 저장소 안 `SyncData\` 폴더에 저장된다. PC 간 동기화는 `git pull` / `git push` 로 처리.

```powershell
# 변경사항 미리보기 (실제 복사 안 함)
.\scripts\sync.ps1 -Mode Status

# PC -> SyncData 백업
.\scripts\sync.ps1 -Mode Backup

# SyncData -> PC 복원 (확인 prompt)
.\scripts\sync.ps1 -Mode Restore

# 단일 앱만
.\scripts\sync.ps1 -Mode Backup -App Notepad++

# 저장소 외 위치 사용
.\scripts\sync.ps1 -Mode Backup -BackupRoot D:\Backup

# 복원 prompt 생략
.\scripts\sync.ps1 -Mode Restore -Force
```

내부적으로 `robocopy` 를 사용한다. `/E /R:1 /W:1` 기본, 제외 규칙은 `scripts\sync.psd1` 의 `ExcludeDirs`(`/XD`), `ExcludeFiles`(`/XF`) 로 전달.

## 동기화 대상 추가/수정

`scripts\sync.psd1` 의 `Apps` 배열에 항목을 추가한다.

```powershell
@{ Name = 'MyApp'; Base = 'APPDATA'; Path = 'MyApp';
   ExcludeDirs = @('cache','logs'); ExcludeFiles = @('*.log') }
```

- `Base`: `APPDATA` (Roaming) / `LOCALAPPDATA` / `USERPROFILE` 중 하나
- `Path`: Base 아래 상대 경로
- `ExcludeDirs` / `ExcludeFiles`: robocopy 패턴 (단순 이름 또는 와일드카드)

## 릴리즈 문서

- `docs\release\RELEASE_NOTES.md`
- `docs\release\RELEASE_CHECKLIST.md`
- `docs\release\RELEASE_COMPLETE.md`
