// collect.go - 파일 수집 모듈
// 시스템 및 SyncData에서 파일 정보를 수집하고 해시를 계산합니다.
package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nir414/pc-setup/syncer/internal/state"
)

type fileMap map[string]*FileInfo

// collectSystemFiles: 시스템 (%APPDATA%, %LOCALAPPDATA% 등)에서 파일 정보 수집
func (e *Engine) collectSystemFiles(ctx context.Context) (fileMap, error) {
	result := make(fileMap)
	for _, section := range e.targets {
		for _, folder := range section.Folders {
			if err := e.collectFolder(ctx, section, folder, folder.SourcePath, result); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

// collectTemplateFiles: SyncData 폴더(템플릿)에서 백업할 파일 목록 수집
func (e *Engine) collectTemplateFiles(ctx context.Context) (fileMap, error) {
	result := make(fileMap)
	for _, section := range e.targets {
		for _, folder := range section.Folders {
			if err := e.collectFolder(ctx, section, folder, folder.TemplatePath, result); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

func (e *Engine) collectFolder(ctx context.Context, section sectionSpec, folder folderSpec, base string, dest fileMap) error {
	if base == "" {
		return nil
	}

	info, err := os.Stat(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}

	return filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		rel, relErr := filepath.Rel(base, path)
		if relErr != nil {
			return relErr
		}
		sectionRelative := combineSectionPath(folder.ConfigPath, rel)
		if section.Matcher != nil && section.Matcher.ShouldSkip(sectionRelative, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		fileInfo, statErr := d.Info()
		if statErr != nil {
			return statErr
		}

		hash, err := hashFile(path)
		if err != nil {
			return err
		}

		key := makeKey(section.Name, sectionRelative)
		dest[key] = &FileInfo{
			Path:    key,
			AbsPath: path,
			Size:    fileInfo.Size(),
			ModTime: fileInfo.ModTime().UTC(),
			Hash:    hash,
		}
		return nil
	})
}

// collectSystemSnapshot: 현재 시스템 파일 상태를 스냅샷으로 생성
// 다음 백업 시 변경사항 추적에 사용됩니다.
func (e *Engine) collectSystemSnapshot(ctx context.Context) (*state.Snapshot, error) {
	files, err := e.collectSystemFiles(ctx)
	if err != nil {
		return nil, err
	}
	snapshot := state.NewSnapshot()
	for key, info := range files {
		snapshot.Files[key] = state.FileRecord{
			Hash:    info.Hash,
			Size:    info.Size,
			ModTime: info.ModTime,
		}
	}
	snapshot.GeneratedAt = time.Now().UTC()
	return snapshot, nil
}

func combineSectionPath(folderConfig, rel string) string {
	rel = toForwardSlashes(rel)
	if rel == "." || rel == "" {
		rel = ""
	}
	base := toForwardSlashes(strings.Trim(folderConfig, "\\/"))
	switch {
	case base == "" && rel == "":
		return ""
	case base == "":
		return rel
	case rel == "":
		return base
	default:
		return base + "/" + rel
	}
}

func makeKey(sectionName, sectionRelative string) string {
	sectionRelative = strings.TrimPrefix(sectionRelative, "./")
	sectionRelative = strings.Trim(sectionRelative, "/")
	if sectionRelative == "" {
		return sectionName
	}
	return sectionName + "/" + sectionRelative
}

// hashFile: 파일의 SHA256 해시값 계산 (파일 변경 감지용)
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
