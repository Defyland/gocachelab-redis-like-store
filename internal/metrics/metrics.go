package metrics

import (
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

type StoreStats struct {
	LiveKeys     int
	PhysicalKeys int
	ExpiredTotal uint64
}

type Snapshot struct {
	StartTime           time.Time
	UptimeSeconds       int64
	ConnectionsAccepted uint64
	ConnectionsActive   int64
	ConnectionsClosed   uint64
	CommandsTotal       uint64
	CommandErrors       uint64
	AOFAppendErrors     uint64
	AOFReplayApplied    uint64
	AOFReplayCorrupted  uint64
	AOFReplayPartial    uint64
}

type Metrics struct {
	startTime           time.Time
	connectionsAccepted atomic.Uint64
	connectionsActive   atomic.Int64
	connectionsClosed   atomic.Uint64
	commandsTotal       atomic.Uint64
	commandErrors       atomic.Uint64
	aofAppendErrors     atomic.Uint64
	aofReplayApplied    atomic.Uint64
	aofReplayCorrupted  atomic.Uint64
	aofReplayPartial    atomic.Uint64
}

func New() *Metrics {
	return &Metrics{startTime: time.Now().UTC()}
}

func (m *Metrics) ConnectionAccepted() {
	m.connectionsAccepted.Add(1)
	m.connectionsActive.Add(1)
}

func (m *Metrics) ConnectionClosed() {
	m.connectionsClosed.Add(1)
	m.connectionsActive.Add(-1)
}

func (m *Metrics) CommandObserved() {
	m.commandsTotal.Add(1)
}

func (m *Metrics) CommandFailed() {
	m.commandErrors.Add(1)
}

func (m *Metrics) AOFAppendFailed() {
	m.aofAppendErrors.Add(1)
}

func (m *Metrics) AOFReplay(applied uint64, corrupted uint64, partial uint64) {
	m.aofReplayApplied.Add(applied)
	m.aofReplayCorrupted.Add(corrupted)
	m.aofReplayPartial.Add(partial)
}

func (m *Metrics) Snapshot() Snapshot {
	now := time.Now().UTC()
	return Snapshot{
		StartTime:           m.startTime,
		UptimeSeconds:       int64(now.Sub(m.startTime).Seconds()),
		ConnectionsAccepted: m.connectionsAccepted.Load(),
		ConnectionsActive:   m.connectionsActive.Load(),
		ConnectionsClosed:   m.connectionsClosed.Load(),
		CommandsTotal:       m.commandsTotal.Load(),
		CommandErrors:       m.commandErrors.Load(),
		AOFAppendErrors:     m.aofAppendErrors.Load(),
		AOFReplayApplied:    m.aofReplayApplied.Load(),
		AOFReplayCorrupted:  m.aofReplayCorrupted.Load(),
		AOFReplayPartial:    m.aofReplayPartial.Load(),
	}
}

func Info(snapshot Snapshot, store StoreStats) string {
	var mem runtime.MemStats
	runtimepkgReadMemStats(&mem)

	var b strings.Builder
	fmt.Fprintf(&b, "# Server\n")
	fmt.Fprintf(&b, "gocachelab_version:0.1.0\n")
	fmt.Fprintf(&b, "uptime_in_seconds:%d\n", snapshot.UptimeSeconds)
	fmt.Fprintf(&b, "go_goroutines:%d\n", runtimepkgNumGoroutine())
	fmt.Fprintf(&b, "\n# Clients\n")
	fmt.Fprintf(&b, "connected_clients:%d\n", snapshot.ConnectionsActive)
	fmt.Fprintf(&b, "connections_accepted_total:%d\n", snapshot.ConnectionsAccepted)
	fmt.Fprintf(&b, "connections_closed_total:%d\n", snapshot.ConnectionsClosed)
	fmt.Fprintf(&b, "\n# Keyspace\n")
	fmt.Fprintf(&b, "keys:%d\n", store.LiveKeys)
	fmt.Fprintf(&b, "physical_keys:%d\n", store.PhysicalKeys)
	fmt.Fprintf(&b, "expired_keys_total:%d\n", store.ExpiredTotal)
	fmt.Fprintf(&b, "\n# Commands\n")
	fmt.Fprintf(&b, "commands_total:%d\n", snapshot.CommandsTotal)
	fmt.Fprintf(&b, "command_errors_total:%d\n", snapshot.CommandErrors)
	fmt.Fprintf(&b, "\n# Persistence\n")
	fmt.Fprintf(&b, "aof_append_errors_total:%d\n", snapshot.AOFAppendErrors)
	fmt.Fprintf(&b, "aof_replay_applied_total:%d\n", snapshot.AOFReplayApplied)
	fmt.Fprintf(&b, "aof_replay_corrupted_total:%d\n", snapshot.AOFReplayCorrupted)
	fmt.Fprintf(&b, "aof_replay_partial_total:%d\n", snapshot.AOFReplayPartial)
	fmt.Fprintf(&b, "\n# Memory\n")
	fmt.Fprintf(&b, "alloc_bytes:%d\n", mem.Alloc)
	fmt.Fprintf(&b, "sys_bytes:%d\n", mem.Sys)
	return b.String()
}

func Prometheus(snapshot Snapshot, store StoreStats) string {
	var mem runtime.MemStats
	runtimepkgReadMemStats(&mem)

	var b strings.Builder
	writeGauge(&b, "gocachelab_uptime_seconds", float64(snapshot.UptimeSeconds))
	writeGauge(&b, "gocachelab_connected_clients", float64(snapshot.ConnectionsActive))
	writeCounter(&b, "gocachelab_connections_accepted_total", float64(snapshot.ConnectionsAccepted))
	writeCounter(&b, "gocachelab_connections_closed_total", float64(snapshot.ConnectionsClosed))
	writeGauge(&b, "gocachelab_keys", float64(store.LiveKeys))
	writeGauge(&b, "gocachelab_physical_keys", float64(store.PhysicalKeys))
	writeCounter(&b, "gocachelab_expired_keys_total", float64(store.ExpiredTotal))
	writeCounter(&b, "gocachelab_commands_total", float64(snapshot.CommandsTotal))
	writeCounter(&b, "gocachelab_command_errors_total", float64(snapshot.CommandErrors))
	writeCounter(&b, "gocachelab_aof_append_errors_total", float64(snapshot.AOFAppendErrors))
	writeCounter(&b, "gocachelab_aof_replay_applied_total", float64(snapshot.AOFReplayApplied))
	writeCounter(&b, "gocachelab_aof_replay_corrupted_total", float64(snapshot.AOFReplayCorrupted))
	writeCounter(&b, "gocachelab_aof_replay_partial_total", float64(snapshot.AOFReplayPartial))
	writeGauge(&b, "gocachelab_go_goroutines", float64(runtimepkgNumGoroutine()))
	writeGauge(&b, "gocachelab_alloc_bytes", float64(mem.Alloc))
	writeGauge(&b, "gocachelab_sys_bytes", float64(mem.Sys))
	return b.String()
}

func writeGauge(b *strings.Builder, name string, value float64) {
	fmt.Fprintf(b, "# TYPE %s gauge\n%s %.0f\n", name, name, value)
}

func writeCounter(b *strings.Builder, name string, value float64) {
	fmt.Fprintf(b, "# TYPE %s counter\n%s %.0f\n", name, name, value)
}

var runtimepkgReadMemStats = runtime.ReadMemStats
var runtimepkgNumGoroutine = runtime.NumGoroutine
