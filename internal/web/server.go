package web

import (
	"context"
	_ "embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"git.zk35.de/secalpha/judo2mqtt/internal/config"
	"git.zk35.de/secalpha/judo2mqtt/internal/state"
)

//go:embed ui.html
var uiHTML []byte

type Server struct {
	st       *state.State
	version  string
	cfg      *config.Config
	onConfig func(config.FileConfig) error
	onValve  func(ctx context.Context, action string) (string, error)
}

func New(st *state.State, version string, cfg *config.Config, onConfig func(config.FileConfig) error, onValve func(ctx context.Context, action string) (string, error)) *Server {
	return &Server{st: st, version: version, cfg: cfg, onConfig: onConfig, onValve: onValve}
}

func (s *Server) Start(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleUI)
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /api/status", s.handleStatus)
	mux.HandleFunc("GET /api/config", s.handleGetConfig)
	mux.HandleFunc("POST /api/config", s.handlePostConfig)
	mux.HandleFunc("POST /api/valve", s.handleValve)

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	slog.Info("web UI listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) handleUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(uiHTML)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": s.version})
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	snap := s.st.Snapshot()
	snap.ConfigReady = s.cfg != nil && s.cfg.IsComplete()
	writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleGetConfig(w http.ResponseWriter, _ *http.Request) {
	if s.cfg == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "config not available"})
		return
	}
	writeJSON(w, http.StatusOK, config.ToFileConfig(s.cfg))
}

func (s *Server) handlePostConfig(w http.ResponseWriter, r *http.Request) {
	if s.cfg == nil || s.onConfig == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "config management not available"})
		return
	}

	var fc config.FileConfig
	if err := json.NewDecoder(r.Body).Decode(&fc); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	if fc.JudoHost == "" || fc.JudoSerial == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "judo_host and judo_serial are required"})
		return
	}

	if err := config.Save(s.cfg.ConfigFile, fc); err != nil {
		slog.Error("config save failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save failed: " + err.Error()})
		return
	}

	if err := s.onConfig(fc); err != nil {
		slog.Error("config apply failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "apply failed: " + err.Error()})
		return
	}

	slog.Info("config saved and applied")
	writeJSON(w, http.StatusOK, map[string]string{"status": "applied"})
}

func (s *Server) handleValve(w http.ResponseWriter, r *http.Request) {
	if s.onValve == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "valve control not available"})
		return
	}
	var req struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.Action != "open" && req.Action != "close" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be open or close"})
		return
	}
	result, err := s.onValve(r.Context(), req.Action)
	if err != nil {
		slog.Error("valve command failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "valve": result})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
