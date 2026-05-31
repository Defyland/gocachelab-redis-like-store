package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Defyland/gocachelab-redis-like-store/internal/aof"
	"github.com/Defyland/gocachelab-redis-like-store/internal/metrics"
	"github.com/Defyland/gocachelab-redis-like-store/internal/protocol"
	"github.com/Defyland/gocachelab-redis-like-store/internal/store"
)

const (
	DefaultMaxLineBytes = 1024 * 1024
	DefaultKeyLimit     = 1000
)

type TCPServer struct {
	store        *store.Store
	appender     *aof.Appender
	metrics      *metrics.Metrics
	logger       *slog.Logger
	maxLineBytes int
	keyLimit     int
	snapshotPath string
}

type TCPOptions struct {
	Appender     *aof.Appender
	Metrics      *metrics.Metrics
	Logger       *slog.Logger
	MaxLineBytes int
	KeyLimit     int
	SnapshotPath string
}

func NewTCPServer(store *store.Store, opts TCPOptions) *TCPServer {
	if opts.Metrics == nil {
		opts.Metrics = metrics.New()
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.MaxLineBytes <= 0 {
		opts.MaxLineBytes = DefaultMaxLineBytes
	}
	if opts.KeyLimit <= 0 {
		opts.KeyLimit = DefaultKeyLimit
	}
	return &TCPServer{
		store:        store,
		appender:     opts.Appender,
		metrics:      opts.Metrics,
		logger:       opts.Logger,
		maxLineBytes: opts.MaxLineBytes,
		keyLimit:     opts.KeyLimit,
		snapshotPath: opts.SnapshotPath,
	}
}

func (s *TCPServer) Serve(ctx context.Context, listener net.Listener) error {
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			s.logger.Error("tcp accept failed", "error", err)
			continue
		}
		s.metrics.ConnectionAccepted()
		go s.handleConn(ctx, conn)
	}
}

func (s *TCPServer) handleConn(ctx context.Context, conn net.Conn) {
	defer func() {
		s.metrics.ConnectionClosed()
		_ = conn.Close()
	}()

	reader := bufio.NewReaderSize(conn, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line, err := reader.ReadString('\n')
		if len(line) > s.maxLineBytes {
			s.metrics.CommandFailed()
			if !writeAll(conn, protocol.Error("command line exceeds max size")) {
				return
			}
			return
		}
		if err != nil {
			if err == io.EOF {
				return
			}
			s.metrics.CommandFailed()
			return
		}

		response, closeConn := s.Execute(line)
		if !writeAll(conn, response) || closeConn {
			return
		}
	}
}

func (s *TCPServer) Execute(line string) ([]byte, bool) {
	command, err := protocol.ParseLine(line)
	if err != nil {
		s.metrics.CommandFailed()
		return protocol.Error(err.Error()), false
	}
	s.metrics.CommandObserved()

	response, closeConn := s.dispatch(command)
	if len(response) > 0 && response[0] == '-' {
		s.metrics.CommandFailed()
	}
	return response, closeConn
}

func (s *TCPServer) dispatch(command protocol.Command) ([]byte, bool) {
	switch command.Name {
	case "PING":
		if len(command.Args) > 1 {
			return protocol.Error("PING expects zero or one argument"), false
		}
		if len(command.Args) == 1 {
			return protocol.BulkString(command.Args[0]), false
		}
		return protocol.SimpleString("PONG"), false
	case "SET":
		return s.handleSet(command.Args), false
	case "GET":
		if len(command.Args) != 1 {
			return protocol.Error("GET expects 1 argument"), false
		}
		value, ok := s.store.Get(command.Args[0])
		if !ok {
			return protocol.NullBulkString(), false
		}
		return protocol.BulkString(value), false
	case "DEL":
		if len(command.Args) < 1 {
			return protocol.Error("DEL expects at least 1 argument"), false
		}
		if err := s.appendAOF(aof.CommandRecord{Name: "DEL", Args: command.Args}); err != nil {
			return protocol.Error(err.Error()), false
		}
		return protocol.Integer(int64(s.store.Del(command.Args...))), false
	case "EXISTS":
		if len(command.Args) < 1 {
			return protocol.Error("EXISTS expects at least 1 argument"), false
		}
		return protocol.Integer(int64(s.store.Exists(command.Args...))), false
	case "EXPIRE":
		return s.handleExpire(command.Args), false
	case "TTL":
		if len(command.Args) != 1 {
			return protocol.Error("TTL expects 1 argument"), false
		}
		return protocol.Integer(s.ttlSeconds(command.Args[0])), false
	case "PERSIST":
		if len(command.Args) != 1 {
			return protocol.Error("PERSIST expects 1 argument"), false
		}
		if s.store.Exists(command.Args[0]) == 0 {
			return protocol.Integer(0), false
		}
		if err := s.appendAOF(aof.CommandRecord{Name: "PERSIST", Args: command.Args}); err != nil {
			return protocol.Error(err.Error()), false
		}
		if s.store.Persist(command.Args[0]) {
			return protocol.Integer(1), false
		}
		return protocol.Integer(0), false
	case "KEYS":
		return s.handleKeys(command.Args), false
	case "INFO":
		storeStats := s.store.Stats()
		info := metrics.Info(s.metrics.Snapshot(), metrics.StoreStats{
			LiveKeys:     storeStats.LiveKeys,
			PhysicalKeys: storeStats.PhysicalKeys,
			ExpiredTotal: storeStats.ExpiredTotal,
		})
		return protocol.BulkString(info), false
	case "SAVE":
		if len(command.Args) != 0 {
			return protocol.Error("SAVE expects no arguments"), false
		}
		if s.snapshotPath == "" {
			return protocol.Error("snapshot path is not configured"), false
		}
		if err := store.SaveSnapshotFile(s.snapshotPath, s.store.Snapshot()); err != nil {
			return protocol.Error(err.Error()), false
		}
		return protocol.SimpleString("OK"), false
	case "QUIT":
		if len(command.Args) != 0 {
			return protocol.Error("QUIT expects no arguments"), false
		}
		return protocol.SimpleString("OK"), true
	default:
		return protocol.Error("unknown command " + command.Name), false
	}
}

func (s *TCPServer) handleSet(args []string) []byte {
	if len(args) != 2 && len(args) != 4 {
		return protocol.Error("SET expects key value [EX seconds|PX milliseconds]")
	}
	key := args[0]
	value := args[1]
	var expiresAt time.Time
	records := []aof.CommandRecord{{Name: "SET", Args: []string{key, value}}}

	if len(args) == 4 {
		ttl, err := parseTTL(args[2], args[3])
		if err != nil {
			return protocol.Error(err.Error())
		}
		expiresAt = time.Now().UTC().Add(ttl)
		records = append(records, aof.CommandRecord{
			Name: "EXPIREAT",
			Args: []string{key, strconv.FormatInt(expiresAt.UnixNano(), 10)},
		})
	}

	if err := s.appendAOF(records...); err != nil {
		return protocol.Error(err.Error())
	}
	s.store.Set(key, value, expiresAt)
	return protocol.SimpleString("OK")
}

func (s *TCPServer) handleExpire(args []string) []byte {
	if len(args) != 2 {
		return protocol.Error("EXPIRE expects key seconds")
	}
	seconds, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return protocol.Error("EXPIRE seconds must be an integer")
	}
	if s.store.Exists(args[0]) == 0 {
		return protocol.Integer(0)
	}
	if seconds <= 0 {
		if err := s.appendAOF(aof.CommandRecord{Name: "DEL", Args: []string{args[0]}}); err != nil {
			return protocol.Error(err.Error())
		}
		return protocol.Integer(int64(s.store.Del(args[0])))
	}
	expiresAt := time.Now().UTC().Add(time.Duration(seconds) * time.Second)
	if err := s.appendAOF(aof.CommandRecord{
		Name: "EXPIREAT",
		Args: []string{args[0], strconv.FormatInt(expiresAt.UnixNano(), 10)},
	}); err != nil {
		return protocol.Error(err.Error())
	}
	if s.store.ExpireAt(args[0], expiresAt) {
		return protocol.Integer(1)
	}
	return protocol.Integer(0)
}

func (s *TCPServer) handleKeys(args []string) []byte {
	if len(args) != 1 && len(args) != 3 {
		return protocol.Error("KEYS expects pattern [LIMIT n]")
	}
	limit := s.keyLimit
	if len(args) == 3 {
		if strings.ToUpper(args[1]) != "LIMIT" {
			return protocol.Error("KEYS optional argument must be LIMIT")
		}
		parsed, err := strconv.Atoi(args[2])
		if err != nil || parsed <= 0 {
			return protocol.Error("KEYS LIMIT must be a positive integer")
		}
		if parsed < limit {
			limit = parsed
		}
	}
	return protocol.Array(s.store.Keys(args[0], limit))
}

func (s *TCPServer) ttlSeconds(key string) int64 {
	ttl, status := s.store.TTL(key)
	switch status {
	case store.TTLMissing:
		return -2
	case store.TTLNoExpiry:
		return -1
	default:
		if ttl <= 0 {
			return -2
		}
		return int64(math.Ceil(ttl.Seconds()))
	}
}

func (s *TCPServer) appendAOF(records ...aof.CommandRecord) error {
	if s.appender == nil {
		return nil
	}
	if err := s.appender.AppendRecords(records...); err != nil {
		s.metrics.AOFAppendFailed()
		return fmt.Errorf("aof append failed: %w", err)
	}
	return nil
}

func parseTTL(unit string, value string) (time.Duration, error) {
	amount, err := strconv.ParseInt(value, 10, 64)
	if err != nil || amount <= 0 {
		return 0, fmt.Errorf("TTL value must be a positive integer")
	}
	switch strings.ToUpper(unit) {
	case "EX":
		return time.Duration(amount) * time.Second, nil
	case "PX":
		return time.Duration(amount) * time.Millisecond, nil
	default:
		return 0, fmt.Errorf("SET optional TTL must use EX or PX")
	}
}

func writeAll(writer io.Writer, payload []byte) bool {
	for len(payload) > 0 {
		n, err := writer.Write(payload)
		if err != nil {
			return false
		}
		payload = payload[n:]
	}
	return true
}
