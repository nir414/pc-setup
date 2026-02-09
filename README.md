# pc-setup (WIP)

SyncData 기반 백업/동기화 구조를 구축하는 중입니다. 문서는 추후 전체 개편 시 정리할 예정입니다.

## 현재 구성

- `SyncData/` : 동기화 대상 실파일
- `sync.toml` : 구조 정의용 메타 데이터


동기화 작업은 2가지 입니다.
 - 백업 데이터 가져오기
	SyncData 에서 폴더 및 파일 구조를 읽어 옵니다.
	SyncData\APPDATA 폴더는 %APPDATA% 경로로 매핑됩니다.
	SyncData\LOCALAPPDATA 폴더는 %LOCALAPPDATA% 경로로 매핑됩니다.
	SyncData\USERPROFILE 폴더는 %USERPROFILE% 경로로 매핑됩니다.
	
	SyncData\backup 폴더를 만들고 내부에 똑같이 데이터를 복사할껀데,
	SyncData 에 존재하는 파일만 복사합니다.
	예시: SyncData\APPDATA 안에서 똑같은 파일을 %APPDATA%에서 찾아 복사해 옵니다.
	
 - 백업 데이터 내보내기
	여기도 먼저 백업 부터 진행합니다.
	위에 백업 데이터 가져오기 작업을 수행합니다.
	그 다음에 로컬 폴더를 덮어 씌웁니다.
	예시: SyncData\APPDATA 안에 있는 파일들을 %APPDATA%에 덮어 씌웁니다.
