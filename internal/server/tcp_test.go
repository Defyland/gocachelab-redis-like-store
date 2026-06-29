package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Defyland/gocachelab-redis-like-store/internal/metrics"
	"github.com/Defyland/gocachelab-redis-like-store/internal/store"
)

func TestExecuteCommandLifecycle(t *testing.T) {
	cache := store.New(nil)
	snapshotPath := filepath.Join(t.TempDir(), "snapshot.json")
	srv := NewTCPServer(cache, TCPOptions{SnapshotPath: snapshotPath})

	assertExecute(t, srv, "PING\r\n", "+PONG\r\n")
	assertExecute(t, srv, "SET user:1 Ada\r\n", "+OK\r\n")
	assertExecute(t, srv, "SET session:1 token EX 60\r\n", "+OK\r\n")
	assertExecute(t, srv, "GET user:1\r\n", "$3\r\nAda\r\n")
	assertExecute(t, srv, "EXISTS user:1 missing\r\n", ":1\r\n")
	assertExecute(t, srv, "TTL user:1\r\n", ":-1\r\n")
	assertExecute(t, srv, "EXPIRE user:1 10\r\n", ":1\r\n")
	assertExecute(t, srv, "PERSIST user:1\r\n", ":1\r\n")
	assertExecute(t, srv, "KEYS user:* LIMIT 10\r\n", "*1\r\n$6\r\nuser:1\r\n")
	response, _ := srv.Execute("INFO\r\n")
	if !strings.Contains(string(response), "commands_total:") {
		t.Fatalf("INFO response missing commands_total: %q", response)
	}
	assertExecute(t, srv, "SAVE\r\n", "+OK\r\n")
	assertExecute(t, srv, "DEL user:1\r\n", ":1\r\n")
	assertExecute(t, srv, "GET user:1\r\n", "$-1\r\n")
}

func TestExecuteInvalidCommand(t *testing.T) {
	srv := NewTCPServer(store.New(nil), TCPOptions{})
	response, _ := srv.Execute("NOPE\r\n")
	if !strings.HasPrefix(string(response), "-ERR unknown command NOPE") {
		t.Fatalf("response = %q", response)
	}
}

func TestAdminHandlerExposesHealthAndMetrics(t *testing.T) {
	cache := store.New(nil)
	cache.Set("k", "v", time.Time{})
	registry := metrics.New()
	handler := NewAdminHandler(cache, registry)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("health status = %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	handler.ServeHTTP(recorder, request)
	if !strings.Contains(recorder.Body.String(), "gocachelab_keys 1") {
		t.Fatalf("metrics response = %s", recorder.Body.String())
	}
}

func TestTCPServerConcurrentClients1000(t *testing.T) {
	cache := store.New(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	addr := startTestServer(t, ctx, cache)

	const clients = 1000
	const clientDeadline = 10 * time.Second
	var wg sync.WaitGroup
	errs := make(chan error, clients)
	for i := 0; i < clients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", addr, clientDeadline)
			if err != nil {
				errs <- err
				return
			}
			defer conn.Close()
			_ = conn.SetDeadline(time.Now().Add(clientDeadline))

			key := fmt.Sprintf("client:%d", id)
			value := fmt.Sprintf("value:%d", id)
			if _, err := fmt.Fprintf(conn, "SET %s %s\r\nGET %s\r\nQUIT\r\n", key, value, key); err != nil {
				errs <- err
				return
			}

			reader := bufio.NewReader(conn)
			if line, err := reader.ReadString('\n'); err != nil || line != "+OK\r\n" {
				errs <- fmt.Errorf("SET response line=%q err=%v", line, err)
				return
			}
			got, err := readBulk(reader)
			if err != nil {
				errs <- err
				return
			}
			if got != value {
				errs <- fmt.Errorf("GET value=%q want=%q", got, value)
				return
			}
			if line, err := reader.ReadString('\n'); err != nil || line != "+OK\r\n" {
				errs <- fmt.Errorf("QUIT response line=%q err=%v", line, err)
				return
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if stats := cache.Stats(); stats.LiveKeys != clients {
		t.Fatalf("LiveKeys = %d, want %d", stats.LiveKeys, clients)
	}
}

func TestClientDisconnectMidCommandDoesNotBlockServer(t *testing.T) {
	cache := store.New(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	addr := startTestServer(t, ctx, cache)

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("DialTimeout returned error: %v", err)
	}
	if _, err := io.WriteString(conn, "SET unfinished"); err != nil {
		t.Fatalf("WriteString returned error: %v", err)
	}
	_ = conn.Close()

	healthy, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("server did not accept another connection: %v", err)
	}
	defer healthy.Close()
	if _, err := io.WriteString(healthy, "PING\r\n"); err != nil {
		t.Fatalf("PING write failed: %v", err)
	}
	line, err := bufio.NewReader(healthy).ReadString('\n')
	if err != nil || line != "+PONG\r\n" {
		t.Fatalf("PING response line=%q err=%v", line, err)
	}
}

func assertExecute(t *testing.T, srv *TCPServer, line string, want string) {
	t.Helper()
	response, closeConn := srv.Execute(line)
	if closeConn {
		t.Fatalf("closeConn = true")
	}
	if string(response) != want {
		t.Fatalf("response = %q, want %q", response, want)
	}
}

func startTestServer(t *testing.T, ctx context.Context, cache *store.Store) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen returned error: %v", err)
	}
	srv := NewTCPServer(cache, TCPOptions{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ctx, listener)
	}()
	t.Cleanup(func() {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		_ = cancelCtx
		_ = listener.Close()
		if err := <-errCh; err != nil {
			t.Fatalf("Serve returned error: %v", err)
		}
	})
	return listener.Addr().String()
}

func readBulk(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(line, "$") {
		return "", fmt.Errorf("bulk header = %q", line)
	}
	var size int
	if _, err := fmt.Sscanf(line, "$%d\r\n", &size); err != nil {
		return "", err
	}
	payload := make([]byte, size+2)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return "", err
	}
	return string(payload[:size]), nil
}
