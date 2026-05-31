package metrics

import (
	"strings"
	"testing"
)

func TestMetricsInfoAndPrometheus(t *testing.T) {
	registry := New()
	registry.ConnectionAccepted()
	registry.CommandObserved()
	registry.CommandFailed()
	registry.AOFAppendFailed()
	registry.AOFReplay(2, 1, 1)
	registry.ConnectionClosed()

	snapshot := registry.Snapshot()
	store := StoreStats{LiveKeys: 3, PhysicalKeys: 4, ExpiredTotal: 5}

	info := Info(snapshot, store)
	for _, expected := range []string{
		"connected_clients:0",
		"commands_total:1",
		"command_errors_total:1",
		"aof_replay_corrupted_total:1",
		"keys:3",
	} {
		if !strings.Contains(info, expected) {
			t.Fatalf("INFO missing %q:\n%s", expected, info)
		}
	}

	prometheus := Prometheus(snapshot, store)
	for _, expected := range []string{
		"gocachelab_connected_clients 0",
		"gocachelab_commands_total 1",
		"gocachelab_expired_keys_total 5",
	} {
		if !strings.Contains(prometheus, expected) {
			t.Fatalf("Prometheus missing %q:\n%s", expected, prometheus)
		}
	}
}
