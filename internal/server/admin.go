package server

import (
	"encoding/json"
	"net/http"
	"net/http/pprof"

	"github.com/Defyland/gocachelab-redis-like-store/internal/metrics"
	"github.com/Defyland/gocachelab-redis-like-store/internal/store"
)

func NewAdminHandler(store *store.Store, metricsRegistry *metrics.Metrics) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		stats := store.Stats()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(metrics.Prometheus(metricsRegistry.Snapshot(), metrics.StoreStats{
			LiveKeys:     stats.LiveKeys,
			PhysicalKeys: stats.PhysicalKeys,
			ExpiredTotal: stats.ExpiredTotal,
		})))
	})

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
