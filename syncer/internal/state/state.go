package state

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type FileRecord struct {
	Hash    string    `json:"hash"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

type Snapshot struct {
	GeneratedAt time.Time             `json:"generated_at"`
	Files       map[string]FileRecord `json:"files"`
}

type Store interface {
	Load(ctx context.Context) (*Snapshot, error)
	Save(ctx context.Context, snapshot *Snapshot) error
}

type FileStore struct {
	path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func NewSnapshot() *Snapshot {
	return &Snapshot{
		GeneratedAt: time.Time{},
		Files:       make(map[string]FileRecord),
	}
}

func (s *FileStore) Load(ctx context.Context) (*Snapshot, error) {
	_ = ctx
	if s == nil || s.path == "" {
		return NewSnapshot(), nil
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewSnapshot(), nil
		}
		return nil, err
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}

	if snapshot.Files == nil {
		snapshot.Files = make(map[string]FileRecord)
	}

	return &snapshot, nil
}

func (s *FileStore) Save(ctx context.Context, snapshot *Snapshot) error {
	_ = ctx
	if s == nil || s.path == "" {
		return errors.New("no snapshot path configured")
	}

	if snapshot == nil {
		snapshot = NewSnapshot()
	}

	snapshot.GeneratedAt = time.Now().UTC()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}

	if err := os.Rename(tmp, s.path); err != nil {
		os.Remove(tmp)
		return err
	}

	return nil
}
