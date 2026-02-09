// diff.go - 파일 차이점 분석 모듈
// 시스템 파일과 SyncData 파일을 비교하여 백업이 필요한 파일을 파악합니다.
package engine

import (
	"sort"

	"github.com/nir414/pc-setup/syncer/internal/state"
)

// buildDiff 함수: 시스템 파일, SyncData 파일, 이전 스냅샷을 비교하여 차이점 목록 생성
func buildDiff(systemFiles, repoFiles fileMap, snapshot *state.Snapshot) *diffResult {
	keys := make(map[string]struct{})
	for key := range systemFiles {
		keys[key] = struct{}{}
	}
	for key := range repoFiles {
		keys[key] = struct{}{}
	}
	if snapshot != nil {
		for key := range snapshot.Files {
			keys[key] = struct{}{}
		}
	}

	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)

	entries := make([]DiffEntry, 0, len(ordered))

	for _, key := range ordered {
		sys := systemFiles[key]
		repo := repoFiles[key]
		if sys == nil && repo == nil {
			continue
		}
		prev, hasPrev := snapshotLookup(snapshot, key)
		status := classifyDifference(sys, repo, prev, hasPrev)
		entries = append(entries, DiffEntry{
			Path:   key,
			Status: status,
			System: sys,
			Repo:   repo,
		})
	}

	return &diffResult{Entries: entries}
}

func snapshotLookup(snapshot *state.Snapshot, key string) (state.FileRecord, bool) {
	if snapshot == nil {
		return state.FileRecord{}, false
	}
	rec, ok := snapshot.Files[key]
	return rec, ok
}

// classifyDifference 함수: 백업 관점에서 파일 상태 분류
// - 시스템 파일과 SyncData 파일을 비교하여 백업이 필요한지 판단
func classifyDifference(sys, repo *FileInfo, prev state.FileRecord, hasPrev bool) DiffStatus {
	switch {
	case sys != nil && repo != nil:
		// 시스템과 SyncData 모두 존재
		if sys.Hash == repo.Hash {
			return DiffStatusUpToDate
		}
		// 해시가 다른 경우 - 이전 스냅샷과 비교
		if hasPrev && prev.Hash != sys.Hash {
			return DiffStatusSystemModified
		}
		return DiffStatusSystemModified

	case sys != nil:
		// 시스템에만 존재 (SyncData에 없음)
		if hasPrev {
			// 이전에는 있었는데 SyncData에서 삭제됨
			return DiffStatusRepoOnly
		}
		// 새로 추가된 파일
		return DiffStatusSystemAdded

	case repo != nil:
		// SyncData에만 존재 (시스템에 없음)
		if hasPrev && prev.Hash == repo.Hash {
			// 시스템에서 삭제됨
			return DiffStatusSystemDeleted
		}
		return DiffStatusRepoOnly

	default:
		return DiffStatusUpToDate
	}
}
