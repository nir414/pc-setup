package engine

import "strings"

type sectionSpec struct {
	Name       string
	EnvVar     string
	SourceBase string
	DestBase   string
	Folders    []folderSpec
	Matcher    *matcher
}

type folderSpec struct {
	ConfigPath string
	SourcePath string
	DestPath   string
}

type pathPair struct {
	SystemBase string
	RepoBase   string
}

type sectionDescriptor struct {
	EnvVar        string
	RepositoryDir string
}

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
