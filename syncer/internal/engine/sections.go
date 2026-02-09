// sections.go - 섹션 정의 및 경로 매핑
// APPDATA, LOCALAPPDATA, USERPROFILE 등의 환경 변수를 SyncData 경로로 매핑합니다.
package engine

import "strings"

// sectionSpec: 각 섹션(APPDATA 등)의 백업 대상 정보
type sectionSpec struct {
	Name         string       // 섹션 이름 (APPDATA, LOCALAPPDATA 등)
	EnvVar       string       // 환경 변수 이름
	SourceBase   string       // 시스템 경로 (%APPDATA% 등)
	TemplateBase string       // SyncData 템플릿 경로 (파일 목록)
	DestBase     string       // SyncData/backup 백업 경로 (실제 백업)
	Folders      []folderSpec // 백업할 폴더 목록
	Matcher      *matcher     // 제외 패턴 매처
}

// folderSpec: 백업할 개별 폴더 정보
type folderSpec struct {
	ConfigPath   string // 설정 파일에서의 경로
	SourcePath   string // 시스템 실제 경로
	TemplatePath string // SyncData 템플릿 경로
	DestPath     string // SyncData/backup 백업 경로
}

// pathPair: 시스템 경로와 SyncData 경로의 쌍
type pathPair struct {
	SystemBase string
	RepoBase   string
}

// sectionDescriptor: 지원하는 섹션의 환경 변수 정의
type sectionDescriptor struct {
	EnvVar        string // 환경 변수 이름 (예: APPDATA)
	RepositoryDir string // SyncData 내 디렉토리 이름
}

// knownSections: 지원하는 Windows 환경 변수 목록
var knownSections = map[string]sectionDescriptor{
	"APPDATA": {
		EnvVar:        "APPDATA",
		RepositoryDir: "APPDATA",
	},
	"LOCALAPPDATA": {
		EnvVar:        "LOCALAPPDATA",
		RepositoryDir: "LOCALAPPDATA",
	},
	"USERPROFILE": {
		EnvVar:        "USERPROFILE",
		RepositoryDir: "USERPROFILE",
	},
}

func normaliseFolder(folder string) string {
	folder = strings.TrimSpace(folder)
	if folder == "" {
		return folder
	}
	folder = strings.Trim(folder, "\\/")
	if folder == "" {
		return folder
	}
	return folder
}
