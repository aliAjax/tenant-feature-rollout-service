package server

import (
	"context"
	"encoding/json"
	d "example.com/feature-rollout-control/internal/shared/domain"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HTTP struct {
	svc *Service
	log *slog.Logger
}

func NewHTTP(s *Service, l *slog.Logger) *HTTP { return &HTTP{svc: s, log: l} }
func (h *HTTP) Routes() *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("/healthz", h.health)
	m.HandleFunc("/readyz", h.health)
	m.HandleFunc("/metrics", h.metrics)
	m.HandleFunc("/v1/projects", h.projects)
	m.HandleFunc("/v1/environments", h.environments)
	m.HandleFunc("/v1/flags", h.flags)
	m.HandleFunc("/v1/flags/", h.flags)
	m.HandleFunc("/v1/evaluate", h.evaluate)
	m.HandleFunc("/v1/changes", h.changes)
	return m
}
func (h *HTTP) Middleware(next http.Handler) http.Handler {
	return requestID(recoverer(timeout(logger(cors(next)))))
}
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = fmt.Sprintf("req-%d", time.Now().UnixNano())
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}
func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Default().Info("http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String())
	})
}
func timeout(next http.Handler) http.Handler {
	return http.TimeoutHandler(next, 10*time.Second, "request timeout")
}
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				writeErr(w, http.StatusInternalServerError, fmt.Errorf("panic: %v", v))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Tenant-ID")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (h *HTTP) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"status": "ok", "time": time.Now().UTC()})
}
func (h *HTTP) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte("# HELP rollout_up service status\n# TYPE rollout_up gauge\nrollout_up 1\n"))
}
func tenant(r *http.Request) string {
	t := r.Header.Get("X-Tenant-ID")
	if t == "" {
		t = "default"
	}
	return t
}
func decode(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(v)
}
func (h *HTTP) projects(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeErr(w, 405, fmt.Errorf("method not allowed"))
		return
	}
	var in struct {
		Name string `json:"name"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, err)
		return
	}
	p, err := h.svc.CreateProject(r.Context(), tenant(r), in.Name)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 201, p)
}
func (h *HTTP) environments(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeErr(w, 405, fmt.Errorf("method not allowed"))
		return
	}
	var in struct {
		ProjectID string `json:"project_id"`
		Name      string `json:"name"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, err)
		return
	}
	e, err := h.svc.CreateEnvironment(r.Context(), tenant(r), in.ProjectID, in.Name)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 201, e)
}
func (h *HTTP) flags(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && strings.Trim(r.URL.Path, "/") == "v1/flags" {
		var in struct {
			ProjectID     string      `json:"project_id"`
			EnvironmentID string      `json:"environment_id"`
			Key           string      `json:"key"`
			Type          d.ValueType `json:"type"`
			Default       any         `json:"default"`
		}
		if err := decode(r, &in); err != nil {
			writeErr(w, 400, err)
			return
		}
		f, err := h.svc.CreateFlag(r.Context(), tenant(r), in.ProjectID, in.EnvironmentID, in.Key, in.Type, in.Default)
		if err != nil {
			writeErr(w, 400, err)
			return
		}
		writeJSON(w, 201, f)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		writeErr(w, 404, fmt.Errorf("flag route not found"))
		return
	}
	project, key := parts[2], parts[4]
	action := parts[5]
	var body struct {
		Actor   string `json:"actor"`
		Reason  string `json:"reason"`
		Version int64  `json:"version"`
		Rule    d.Rule `json:"rule"`
	}
	if r.Method == "POST" {
		_ = decode(r, &body)
	}
	var (
		f   d.Flag
		err error
	)
	switch action {
	case "publish":
		f, err = h.svc.Publish(r.Context(), tenant(r), project, key, body.Actor, body.Reason)
	case "pause":
		f, err = h.svc.Pause(r.Context(), tenant(r), project, key, body.Actor, body.Reason)
	case "archive":
		f, err = h.svc.Archive(r.Context(), tenant(r), project, key, body.Actor, body.Reason)
	case "rollback":
		f, err = h.svc.Rollback(r.Context(), tenant(r), project, key, body.Version, body.Actor, body.Reason)
	case "rules":
		f, err = h.svc.AddRule(r.Context(), tenant(r), project, key, body.Rule, body.Actor, body.Reason)
	default:
		writeErr(w, 404, fmt.Errorf("unknown flag action"))
		return
	}
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, f)
}
func (h *HTTP) flagPath(w http.ResponseWriter, r *http.Request) (d.Flag, bool) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		return d.Flag{}, false
	}
	f, err := h.svc.GetFlag(r.Context(), tenant(r), parts[2], parts[4])
	if err != nil {
		writeErr(w, 404, err)
		return d.Flag{}, false
	}
	return f, true
}
func (h *HTTP) evaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeErr(w, 405, fmt.Errorf("method not allowed"))
		return
	}
	var in struct {
		ProjectID     string         `json:"project_id"`
		EnvironmentID string         `json:"environment_id"`
		Key           string         `json:"key"`
		UserID        string         `json:"user_id"`
		Keys          []string       `json:"keys"`
		Attributes    map[string]any `json:"attributes"`
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, err)
		return
	}
	ec := d.EvaluationContext{TenantID: tenant(r), ProjectID: in.ProjectID, EnvironmentID: in.EnvironmentID, UserID: in.UserID, Attributes: in.Attributes}
	if len(in.Keys) > 0 {
		v, err := h.svc.EvaluateBatch(r.Context(), tenant(r), in.ProjectID, in.Keys, ec)
		if err != nil {
			writeErr(w, 400, err)
			return
		}
		writeJSON(w, 200, v)
		return
	}
	v, err := h.svc.Evaluate(r.Context(), tenant(r), in.ProjectID, in.Key, ec)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	writeJSON(w, 200, v)
}
func (h *HTTP) changes(w http.ResponseWriter, r *http.Request) {
	cur, _ := strconv.ParseInt(r.URL.Query().Get("cursor"), 10, 64)
	v, err := h.svc.Changes(r.Context(), tenant(r), cur)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, v)
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}
func Start(ctx context.Context, addr string, svc *Service, log *slog.Logger) error {
	srv := &http.Server{Addr: addr, Handler: NewHTTP(svc, log).Middleware(NewHTTP(svc, log).Routes()), ReadHeaderTimeout: 3 * time.Second}
	go func() {
		<-ctx.Done()
		shut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shut)
	}()
	log.Info("server listening", "address", addr)
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
