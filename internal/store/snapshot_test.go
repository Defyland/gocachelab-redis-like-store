package store

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadSnapshotFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "snapshot.json")
	snapshot := Snapshot{
		Version:   SnapshotVersion,
		CreatedAt: time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC),
		Entries: map[string]SnapshotEntry{
			"k": {Value: "v"},
		},
	}

	if err := SaveSnapshotFile(filePath, snapshot); err != nil {
		t.Fatalf("SaveSnapshotFile returned error: %v", err)
	}
	loaded, err := LoadSnapshotFile(filePath)
	if err != nil {
		t.Fatalf("LoadSnapshotFile returned error: %v", err)
	}
	if loaded.Entries["k"].Value != "v" {
		t.Fatalf("loaded entry = %#v", loaded.Entries["k"])
	}
}
