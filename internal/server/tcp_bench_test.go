package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sort"
	"testing"
	"time"

	"github.com/Defyland/gocachelab-redis-like-store/internal/store"
)

func BenchmarkTCPSetGetLatencyPercentiles(b *testing.B) {
	cache := store.New(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	addr := startBenchmarkServer(b, ctx, cache)

	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		b.Fatalf("DialTimeout returned error: %v", err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)

	samples := b.N
	if samples < 1000 {
		samples = 1000
	}
	setLatencies := make([]time.Duration, 0, samples)
	getLatencies := make([]time.Duration, 0, samples)

	b.ResetTimer()
	for i := 0; i < samples; i++ {
		key := fmt.Sprintf("latency:%d", i)
		start := time.Now()
		if _, err := fmt.Fprintf(conn, "SET %s value\r\n", key); err != nil {
			b.Fatalf("SET write failed: %v", err)
		}
		if line, err := reader.ReadString('\n'); err != nil || line != "+OK\r\n" {
			b.Fatalf("SET response line=%q err=%v", line, err)
		}
		setLatencies = append(setLatencies, time.Since(start))

		start = time.Now()
		if _, err := fmt.Fprintf(conn, "GET %s\r\n", key); err != nil {
			b.Fatalf("GET write failed: %v", err)
		}
		if value, err := readBulk(reader); err != nil || value != "value" {
			b.Fatalf("GET response value=%q err=%v", value, err)
		}
		getLatencies = append(getLatencies, time.Since(start))
	}
	b.StopTimer()

	reportLatencyPercentiles(b, "set", setLatencies)
	reportLatencyPercentiles(b, "get", getLatencies)
}

func startBenchmarkServer(b *testing.B, ctx context.Context, cache *store.Store) string {
	b.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("Listen returned error: %v", err)
	}
	srv := NewTCPServer(cache, TCPOptions{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ctx, listener)
	}()
	b.Cleanup(func() {
		_ = listener.Close()
		if err := <-errCh; err != nil {
			b.Fatalf("Serve returned error: %v", err)
		}
	})
	return listener.Addr().String()
}

func reportLatencyPercentiles(b *testing.B, name string, values []time.Duration) {
	b.Helper()
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	b.ReportMetric(float64(percentile(values, 0.50).Microseconds()), name+"_p50_us")
	b.ReportMetric(float64(percentile(values, 0.95).Microseconds()), name+"_p95_us")
	b.ReportMetric(float64(percentile(values, 0.99).Microseconds()), name+"_p99_us")
}

func percentile(values []time.Duration, p float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values)-1) * p)
	return values[index]
}
