package aof

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/Defyland/gocachelab-redis-like-store/internal/protocol"
)

func BenchmarkAOFReplay100kRecords(b *testing.B) {
	var payload bytes.Buffer
	for i := 0; i < 100_000; i++ {
		payload.Write(EncodeRecord([]byte(protocol.EncodeCommand("SET", fmt.Sprintf("key:%d", i), "value"))))
	}
	data := payload.Bytes()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		report := ReplayReader(bytes.NewReader(data), func(command protocol.Command) error {
			return nil
		})
		if report.AppliedRecords != 100_000 {
			b.Fatalf("AppliedRecords = %d", report.AppliedRecords)
		}
	}
}
