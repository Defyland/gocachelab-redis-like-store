package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const SnapshotVersion = 1

type Snapshot struct {
	Version   int                      `json:"version"`
	CreatedAt time.Time                `json:"created_at"`
	Entries   map[string]SnapshotEntry `json:"entries"`
}

type SnapshotEntry struct {
	Value             string `json:"value"`
	ExpiresAtUnixNano int64  `json:"expires_at_unix_nano,omitempty"`
}

func SaveSnapshotFile(filePath string, snapshot Snapshot) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}

	temp, err := os.CreateTemp(filepath.Dir(filePath), ".snapshot-*.json")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempName)
		}
	}()

	encoder := json.NewEncoder(temp)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(snapshot); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempName, filePath); err != nil {
		return err
	}
	removeTemp = false
	return nil
}

func LoadSnapshotFile(filePath string) (Snapshot, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Snapshot{Version: SnapshotVersion, Entries: map[string]SnapshotEntry{}}, nil
		}
		return Snapshot{}, err
	}
	defer file.Close()

	var snapshot Snapshot
	if err := json.NewDecoder(file).Decode(&snapshot); err != nil {
		return Snapshot{}, err
	}
	if snapshot.Version != SnapshotVersion {
		return Snapshot{}, errors.New("unsupported snapshot version")
	}
	if snapshot.Entries == nil {
		snapshot.Entries = map[string]SnapshotEntry{}
	}
	return snapshot, nil
}
