package engine

import (
	"sort"

	"github.com/nir414/pc-setup/syncer/internal/state"
)

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

func classifyDifference(sys, repo *FileInfo, prev state.FileRecord, hasPrev bool) DiffStatus {
	switch {
	case sys != nil && repo != nil:
		if sys.Hash == repo.Hash {
			return DiffStatusUpToDate
		}
		if hasPrev {
			switch {
			case prev.Hash == repo.Hash:
				return DiffStatusSystemModified
			case prev.Hash == sys.Hash:
				return DiffStatusRepoModified
			}
		}
		return DiffStatusConflict
	case sys != nil:
		if hasPrev {
			if prev.Hash == sys.Hash {
				return DiffStatusRepoDeleted
			}
			return DiffStatusConflict
		}
		return DiffStatusSystemAdded
	case repo != nil:
		if hasPrev {
			if prev.Hash == repo.Hash {
				return DiffStatusSystemDeleted
			}
			return DiffStatusRepoModified
		}
		return DiffStatusRepoAdded
	default:
		return DiffStatusUpToDate
	}
}
