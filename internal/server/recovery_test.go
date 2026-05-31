package server

import (
	"strconv"
	"testing"
	"time"

	"github.com/Defyland/gocachelab-redis-like-store/internal/protocol"
	"github.com/Defyland/gocachelab-redis-like-store/internal/store"
)

func TestApplyRecoveredCommand(t *testing.T) {
	cache := store.New(nil)

	if err := ApplyRecoveredCommand(cache, protocol.Command{Name: "SET", Args: []string{"k", "v"}}); err != nil {
		t.Fatalf("SET replay returned error: %v", err)
	}
	expiresAt := time.Now().Add(time.Hour).UnixNano()
	if err := ApplyRecoveredCommand(cache, protocol.Command{Name: "EXPIREAT", Args: []string{"k", formatInt(expiresAt)}}); err != nil {
		t.Fatalf("EXPIREAT replay returned error: %v", err)
	}
	if value, ok := cache.Get("k"); !ok || value != "v" {
		t.Fatalf("recovered value = %q/%v", value, ok)
	}
	if err := ApplyRecoveredCommand(cache, protocol.Command{Name: "PERSIST", Args: []string{"k"}}); err != nil {
		t.Fatalf("PERSIST replay returned error: %v", err)
	}
	if err := ApplyRecoveredCommand(cache, protocol.Command{Name: "DEL", Args: []string{"k"}}); err != nil {
		t.Fatalf("DEL replay returned error: %v", err)
	}
	if _, ok := cache.Get("k"); ok {
		t.Fatalf("key still exists after DEL replay")
	}
}

func TestApplyRecoveredCommandRejectsInvalidCommand(t *testing.T) {
	cache := store.New(nil)
	if err := ApplyRecoveredCommand(cache, protocol.Command{Name: "SET", Args: []string{"missing-value"}}); err == nil {
		t.Fatalf("expected SET arity error")
	}
	if err := ApplyRecoveredCommand(cache, protocol.Command{Name: "EXPIREAT", Args: []string{"k", "bad"}}); err == nil {
		t.Fatalf("expected EXPIREAT parse error")
	}
	if err := ApplyRecoveredCommand(cache, protocol.Command{Name: "UNKNOWN"}); err == nil {
		t.Fatalf("expected unknown command error")
	}
}

func formatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}
