package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	Bucket     string    `json:"bucket"`
	Key        string    `json:"key"`
	LocalPath  string    `json:"local_path"`
	URL        string    `json:"url"`
	ExpiresAt  time.Time `json:"expires_at"`
	UploadedAt time.Time `json:"uploaded_at"`
	Size       int64     `json:"size"`
}

type Store struct {
	Entries []Entry `json:"entries"`
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	dir := filepath.Join(home, ".local", "share", "presign")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating state directory %s: %w", dir, err)
	}
	return filepath.Join(dir, "state.json"), nil
}

func Load(path string) (*Store, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{}, nil
		}
		return nil, fmt.Errorf("reading state %s: %w", path, err)
	}

	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("parsing state %s: %w", path, err)
	}
	return &store, nil
}

func Save(path string, store *Store) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("writing state %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("renaming state %s: %w", path, err)
	}
	return nil
}

func (s *Store) Add(entry Entry) {
	for i, e := range s.Entries {
		if e.Bucket == entry.Bucket && e.Key == entry.Key {
			s.Entries[i] = entry
			return
		}
	}
	s.Entries = append(s.Entries, entry)
}

func (s *Store) Remove(bucket, key string) bool {
	for i, e := range s.Entries {
		if e.Bucket == bucket && e.Key == key {
			s.Entries = append(s.Entries[:i], s.Entries[i+1:]...)
			return true
		}
	}
	return false
}
