package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadNonexistent(t *testing.T) {
	store, err := Load("/nonexistent/state.json")
	if err != nil {
		t.Fatalf("Load() with missing file should not error, got: %v", err)
	}
	if len(store.Entries) != 0 {
		t.Errorf("Entries = %d, want 0", len(store.Entries))
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	now := time.Now().Truncate(time.Second)
	store := &Store{
		Entries: []Entry{
			{
				Bucket:     "test-bucket",
				Key:        "test-key.txt",
				LocalPath:  "/tmp/test.txt",
				URL:        "https://example.com/presigned",
				ExpiresAt:  now.Add(24 * time.Hour),
				UploadedAt: now,
				Size:       1234,
			},
		},
	}

	if err := Save(path, store); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(loaded.Entries) != 1 {
		t.Fatalf("Entries = %d, want 1", len(loaded.Entries))
	}

	e := loaded.Entries[0]
	if e.Bucket != "test-bucket" {
		t.Errorf("Bucket = %q, want %q", e.Bucket, "test-bucket")
	}
	if e.Key != "test-key.txt" {
		t.Errorf("Key = %q, want %q", e.Key, "test-key.txt")
	}
	if e.Size != 1234 {
		t.Errorf("Size = %d, want %d", e.Size, 1234)
	}
}

func TestSaveAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	store := &Store{Entries: []Entry{{Bucket: "b", Key: "k"}}}
	if err := Save(path, store); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify no .tmp file is left behind
	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temporary file should not remain after save")
	}

	// Verify the file is valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var check Store
	if err := json.Unmarshal(data, &check); err != nil {
		t.Errorf("saved file is not valid JSON: %v", err)
	}
}

func TestStoreAdd(t *testing.T) {
	store := &Store{}

	store.Add(Entry{Bucket: "b1", Key: "k1", Size: 100})
	if len(store.Entries) != 1 {
		t.Fatalf("Entries = %d, want 1", len(store.Entries))
	}

	// Add a different entry
	store.Add(Entry{Bucket: "b1", Key: "k2", Size: 200})
	if len(store.Entries) != 2 {
		t.Fatalf("Entries = %d, want 2", len(store.Entries))
	}

	// Update existing entry (same bucket+key)
	store.Add(Entry{Bucket: "b1", Key: "k1", Size: 999})
	if len(store.Entries) != 2 {
		t.Fatalf("Entries = %d after update, want 2", len(store.Entries))
	}
	if store.Entries[0].Size != 999 {
		t.Errorf("Size = %d, want 999 after update", store.Entries[0].Size)
	}
}

func TestStoreRemove(t *testing.T) {
	store := &Store{
		Entries: []Entry{
			{Bucket: "b1", Key: "k1"},
			{Bucket: "b1", Key: "k2"},
			{Bucket: "b2", Key: "k1"},
		},
	}

	removed := store.Remove("b1", "k2")
	if !removed {
		t.Error("Remove() returned false for existing entry")
	}
	if len(store.Entries) != 2 {
		t.Fatalf("Entries = %d, want 2", len(store.Entries))
	}

	// Verify remaining entries
	for _, e := range store.Entries {
		if e.Bucket == "b1" && e.Key == "k2" {
			t.Error("removed entry still present")
		}
	}
}

func TestStoreRemoveNonexistent(t *testing.T) {
	store := &Store{
		Entries: []Entry{{Bucket: "b1", Key: "k1"}},
	}

	removed := store.Remove("b1", "missing")
	if removed {
		t.Error("Remove() returned true for nonexistent entry")
	}
	if len(store.Entries) != 1 {
		t.Errorf("Entries = %d, want 1", len(store.Entries))
	}
}

func TestStoreRemoveFromEmpty(t *testing.T) {
	store := &Store{}
	removed := store.Remove("b", "k")
	if removed {
		t.Error("Remove() returned true on empty store")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("{invalid}"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("Load() with invalid JSON should return error")
	}
}

func TestDefaultPath(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error: %v", err)
	}
	if filepath.Base(path) != "state.json" {
		t.Errorf("DefaultPath() = %q, want state.json basename", path)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("DefaultPath() = %q, want absolute path", path)
	}
}
