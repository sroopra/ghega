// Package server implements the Ghega HTTP API and BFF middleware.
package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sroopra/ghega"
	"github.com/sroopra/ghega/internal/alerts"
	"github.com/sroopra/ghega/internal/config"
	"github.com/sroopra/ghega/internal/session"
	"github.com/sroopra/ghega/pkg/channelstore"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/migration"
	"github.com/sroopra/ghega/pkg/payloadref"
	"gopkg.in/yaml.v3"
)

// Server holds dependencies for the HTTP API.
type Server struct {
	store         messagestore.Store
	alertStore    alerts.AlertStore
	channelStore  channelstore.ChannelStore
	migrationsDir string
	authConfig    config.AuthConfig
	sessionMgr    *session.Manager
	oidcProvider  *OIDCProvider
}

// ServerOption configures a Server.
type ServerOption func(*Server)

// WithAuthConfig sets the authentication configuration.
func WithAuthConfig(cfg config.AuthConfig) ServerOption {
	return func(s *Server) { s.authConfig = cfg }
}

// WithSessionManager sets the session manager.
func WithSessionManager(mgr *session.Manager) ServerOption {
	return func(s *Server) { s.sessionMgr = mgr }
}

// WithOIDCProvider sets the OIDC provider.
func WithOIDCProvider(op *OIDCProvider) ServerOption {
	return func(s *Server) { s.oidcProvider = op }
}

// WithChannelStore sets the channel store for listing deployed channels.
func WithChannelStore(cs channelstore.ChannelStore) ServerOption {
	return func(s *Server) { s.channelStore = cs }
}

// New creates a new Server with the given store and alertStore.
func New(store messagestore.Store, alertStore alerts.AlertStore, opts ...ServerOption) *Server {
	s := &Server{store: store, alertStore: alertStore}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// SetMigrationsDir configures the directory where migration reports are read from.
func (s *Server) SetMigrationsDir(dir string) {
	s.migrationsDir = dir
}

// messageMetadataResponse is the JSON shape expected by the UI.
type messageMetadataResponse struct {
	ID         string `json:"id"`
	ChannelID  string `json:"channel_id"`
	MessageID  string `json:"message_id"`
	Status     string `json:"status"`
	ReceivedAt string `json:"received_at"`
	StorageID  string `json:"storage_id"`
	Location   string `json:"location"`
}

// channelResponse is the JSON shape for a channel.
type channelResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// migrationListItem is the JSON shape for a migration summary entry.
type migrationListItem struct {
	ID                string `json:"id"`
	ChannelName       string `json:"channel_name"`
	OriginalName      string `json:"original_name"`
	Status            string `json:"status"`
	RewriteTasksCount int    `json:"rewrite_tasks_count"`
	WarningsCount     int    `json:"warnings_count"`
}

// migrationReportResponse is the JSON shape for a full migration report.
type migrationReportResponse struct {
	ChannelName   string                  `json:"channel_name"`
	OriginalName  string                  `json:"original_name"`
	Status        string                  `json:"status"`
	AutoConverted []migrationConvertedItem  `json:"auto_converted"`
	NeedsRewrite  []migrationRewriteTaskItem `json:"needs_rewrite"`
	Unsupported   []migrationUnsupportedItem `json:"unsupported"`
	Warnings      []string                `json:"warnings"`
}

type migrationConvertedItem struct {
	Element     string `json:"element"`
	Description string `json:"description"`
}

type migrationRewriteTaskItem struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Category    string `json:"category,omitempty"`
}

type migrationUnsupportedItem struct {
	Feature     string `json:"feature"`
	Description string `json:"description"`
}

func envelopeToResponse(env *payloadref.Envelope) messageMetadataResponse {
	return messageMetadataResponse{
		ID:         env.MessageID,
		ChannelID:  env.ChannelID,
		MessageID:  env.MessageID,
		Status:     env.Status,
		ReceivedAt: env.ReceivedAt.Format(time.RFC3339Nano),
		StorageID:  env.Ref.StorageID,
		Location:   env.Ref.Location,
	}
}

// Handler returns the HTTP handler for the Ghega API.
func (s *Server) Handler() http.Handler {
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("GET /messages", s.handleListMessages)
	apiMux.HandleFunc("GET /messages/{id}", s.handleGetMessage)
	apiMux.HandleFunc("POST /messages/{id}/redeliver", s.handleRedeliver)
	apiMux.HandleFunc("POST /messages/{id}/replay", s.handleReplay)
	apiMux.HandleFunc("GET /channels", s.handleListChannels)
	apiMux.HandleFunc("GET /alerts", s.handleListAlerts)
	apiMux.HandleFunc("GET /migrations", s.handleListMigrations)
	apiMux.HandleFunc("GET /migrations/{id}", s.handleGetMigration)
	apiMux.HandleFunc("GET /me", s.handleMe)

	wrapped := s.CORSMiddleware(s.CSRFMiddleware(s.AuthMiddleware(apiMux)))

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", wrapped))

	if s.oidcProvider != nil {
		mux.HandleFunc("GET /auth/login", s.oidcProvider.HandleLogin)
		mux.HandleFunc("GET /auth/callback", s.oidcProvider.HandleCallback)
		mux.HandleFunc("GET /auth/logout", s.oidcProvider.HandleLogout)
		mux.HandleFunc("POST /auth/logout", s.oidcProvider.HandleLogout)
	}

	mux.HandleFunc("/healthz", s.handleHealthz)

	sub, err := fs.Sub(ghega.UIFS, "ui/dist")
	if err != nil {
		panic("ui/dist not embedded: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))
	mux.Handle("/", fileServer)

	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleListMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	channelID := r.URL.Query().Get("channel_id")

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	var envelopes []*payloadref.Envelope
	var err error
	if channelID != "" {
		envelopes, err = s.store.ListByChannel(ctx, channelID, limit, offset)
	} else {
		envelopes, err = s.store.ListAll(ctx, limit, offset)
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list messages")
		return
	}

	resp := make([]messageMetadataResponse, len(envelopes))
	for i, env := range envelopes {
		resp[i] = envelopeToResponse(env)
	}
	writeJSON(w, resp)
}

func (s *Server) handleGetMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "missing message id")
		return
	}

	env, err := s.store.GetMetadata(ctx, id)
	if err != nil {
		if _, ok := err.(*messagestore.ErrNotFound); ok {
			writeJSONError(w, http.StatusNotFound, "message not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to get message")
		return
	}

	writeJSON(w, envelopeToResponse(env))
}

func (s *Server) handleListChannels(w http.ResponseWriter, r *http.Request) {
	if s.channelStore == nil {
		writeJSON(w, []channelResponse{})
		return
	}

	records, err := s.channelStore.ListChannels(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list channels")
		return
	}

	resp := make([]channelResponse, len(records))
	for i, rec := range records {
		resp[i] = channelResponse{
			ID:   rec.Name,
			Name: rec.Name,
		}
	}
	writeJSON(w, resp)
}

func (s *Server) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	alertsList, err := s.alertStore.List()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}
	writeJSON(w, alertsList)
}

func (s *Server) handleRedeliver(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusNotImplemented, "not yet implemented")
}

func (s *Server) handleReplay(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusNotImplemented, "not yet implemented")
}

func (s *Server) handleListMigrations(w http.ResponseWriter, r *http.Request) {
	if s.migrationsDir == "" {
		writeJSON(w, []migrationListItem{})
		return
	}

	entries, err := os.ReadDir(s.migrationsDir)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to read migrations directory")
		return
	}

	var items []migrationListItem
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		reportPath := filepath.Join(s.migrationsDir, entry.Name(), "migration-report.yaml")
		data, err := os.ReadFile(reportPath)
		if err != nil {
			continue
		}
		var report migration.ChannelMigrationReport
		if err := yaml.Unmarshal(data, &report); err != nil {
			continue
		}
		items = append(items, migrationListItem{
			ID:                entry.Name(),
			ChannelName:       report.ChannelName,
			OriginalName:      report.OriginalName,
			Status:            report.Status,
			RewriteTasksCount: len(report.NeedsRewrite),
			WarningsCount:     len(report.Warnings),
		})
	}

	writeJSON(w, items)
}

func (s *Server) handleGetMigration(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "missing migration id")
		return
	}

	if s.migrationsDir == "" {
		writeJSONError(w, http.StatusNotFound, "migration not found")
		return
	}

	reportPath := filepath.Join(s.migrationsDir, id, "migration-report.yaml")
	data, err := os.ReadFile(reportPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSONError(w, http.StatusNotFound, "migration not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to read migration report")
		return
	}

	var report migration.ChannelMigrationReport
	if err := yaml.Unmarshal(data, &report); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to parse migration report")
		return
	}

	resp := migrationReportResponse{
		ChannelName:  report.ChannelName,
		OriginalName: report.OriginalName,
		Status:       report.Status,
		Warnings:     report.Warnings,
	}
	for _, c := range report.AutoConverted {
		resp.AutoConverted = append(resp.AutoConverted, migrationConvertedItem{
			Element:     c.Element,
			Description: c.Description,
		})
	}
	for _, t := range report.NeedsRewrite {
		resp.NeedsRewrite = append(resp.NeedsRewrite, migrationRewriteTaskItem{
			Severity:    t.Severity,
			Description: t.Description,
			Category:    t.Category,
		})
	}
	for _, u := range report.Unsupported {
		resp.Unsupported = append(resp.Unsupported, migrationUnsupportedItem{
			Feature:     u.Feature,
			Description: u.Description,
		})
	}

	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
