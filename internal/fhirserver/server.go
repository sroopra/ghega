// Package fhirserver implements a FHIR REST source connector for Ghega.
//
// It accepts POST, PUT, GET, and DELETE interactions on a configurable base
// path, validates Content-Type headers, persists received resources via
// messagestore.Store, and returns proper FHIR responses.
//
// Payload bytes are never logged; only metadata (resource type, count, status)
// is emitted.
package fhirserver

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sroopra/ghega/pkg/fhir"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/payloadref"
)

// Server is an http.Handler that exposes a FHIR REST endpoint.
type Server struct {
	BasePath  string
	Store     messagestore.Store
	Logger    *slog.Logger
	ChannelID string

	mu        sync.RWMutex
	resources map[string][]byte // keyed by resource ID for GET retrieval
}

// Option configures a Server.
type Option func(*Server)

// WithBasePath sets the FHIR base path (e.g. "/fhir/R4").
func WithBasePath(path string) Option {
	return func(s *Server) { s.BasePath = path }
}

// WithChannelID sets the channel ID used when persisting envelopes.
func WithChannelID(id string) Option {
	return func(s *Server) { s.ChannelID = id }
}

// WithLogger sets the structured logger.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) { s.Logger = logger }
}

// New creates a new FHIR server with the given store.
func New(store messagestore.Store, opts ...Option) *Server {
	s := &Server{
		BasePath:  "/fhir/R4",
		Store:     store,
		ChannelID: "fhir",
		resources: make(map[string][]byte),
	}
	for _, opt := range opts {
		opt(s)
	}
	if s.Logger == nil {
		s.Logger = slog.Default()
	}
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, s.BasePath) {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodPost, http.MethodPut:
		s.handleCreateOrUpdate(w, r)
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodDelete:
		s.handleDelete(w, r)
	default:
		writeOutcome(w, http.StatusMethodNotAllowed, "error", "Method not allowed")
	}
}

func (s *Server) handleCreateOrUpdate(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if !isValidFHIRContentType(ct) {
		writeOutcome(w, http.StatusBadRequest, "error",
			"Invalid Content-Type. Expected application/fhir+json or application/fhir+xml")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeOutcome(w, http.StatusBadRequest, "error", "Failed to read request body")
		return
	}
	defer r.Body.Close()

	// Attempt to parse as Bundle first.
	var bundle fhir.Bundle
	if err := json.Unmarshal(body, &bundle); err == nil && bundle.ResourceType == "Bundle" {
		s.handleBundle(w, r, &bundle, body)
		return
	}

	// Parse as an individual resource.
	var resource fhir.GenericResource
	if err := json.Unmarshal(body, &resource); err != nil {
		writeOutcome(w, http.StatusBadRequest, "error", "Invalid JSON: "+err.Error())
		return
	}
	if resource.ResourceType == "" {
		writeOutcome(w, http.StatusBadRequest, "error", "Missing resourceType")
		return
	}

	resourceID := resource.ID
	if resourceID == "" {
		resourceID = uuid.Must(uuid.NewV7()).String()
	}

	ctx := r.Context()
	messageID := uuid.Must(uuid.NewV7()).String()
	storageID := "fhir:" + messageID
	env := &payloadref.Envelope{
		ChannelID:  s.ChannelID,
		MessageID:  messageID,
		ReceivedAt: time.Now(),
		Status:     "received",
		Ref: payloadref.PayloadRef{
			StorageID: storageID,
			Location:  "fhir://resources/" + resource.ResourceType + "/" + resourceID,
		},
	}

	if err := s.Store.Save(ctx, env, body); err != nil {
		s.Logger.Error("failed to store FHIR resource",
			slog.String("resource_type", resource.ResourceType),
			slog.String("error", err.Error()),
		)
		writeOutcome(w, http.StatusInternalServerError, "error", "Failed to store resource")
		return
	}

	s.mu.Lock()
	s.resources[resourceID] = body
	s.mu.Unlock()

	s.Logger.Info("FHIR resource stored",
		slog.String("resource_type", resource.ResourceType),
		slog.String("resource_id", resourceID),
		slog.String("message_id", messageID),
	)

	w.Header().Set("Content-Type", "application/fhir+json")
	w.Header().Set("Location", s.BasePath+"/"+resource.ResourceType+"/"+resourceID)

	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	// Return the stored resource with the assigned ID.
	var respMap map[string]any
	_ = json.Unmarshal(body, &respMap)
	respMap["id"] = resourceID
	_ = json.NewEncoder(w).Encode(respMap)
}

func (s *Server) handleBundle(w http.ResponseWriter, r *http.Request, bundle *fhir.Bundle, rawBody []byte) {
	ctx := r.Context()
	storedCount := 0

	for _, entry := range bundle.Entry {
		var resource fhir.GenericResource
		if err := json.Unmarshal(entry.Resource, &resource); err != nil {
			s.Logger.Warn("failed to parse bundle entry",
				slog.String("error", err.Error()),
			)
			continue
		}
		if resource.ResourceType == "" {
			continue
		}

		resourceID := resource.ID
		if resourceID == "" {
			resourceID = uuid.Must(uuid.NewV7()).String()
		}

		messageID := uuid.Must(uuid.NewV7()).String()
		storageID := "fhir:" + messageID
		env := &payloadref.Envelope{
			ChannelID:  s.ChannelID,
			MessageID:  messageID,
			ReceivedAt: time.Now(),
			Status:     "received",
			Ref: payloadref.PayloadRef{
				StorageID: storageID,
				Location:  "fhir://resources/" + resource.ResourceType + "/" + resourceID,
			},
		}

		if err := s.Store.Save(ctx, env, entry.Resource); err != nil {
			s.Logger.Error("failed to store bundle entry",
				slog.String("resource_type", resource.ResourceType),
				slog.String("error", err.Error()),
			)
			continue
		}

		s.mu.Lock()
		s.resources[resourceID] = entry.Resource
		s.mu.Unlock()

		storedCount++
	}

	s.Logger.Info("FHIR bundle stored",
		slog.Int("entry_count", len(bundle.Entry)),
		slog.Int("stored_count", storedCount),
	)

	w.Header().Set("Content-Type", "application/fhir+json")
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"resourceType": "Bundle",
		"type":         bundle.Type,
		"entry":        len(bundle.Entry),
	})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, s.BasePath)
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		writeOutcome(w, http.StatusBadRequest, "error", "Invalid resource path")
		return
	}
	resourceID := parts[1]

	s.mu.RLock()
	data, ok := s.resources[resourceID]
	s.mu.RUnlock()

	if !ok {
		writeOutcome(w, http.StatusNotFound, "error", "Resource not found")
		return
	}

	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, s.BasePath)
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		writeOutcome(w, http.StatusBadRequest, "error", "Invalid resource path")
		return
	}
	resourceID := parts[1]

	s.mu.Lock()
	delete(s.resources, resourceID)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusNoContent)
}

func isValidFHIRContentType(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.Contains(ct, "application/fhir+json") || strings.Contains(ct, "application/fhir+xml")
}

func writeOutcome(w http.ResponseWriter, status int, severity, diagnostics string) {
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(status)
	outcome := fhir.OperationOutcome{
		ResourceType: "OperationOutcome",
		Issue: []fhir.OutcomeIssue{
			{
				Severity:    severity,
				Code:        "processing",
				Diagnostics: diagnostics,
			},
		},
	}
	_ = json.NewEncoder(w).Encode(outcome)
}
