package engine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nir414/pc-setup/syncer/internal/config"
	"github.com/nir414/pc-setup/syncer/internal/state"
)

// Options configures the sync engine.
// Engine 설정 옵션
type Options struct {
	Root          string         // 프로젝트 루트 경로
	Config        *config.Config // 설정 파일 내용
	SnapshotStore state.Store    // 스냅샷 저장소
	Logger        *log.Logger    // 로거
}

// Engine orchestrates backup and synchronization operations.
// 백업 작업을 조율하는 핵심 엔진
type Engine struct {
	root      string              // 프로젝트 루트 경로
	cfg       *config.Config      // 설정
	store     state.Store         // 스냅샷 저장소
	logger    *log.Logger         // 로거
	targets   []sectionSpec       // 백업 대상 섹션들 (APPDATA, LOCALAPPDATA 등)
	pathIndex map[string]pathPair // 경로 매핑 인덱스
}

// BackupResult captures statistics from a backup run.
type BackupResult struct {
	CopiedFiles  int
	SkippedFiles int
	CopiedBytes  int64
	RemovedFiles int
}

// StatusReport summarises the current difference between system and repository.
type StatusReport struct {
	GeneratedAt time.Time
	Summary     StatusSummary
	Entries     []DiffEntry
}

// StatusSummary aggregates counts for diff categories.
type StatusSummary struct {
	UpToDate    int // 이미 최신 상태
	NeedsBackup int // 백업 필요 (SyncData에 정의된 파일 중 변경/미백업/삭제됨)
	Ignored     int // 무시됨 (시스템에만 있고 SyncData에 정의되지 않음)
}

// DiffStatus categorises a difference between system and repository content.
type DiffStatus string

// Diff status values.
const (
	DiffStatusUpToDate       DiffStatus = "up_to_date"      // 시스템과 SyncData가 동일
	DiffStatusSystemAdded    DiffStatus = "system_added"    // 시스템에만 존재
	DiffStatusSystemModified DiffStatus = "system_modified" // 시스템에서 변경됨
	DiffStatusSystemDeleted  DiffStatus = "system_deleted"  // 시스템에서 삭제됨
	DiffStatusRepoOnly       DiffStatus = "repo_only"       // SyncData에만 존재 (시스템에 없음)
)

// DiffEntry describes the state of a single logical file.
type DiffEntry struct {
	Path       string
	Status     DiffStatus
	System     *FileInfo
	Repo       *FileInfo
	SystemPath string
	RepoPath   string
}

// FileInfo represents a tracked file instance.
type FileInfo struct {
	Path    string
	AbsPath string
	Size    int64
	ModTime time.Time
	Hash    string
}

// New constructs an Engine from the provided options.
func New(opts Options) *Engine {
	root := opts.Root
	if root == "" {
		root, _ = os.Getwd()
	}

	logger := opts.Logger
	if logger == nil {
		logger = log.New(io.Discard, "", log.LstdFlags)
	}

	e := &Engine{
		root:   root,
		cfg:    opts.Config,
		store:  opts.SnapshotStore,
		logger: logger,
	}
	sections, index := e.buildTargets()
	e.targets = sections
	e.pathIndex = index
	return e
}

func (e *Engine) buildTargets() ([]sectionSpec, map[string]pathPair) {
	if e.cfg == nil {
		return nil, make(map[string]pathPair)
	}

	sections := make([]sectionSpec, 0, len(e.cfg.SyncData))
	index := make(map[string]pathPair)
	for name, section := range e.cfg.SyncData {
		descriptor, ok := knownSections[strings.ToUpper(name)]
		if !ok {
			e.logger.Printf("warning: unsupported section %q ignored", name)
			continue
		}

		sourceBase := os.Getenv(descriptor.EnvVar)
		if sourceBase == "" {
			e.logger.Printf("warning: environment variable %s not set; skipping section %s", descriptor.EnvVar, name)
			continue
		}

		// 템플릿 경로 (어떤 파일을 백업할지 정의)
		templateBase := filepath.Join(e.root, "SyncData", descriptor.RepositoryDir)
		// 실제 백업 저장 경로
		destBase := filepath.Join(e.root, "SyncData", "backup", descriptor.RepositoryDir)
		matcher := newMatcher(section.Excludes)

		folders := make([]folderSpec, 0, len(section.Folders))
		for _, folder := range section.Folders {
			normalized := normaliseFolder(folder)
			if normalized == "" {
				continue
			}
			folderInfo := folderSpec{
				ConfigPath:   normalized,
				SourcePath:   filepath.Join(sourceBase, normalized),
				TemplatePath: filepath.Join(templateBase, normalized),
				DestPath:     filepath.Join(destBase, normalized),
			}
			folders = append(folders, folderSpec{
				ConfigPath:   normalized,
				SourcePath:   folderInfo.SourcePath,
				TemplatePath: folderInfo.TemplatePath,
				DestPath:     folderInfo.DestPath,
			})

			prefix := makeKey(descriptor.RepositoryDir, normalized)
			index[prefix] = pathPair{
				SystemBase: folderInfo.SourcePath,
				RepoBase:   folderInfo.DestPath,
			}
		}

		spec := sectionSpec{
			Name:         descriptor.RepositoryDir,
			EnvVar:       descriptor.EnvVar,
			SourceBase:   sourceBase,
			TemplateBase: templateBase,
			DestBase:     destBase,
			Folders:      folders,
			Matcher:      matcher,
		}

		sections = append(sections, spec)
	}

	sort.Slice(sections, func(i, j int) bool { return sections[i].Name < sections[j].Name })

	// ensure section root prefixes exist even if folders empty
	for _, section := range sections {
		prefix := makeKey(section.Name, "")
		if _, exists := index[prefix]; !exists {
			index[prefix] = pathPair{
				SystemBase: section.SourceBase,
				RepoBase:   section.DestBase,
			}
		}
	}

	return sections, index
}

// Backup synchronises files from the system into the repository.
// SyncData 템플릿에 정의된 파일만 시스템에서 찾아 백업합니다.
func (e *Engine) Backup(ctx context.Context) (*BackupResult, error) {
	_, diff, err := e.computeDiff(ctx)
	if err != nil {
		return nil, err
	}

	stats := &BackupResult{}

	for _, entry := range diff.Entries {
		switch entry.Status {
		case DiffStatusSystemModified:
			// SyncData에 정의된 파일이 시스템에서 수정됨 -> 백업
			if entry.SystemPath == "" || entry.RepoPath == "" {
				continue
			}
			if err := e.copyFile(entry.SystemPath, entry.RepoPath); err != nil {
				return nil, fmt.Errorf("copy %s: %w", entry.Path, err)
			}
			stats.CopiedFiles++
			if entry.System != nil {
				stats.CopiedBytes += entry.System.Size
			}

		case DiffStatusRepoOnly:
			// SyncData에 정의된 파일 -> 시스템에서 찾아서 백업
			if entry.SystemPath == "" || entry.RepoPath == "" {
				stats.SkippedFiles++
				continue
			}
			// 시스템에 파일이 존재하는지 확인
			if _, err := os.Stat(entry.SystemPath); err != nil {
				if os.IsNotExist(err) {
					stats.SkippedFiles++
					continue
				}
				return nil, fmt.Errorf("stat %s: %w", entry.Path, err)
			}
			// 백업 복사
			if err := e.copyFile(entry.SystemPath, entry.RepoPath); err != nil {
				return nil, fmt.Errorf("copy %s: %w", entry.Path, err)
			}
			stats.CopiedFiles++
			// 파일 크기 계산
			if info, err := os.Stat(entry.SystemPath); err == nil {
				stats.CopiedBytes += info.Size()
			}

		case DiffStatusSystemDeleted:
			// 시스템에서 삭제된 파일 -> backup에서도 삭제
			if entry.RepoPath == "" {
				continue
			}
			if err := os.Remove(entry.RepoPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("remove %s: %w", entry.Path, err)
			}
			stats.RemovedFiles++

		case DiffStatusSystemAdded:
			// 시스템에만 있는 파일 (SyncData에 정의되지 않음) -> 무시
			stats.SkippedFiles++

		default:
			// DiffStatusUpToDate - 이미 최신 상태
		}
	}

	// 현재 시스템 상태를 스냅샷으로 저장
	freshSnapshot, err := e.collectSystemSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	if err := e.store.Save(ctx, freshSnapshot); err != nil {
		return nil, fmt.Errorf("save snapshot: %w", err)
	}

	return stats, nil
}

// Status computes a status report describing pending changes.
// 백업이 필요한 변경사항을 보고합니다.
func (e *Engine) Status(ctx context.Context) (*StatusReport, error) {
	_, diff, err := e.computeDiff(ctx)
	if err != nil {
		return nil, err
	}

	report := &StatusReport{
		GeneratedAt: time.Now(),
		Entries:     diff.Entries,
	}

	for _, entry := range diff.Entries {
		switch entry.Status {
		case DiffStatusUpToDate:
			report.Summary.UpToDate++
		case DiffStatusSystemModified, DiffStatusRepoOnly, DiffStatusSystemDeleted:
			// SyncData에 정의된 파일이 백업 필요
			report.Summary.NeedsBackup++
		case DiffStatusSystemAdded:
			// 시스템에만 있는 파일 (무시)
			report.Summary.Ignored++
		}
	}

	// up-to-date와 ignored 항목은 출력에서 제외
	pruned := report.Entries[:0]
	for _, entry := range report.Entries {
		if entry.Status == DiffStatusUpToDate || entry.Status == DiffStatusSystemAdded {
			continue
		}
		pruned = append(pruned, entry)
	}
	report.Entries = pruned

	return report, nil
}

func (e *Engine) computeDiff(ctx context.Context) (*state.Snapshot, *diffResult, error) {
	snapshot, err := e.store.Load(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("load snapshot: %w", err)
	}

	// SyncData 템플릿에서 백업할 파일 목록만 수집
	templateFiles, err := e.collectTemplateFiles(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("collect template files: %w", err)
	}

	// 템플릿에 정의된 파일의 시스템/백업 상태 수집
	systemFiles := make(fileMap)
	backupFiles := make(fileMap)

	for key := range templateFiles {
		systemPath, backupPath, ok := e.resolvePaths(key)
		if !ok {
			continue
		}

		// 시스템 파일 확인
		if info, err := os.Stat(systemPath); err == nil && !info.IsDir() {
			hash, err := hashFile(systemPath)
			if err != nil {
				return nil, nil, fmt.Errorf("hash %s: %w", systemPath, err)
			}
			systemFiles[key] = &FileInfo{
				Path:    key,
				AbsPath: systemPath,
				Size:    info.Size(),
				ModTime: info.ModTime().UTC(),
				Hash:    hash,
			}
		}

		// 백업 파일 확인 (SyncData/backup)
		if info, err := os.Stat(backupPath); err == nil && !info.IsDir() {
			hash, err := hashFile(backupPath)
			if err != nil {
				return nil, nil, fmt.Errorf("hash %s: %w", backupPath, err)
			}
			backupFiles[key] = &FileInfo{
				Path:    key,
				AbsPath: backupPath,
				Size:    info.Size(),
				ModTime: info.ModTime().UTC(),
				Hash:    hash,
			}
		}
	}

	// 시스템 파일과 백업 파일 비교
	diff := buildDiff(systemFiles, backupFiles, snapshot)
	for i := range diff.Entries {
		sysPath, backupPath, ok := e.resolvePaths(diff.Entries[i].Path)
		if diff.Entries[i].System != nil {
			diff.Entries[i].SystemPath = diff.Entries[i].System.AbsPath
		} else if ok {
			diff.Entries[i].SystemPath = sysPath
		}
		// RepoPath는 실제 백업 저장 경로(SyncData/backup)를 가리킴
		if diff.Entries[i].Repo != nil {
			// 템플릿 경로를 백업 경로로 변환
			diff.Entries[i].RepoPath = strings.Replace(diff.Entries[i].Repo.AbsPath,
				filepath.Join(e.root, "SyncData")+string(filepath.Separator),
				filepath.Join(e.root, "SyncData", "backup")+string(filepath.Separator), 1)
		} else if ok {
			diff.Entries[i].RepoPath = backupPath
		}
	}
	return snapshot, diff, nil
}

// copyFile copies a file from src to dst, creating parent directories as needed.
func (e *Engine) copyFile(src, dst string) error {
	if src == "" || dst == "" {
		return fmt.Errorf("invalid copy source (%q) or destination (%q)", src, dst)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	return copyFileContents(src, dst)
}

type diffResult struct {
	Entries []DiffEntry
}

func (e *Engine) resolvePaths(key string) (string, string, bool) {
	prefix := key
	for {
		if pair, ok := e.pathIndex[prefix]; ok {
			remainder := strings.TrimPrefix(key, prefix)
			remainder = strings.TrimPrefix(remainder, "/")
			if remainder == "" {
				return pair.SystemBase, pair.RepoBase, true
			}
			return filepath.Join(pair.SystemBase, filepath.FromSlash(remainder)),
				filepath.Join(pair.RepoBase, filepath.FromSlash(remainder)), true
		}
		idx := strings.LastIndex(prefix, "/")
		if idx < 0 {
			break
		}
		prefix = prefix[:idx]
	}
	return "", "", false
}
