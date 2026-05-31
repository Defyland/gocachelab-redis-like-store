package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/Defyland/gocachelab-redis-like-store/internal/aof"
	"github.com/Defyland/gocachelab-redis-like-store/internal/metrics"
	"github.com/Defyland/gocachelab-redis-like-store/internal/protocol"
	"github.com/Defyland/gocachelab-redis-like-store/internal/server"
	"github.com/Defyland/gocachelab-redis-like-store/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("gocachelab stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	cache := store.New(nil)
	snapshot, err := store.LoadSnapshotFile(cfg.snapshotPath)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}
	cache.Restore(snapshot)

	registry := metrics.New()
	replayReport, err := aof.Replay(cfg.aofPath, func(command protocol.Command) error {
		return server.ApplyRecoveredCommand(cache, command)
	})
	if err != nil {
		return fmt.Errorf("replay aof: %w", err)
	}
	registry.AOFReplay(replayReport.AppliedRecords, replayReport.CorruptedRecords, replayReport.PartialRecords)

	appender, err := aof.OpenAppender(cfg.aofPath, cfg.aofSyncPolicy)
	if err != nil {
		return fmt.Errorf("open aof: %w", err)
	}
	defer appender.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go runTTLCleanup(ctx, cache, cfg.cleanupInterval, cfg.cleanupBatch, logger)

	admin := &http.Server{
		Addr:              cfg.adminAddr,
		Handler:           server.NewAdminHandler(cache, registry),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		logger.Info("admin listener started", "addr", cfg.adminAddr)
		if err := admin.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("admin listener failed", "error", err)
			stop()
		}
	}()

	listener, err := net.Listen("tcp", cfg.tcpAddr)
	if err != nil {
		return fmt.Errorf("listen tcp: %w", err)
	}
	tcpServer := server.NewTCPServer(cache, server.TCPOptions{
		Appender:     appender,
		Metrics:      registry,
		Logger:       logger,
		MaxLineBytes: cfg.maxLineBytes,
		KeyLimit:     cfg.keyLimit,
		SnapshotPath: cfg.snapshotPath,
	})

	errCh := make(chan error, 1)
	go func() {
		logger.Info("tcp listener started", "addr", cfg.tcpAddr)
		errCh <- tcpServer.Serve(ctx, listener)
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = admin.Shutdown(shutdownCtx)
	return nil
}

type config struct {
	tcpAddr         string
	adminAddr       string
	dataDir         string
	aofPath         string
	snapshotPath    string
	aofSyncPolicy   aof.SyncPolicy
	cleanupInterval time.Duration
	cleanupBatch    int
	maxLineBytes    int
	keyLimit        int
}

func loadConfig() (config, error) {
	dataDir := envString("GOCACHELAB_DATA_DIR", "./data")
	policy, err := aof.ParseSyncPolicy(envString("GOCACHELAB_AOF_FSYNC", string(aof.SyncEverySec)))
	if err != nil {
		return config{}, err
	}
	cleanupInterval, err := envDuration("GOCACHELAB_TTL_CLEANUP_INTERVAL", time.Second)
	if err != nil {
		return config{}, err
	}

	return config{
		tcpAddr:         envString("GOCACHELAB_TCP_ADDR", "127.0.0.1:7379"),
		adminAddr:       envString("GOCACHELAB_ADMIN_ADDR", "127.0.0.1:8080"),
		dataDir:         dataDir,
		aofPath:         envString("GOCACHELAB_AOF_PATH", filepath.Join(dataDir, "appendonly.aof")),
		snapshotPath:    envString("GOCACHELAB_SNAPSHOT_PATH", filepath.Join(dataDir, "snapshot.json")),
		aofSyncPolicy:   policy,
		cleanupInterval: cleanupInterval,
		cleanupBatch:    envInt("GOCACHELAB_TTL_CLEANUP_BATCH", 1000),
		maxLineBytes:    envInt("GOCACHELAB_MAX_LINE_BYTES", server.DefaultMaxLineBytes),
		keyLimit:        envInt("GOCACHELAB_KEYS_LIMIT", server.DefaultKeyLimit),
	}, nil
}

func runTTLCleanup(ctx context.Context, cache *store.Store, interval time.Duration, batch int, logger *slog.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cleaned := cache.CleanupExpired(batch)
			if cleaned > 0 {
				logger.Info("ttl cleanup removed expired keys", "count", cleaned)
			}
		}
	}
}

func envString(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func envInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envDuration(name string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a Go duration: %w", name, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}
	return parsed, nil
}
