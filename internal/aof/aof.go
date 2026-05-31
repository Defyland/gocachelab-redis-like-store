package aof

import (
	"bufio"
	"context"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Defyland/gocachelab-redis-like-store/internal/protocol"
)

const (
	headerPrefix     = "GCL-AOF-1"
	maxRecordPayload = 16 * 1024 * 1024
)

type SyncPolicy string

const (
	SyncNever    SyncPolicy = "never"
	SyncEverySec SyncPolicy = "everysec"
	SyncAlways   SyncPolicy = "always"
)

type CommandRecord struct {
	Name string
	Args []string
}

type ReplayReport struct {
	AppliedRecords   uint64
	CorruptedRecords uint64
	PartialRecords   uint64
}

type Appender struct {
	mu     sync.Mutex
	file   *os.File
	policy SyncPolicy
	cancel context.CancelFunc
	done   chan struct{}
}

func ParseSyncPolicy(value string) (SyncPolicy, error) {
	switch SyncPolicy(strings.ToLower(strings.TrimSpace(value))) {
	case "", SyncEverySec:
		return SyncEverySec, nil
	case SyncNever:
		return SyncNever, nil
	case SyncAlways:
		return SyncAlways, nil
	default:
		return "", fmt.Errorf("unsupported AOF fsync policy %q", value)
	}
}

func OpenAppender(filePath string, policy SyncPolicy) (*Appender, error) {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}

	appender := &Appender{file: file, policy: policy}
	if policy == SyncEverySec {
		ctx, cancel := context.WithCancel(context.Background())
		appender.cancel = cancel
		appender.done = make(chan struct{})
		go appender.syncLoop(ctx)
	}
	return appender, nil
}

func (a *Appender) Append(name string, args ...string) error {
	return a.AppendRecords(CommandRecord{Name: name, Args: args})
}

func (a *Appender) AppendRecords(records ...CommandRecord) error {
	if a == nil || len(records) == 0 {
		return nil
	}

	payload := make([]byte, 0, 256*len(records))
	for _, record := range records {
		line := protocol.EncodeCommand(record.Name, record.Args...)
		payload = append(payload, EncodeRecord([]byte(line))...)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, err := a.file.Write(payload); err != nil {
		return err
	}
	if a.policy == SyncAlways {
		return a.file.Sync()
	}
	return nil
}

func (a *Appender) Close() error {
	if a == nil {
		return nil
	}
	if a.cancel != nil {
		a.cancel()
		<-a.done
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.file.Sync(); err != nil {
		_ = a.file.Close()
		return err
	}
	return a.file.Close()
}

func (a *Appender) syncLoop(ctx context.Context) {
	defer close(a.done)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.mu.Lock()
			_ = a.file.Sync()
			a.mu.Unlock()
		}
	}
}

func EncodeRecord(payload []byte) []byte {
	checksum := crc32.ChecksumIEEE(payload)
	header := fmt.Sprintf("%s %d %08x\n", headerPrefix, len(payload), checksum)
	record := make([]byte, 0, len(header)+len(payload)+1)
	record = append(record, header...)
	record = append(record, payload...)
	record = append(record, '\n')
	return record
}

func Replay(filePath string, apply func(protocol.Command) error) (ReplayReport, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ReplayReport{}, nil
		}
		return ReplayReport{}, err
	}
	defer file.Close()
	return ReplayReader(file, apply), nil
}

func ReplayReader(reader io.Reader, apply func(protocol.Command) error) ReplayReport {
	buffered := bufio.NewReader(reader)
	var report ReplayReport

	for {
		header, err := buffered.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if strings.TrimSpace(header) != "" {
					report.PartialRecords++
				}
				break
			}
			report.CorruptedRecords++
			break
		}

		length, checksum, ok := parseHeader(header)
		if !ok {
			report.CorruptedRecords++
			continue
		}

		payloadWithNewline := make([]byte, length+1)
		if _, err := io.ReadFull(buffered, payloadWithNewline); err != nil {
			report.PartialRecords++
			break
		}
		if payloadWithNewline[length] != '\n' {
			report.CorruptedRecords++
			continue
		}

		payload := payloadWithNewline[:length]
		if crc32.ChecksumIEEE(payload) != checksum {
			report.CorruptedRecords++
			continue
		}

		command, err := protocol.ParseLine(string(payload))
		if err != nil {
			report.CorruptedRecords++
			continue
		}
		if err := apply(command); err != nil {
			report.CorruptedRecords++
			continue
		}
		report.AppliedRecords++
	}

	return report
}

func parseHeader(header string) (int, uint32, bool) {
	fields := strings.Fields(strings.TrimRight(header, "\r\n"))
	if len(fields) != 3 || fields[0] != headerPrefix {
		return 0, 0, false
	}
	length, err := strconv.Atoi(fields[1])
	if err != nil || length < 0 || length > maxRecordPayload {
		return 0, 0, false
	}
	checksum64, err := strconv.ParseUint(fields[2], 16, 32)
	if err != nil {
		return 0, 0, false
	}
	return length, uint32(checksum64), true
}
