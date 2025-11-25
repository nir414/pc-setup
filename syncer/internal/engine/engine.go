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
type Options struct {
	Root          string
	Config        *config.Config
	SnapshotStore state.Store
	Logger        *log.Logger
}

// Engine orchestrates backup and synchronization operations.
type Engine struct {
	root      string
	cfg       *config.Config
	store     state.Store
	logger    *log.Logger
	targets   []sectionSpec
	pathIndex map[string]pathPair
}

// BackupResult captures statistics from a backup run.
type BackupResult struct {
	CopiedFiles  int
	SkippedFiles int
	CopiedBytes  int64
	RemovedFiles int
}

// SyncResult captures statistics from a sync run.
type SyncResult struct {
	UpdatedFiles int
	UpdatedBytes int64
	RemovedFiles int
	SkippedFiles int
}

// StatusReport summarises the current difference between system and repository.
type StatusReport struct {
	GeneratedAt time.Time
	Summary     StatusSummary
	Entries     []DiffEntry
}

// StatusSummary aggregates counts for diff categories.
type StatusSummary struct {
	UpToDate    int
	NeedsBackup int
	NeedsSync   int
	Conflicts   int
}

// DiffStatus categorises a difference between system and repository content.
type DiffStatus string

// Diff status values.
const (
	DiffStatusUpToDate       DiffStatus = "up_to_date"
	DiffStatusSystemAdded    DiffStatus = "system_added"
	DiffStatusSystemModified DiffStatus = "system_modified"
	DiffStatusSystemDeleted  DiffStatus = "system_deleted"
	DiffStatusRepoAdded      DiffStatus = "repo_added"
	DiffStatusRepoModified   DiffStatus = "repo_modified"
	DiffStatusRepoDeleted    DiffStatus = "repo_deleted"
	DiffStatusConflict       DiffStatus = "conflict"
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

		destBase := filepath.Join(e.root, "SyncData", descriptor.RepositoryDir)
		matcher := newMatcher(section.Excludes)

		folders := make([]folderSpec, 0, len(section.Folders))
		for _, folder := range section.Folders {
			normalized := normaliseFolder(folder)
			if normalized == "" {
				continue
			}
			folderInfo := folderSpec{
				ConfigPath: normalized,
				SourcePath: filepath.Join(sourceBase, normalized),
				DestPath:   filepath.Join(destBase, normalized),
			}
			folders = append(folders, folderSpec{
				ConfigPath: normalized,
				SourcePath: folderInfo.SourcePath,
				DestPath:   folderInfo.DestPath,
			})

			prefix := makeKey(descriptor.RepositoryDir, normalized)
			index[prefix] = pathPair{
				SystemBase: folderInfo.SourcePath,
				RepoBase:   folderInfo.DestPath,
			}
		}

		spec := sectionSpec{
			Name:       descriptor.RepositoryDir,
			EnvVar:     descriptor.EnvVar,
			SourceBase: sourceBase,
			DestBase:   destBase,
			Folders:    folders,
			Matcher:    matcher,
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
func (e *Engine) Backup(ctx context.Context) (*BackupResult, error) {
	_, diff, err := e.computeDiff(ctx)
	if err != nil {
		return nil, err
	}

	stats := &BackupResult{}

	for _, entry := range diff.Entries {
		switch entry.Status {
		case DiffStatusSystemAdded, DiffStatusSystemModified:
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
		case DiffStatusSystemDeleted:
			if entry.RepoPath == "" {
				continue
			}
			if err := os.Remove(entry.RepoPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("remove %s: %w", entry.Path, err)
			}
			stats.RemovedFiles++
		case DiffStatusConflict:
			stats.SkippedFiles++
		default:
			// changes owned by repo, skip during backup
		}
	}

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
		case DiffStatusSystemAdded, DiffStatusSystemModified, DiffStatusSystemDeleted:
			report.Summary.NeedsBackup++
		case DiffStatusRepoAdded, DiffStatusRepoModified, DiffStatusRepoDeleted:
			report.Summary.NeedsSync++
		case DiffStatusConflict:
			report.Summary.Conflicts++
		}
	}

	// remove entries marked up-to-date from listing to keep output concise
	pruned := report.Entries[:0]
	for _, entry := range report.Entries {
		if entry.Status == DiffStatusUpToDate {
			continue
		}
		pruned = append(pruned, entry)
	}
	report.Entries = pruned

	return report, nil
}

// Sync applies repository changes to the system.
func (e *Engine) Sync(ctx context.Context) (*SyncResult, error) {
	_, diff, err := e.computeDiff(ctx)
	if err != nil {
		return nil, err
	}
	stats := &SyncResult{}

	for _, entry := range diff.Entries {
		switch entry.Status {
		case DiffStatusRepoAdded, DiffStatusRepoModified:
			if entry.SystemPath == "" || entry.RepoPath == "" {
				continue
			}
			if err := e.copyFile(entry.RepoPath, entry.SystemPath); err != nil {
				return nil, fmt.Errorf("sync copy %s: %w", entry.Path, err)
			}
			stats.UpdatedFiles++
			if entry.Repo != nil {
				stats.UpdatedBytes += entry.Repo.Size
			}
		case DiffStatusRepoDeleted:
			if entry.SystemPath == "" {
				continue
			}
			if err := os.Remove(entry.SystemPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("sync remove %s: %w", entry.Path, err)
			}
			stats.RemovedFiles++
		case DiffStatusConflict, DiffStatusSystemAdded, DiffStatusSystemModified, DiffStatusSystemDeleted:
			stats.SkippedFiles++
		}
	}

	freshSnapshot, err := e.collectSystemSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	if err := e.store.Save(ctx, freshSnapshot); err != nil {
		return nil, fmt.Errorf("save snapshot: %w", err)
	}

	return stats, nil
}

func (e *Engine) computeDiff(ctx context.Context) (*state.Snapshot, *diffResult, error) {
	snapshot, err := e.store.Load(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("load snapshot: %w", err)
	}

	systemFiles, err := e.collectSystemFiles(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("collect system files: %w", err)
	}

	repoFiles, err := e.collectRepoFiles(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("collect repo files: %w", err)
	}

	diff := buildDiff(systemFiles, repoFiles, snapshot)
	for i := range diff.Entries {
		sysPath, repoPath, ok := e.resolvePaths(diff.Entries[i].Path)
		if diff.Entries[i].System != nil {
			diff.Entries[i].SystemPath = diff.Entries[i].System.AbsPath
		} else if ok {
			diff.Entries[i].SystemPath = sysPath
		}
		if diff.Entries[i].Repo != nil {
			diff.Entries[i].RepoPath = diff.Entries[i].Repo.AbsPath
		} else if ok {
			diff.Entries[i].RepoPath = repoPath
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
