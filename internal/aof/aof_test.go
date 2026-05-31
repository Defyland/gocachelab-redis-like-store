package aof

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defyland/gocachelab-redis-like-store/internal/protocol"
)

func TestReplayReaderAppliesValidRecords(t *testing.T) {
	var input bytes.Buffer
	input.Write(EncodeRecord([]byte(protocol.EncodeCommand("SET", "k", "v"))))
	input.Write(EncodeRecord([]byte(protocol.EncodeCommand("DEL", "k"))))

	var commands []protocol.Command
	report := ReplayReader(&input, func(command protocol.Command) error {
		commands = append(commands, command)
		return nil
	})

	if report.AppliedRecords != 2 || report.CorruptedRecords != 0 || report.PartialRecords != 0 {
		t.Fatalf("report = %#v", report)
	}
	if len(commands) != 2 || commands[0].Name != "SET" || commands[1].Name != "DEL" {
		t.Fatalf("commands = %#v", commands)
	}
}

func TestReplayReaderReportsCorruptedRecordAndContinues(t *testing.T) {
	var input bytes.Buffer
	input.WriteString("GCL-AOF-1 3 00000000\nSET\n")
	input.Write(EncodeRecord([]byte(protocol.EncodeCommand("SET", "after", "ok"))))

	report := ReplayReader(&input, func(command protocol.Command) error {
		return nil
	})

	if report.CorruptedRecords != 1 || report.AppliedRecords != 1 {
		t.Fatalf("report = %#v", report)
	}
}

func TestReplayReaderReportsPartialTrailingRecord(t *testing.T) {
	record := EncodeRecord([]byte(protocol.EncodeCommand("SET", "k", "v")))
	truncated := record[:len(record)-3]

	report := ReplayReader(bytes.NewReader(truncated), func(command protocol.Command) error {
		t.Fatalf("partial record should not be applied")
		return nil
	})

	if report.PartialRecords != 1 || report.AppliedRecords != 0 {
		t.Fatalf("report = %#v", report)
	}
}

func TestReplayReaderCountsApplyErrorsAsCorruption(t *testing.T) {
	input := bytes.NewReader(EncodeRecord([]byte(protocol.EncodeCommand("UNKNOWN"))))
	report := ReplayReader(input, func(command protocol.Command) error {
		return os.ErrInvalid
	})
	if report.CorruptedRecords != 1 || report.AppliedRecords != 0 {
		t.Fatalf("report = %#v", report)
	}
}

func TestAppenderWritesReplayableRecords(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "appendonly.aof")
	appender, err := OpenAppender(filePath, SyncAlways)
	if err != nil {
		t.Fatalf("OpenAppender returned error: %v", err)
	}
	if err := appender.AppendRecords(
		CommandRecord{Name: "SET", Args: []string{"k", "v"}},
		CommandRecord{Name: "PERSIST", Args: []string{"k"}},
	); err != nil {
		t.Fatalf("AppendRecords returned error: %v", err)
	}
	if err := appender.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	report, err := Replay(filePath, func(command protocol.Command) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Replay returned error: %v", err)
	}
	if report.AppliedRecords != 2 {
		t.Fatalf("AppliedRecords = %d, want 2", report.AppliedRecords)
	}
}
