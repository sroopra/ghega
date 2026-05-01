// Package fhirserver implements a FHIR REST source connector for Ghega.
//
// It exposes an HTTP handler that accepts FHIR R4 JSON resources and stores
// them via messagestore.Store. Payload bytes are never logged; only metadata
// such as method, path, resource type, entry count, and status code are logged.
package fhirserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/sroopra/ghega/pkg/fhir"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/payloadref"
)

// Config configures the FHIR server handler.
type Config struct {
	BasePath string
	Store    messagestore.Store
	Logger   *slog.Logger
}

// NewHandler returns an http.Handler that serves FHIR REST interactions.
func NewHandler(cfg Config) http.Handler {
	mux := http.NewServeMux()
	h := &handler{
		basePath: strings.TrimSuffix(cfg.BasePath, "/"),
		store:    cfg.Store,
		logger:   cfg.Logger,
	}
	if h.logger == nil {
		h.logger = slog.Default()
	}

	base := h.basePath
	mux.HandleFunc(base, h.handleBase)
	mux.HandleFunc(base+"/{resourceType}/{id}", h.handleResource)

	return mux
}

type handler struct {
	basePath string
	store    messagestore.Store
	logger   *slog.Logger
}

func (h *handler) handleBase(w http.ResponseWriter, r *http.Request) {
	if !h.validateContentType(w, r) {
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.handlePost(w, r, "", "")
	default:
		writeOperationOutcome(w, http.StatusMethodNotAllowed, "method not allowed at base path")
	}
}

func (h *handler) handleResource(w http.ResponseWriter, r *http.Request) {
	// GET and DELETE do not require a Content-Type header.
	if r.Method != http.MethodGet && r.Method != http.MethodDelete {
		if !h.validateContentType(w, r) {
			return
		}
	}

	resourceType := r.PathValue("resourceType")
	id := r.PathValue("id")

	switch r.Method {
	case http.MethodPost:
		h.handlePost(w, r, resourceType, id)
	case http.MethodPut:
		h.handlePut(w, r, resourceType, id)
	case http.MethodGet:
		h.handleGet(w, r, resourceType, id)
	case http.MethodDelete:
		h.handleDelete(w, r, resourceType, id)
	default:
		writeOperationOutcome(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *handler) validateContentType(w http.ResponseWriter, r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		// Empty Content-Type is only acceptable for GET/DELETE, which are
		// filtered before this function is called for other methods.
		return true
	}
	if strings.Contains(ct, "application/fhir+json") ||
		strings.Contains(ct, "application/fhir+xml") ||
		strings.Contains(ct, "application/json") {
		return true
	}
	writeOperationOutcome(w, http.StatusBadRequest, "unsupported Content-Type: expected application/fhir+json")
	return false
}

func (h *handler) handlePost(w http.ResponseWriter, r *http.Request, resourceType, id string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeOperationOutcome(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var resourceCheck struct {
		ResourceType string `json:"resourceType"`
	}
	if err := json.Unmarshal(body, &resourceCheck); err != nil {
		writeOperationOutcome(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if resourceCheck.ResourceType == "Bundle" {
		h.handleBundle(w, r, body)
		return
	}

	// Individual resource.
	if resourceType == "" {
		resourceType = resourceCheck.ResourceType
	}
	if id == "" {
		var idCheck struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(body, &idCheck)
		id = idCheck.ID
	}
	if id == "" {
		writeOperationOutcome(w, http.StatusBadRequest, "missing resource id")
		return
	}

	if err := h.saveResource(r.Context(), resourceType, id, body); err != nil {
		writeOperationOutcome(w, http.StatusInternalServerError, "failed to store resource: "+err.Error())
		return
	}

	location := fmt.Sprintf("%s/%s/%s", h.basePath, resourceType, id)
	w.Header().Set("Content-Type", "application/fhir+json")
	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusCreated)
	h.logger.Info("FHIR resource stored",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("resource_type", resourceType),
		slog.Int("status_code", http.StatusCreated),
	)
}

func (h *handler) handlePut(w http.ResponseWriter, r *http.Request, resourceType, id string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeOperationOutcome(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	if resourceType == "" || id == "" {
		writeOperationOutcome(w, http.StatusBadRequest, "missing resource type or id")
		return
	}

	if err := h.saveResource(r.Context(), resourceType, id, body); err != nil {
		writeOperationOutcome(w, http.StatusInternalServerError, "failed to store resource: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusOK)
	h.logger.Info("FHIR resource updated",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("resource_type", resourceType),
		slog.Int("status_code", http.StatusOK),
	)
}

func (h *handler) handleGet(w http.ResponseWriter, r *http.Request, resourceType, id string) {
	if resourceType == "" || id == "" {
		writeOperationOutcome(w, http.StatusBadRequest, "missing resource type or id")
		return
	}

	messageID := resourceType + "/" + id
	env, err := h.store.GetMetadata(r.Context(), messageID)
	if err != nil {
		if _, ok := err.(*messagestore.ErrNotFound); ok {
			writeOperationOutcome(w, http.StatusNotFound, "resource not found")
			return
		}
		writeOperationOutcome(w, http.StatusInternalServerError, "failed to get resource: "+err.Error())
		return
	}

	payload, ok, err := getPayload(r.Context(), h.store, env.Ref.StorageID)
	if err != nil {
		writeOperationOutcome(w, http.StatusInternalServerError, "failed to retrieve payload: "+err.Error())
		return
	}
	if !ok {
		writeOperationOutcome(w, http.StatusNotFound, "resource payload not found")
		return
	}

	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
	h.logger.Info("FHIR resource retrieved",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("resource_type", resourceType),
		slog.Int("status_code", http.StatusOK),
	)
}

func (h *handler) handleDelete(w http.ResponseWriter, r *http.Request, resourceType, id string) {
	if resourceType == "" || id == "" {
		writeOperationOutcome(w, http.StatusBadRequest, "missing resource type or id")
		return
	}

	messageID := resourceType + "/" + id
	deleter, ok := h.store.(interface {
		Delete(ctx context.Context, messageID string) error
	})
	if !ok {
		writeOperationOutcome(w, http.StatusInternalServerError, "store does not support deletion")
		return
	}

	if err := deleter.Delete(r.Context(), messageID); err != nil {
		if _, ok := err.(*messagestore.ErrNotFound); ok {
			writeOperationOutcome(w, http.StatusNotFound, "resource not found")
			return
		}
		writeOperationOutcome(w, http.StatusInternalServerError, "failed to delete resource: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
	h.logger.Info("FHIR resource deleted",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("resource_type", resourceType),
		slog.Int("status_code", http.StatusNoContent),
	)
}

func (h *handler) handleBundle(w http.ResponseWriter, r *http.Request, body []byte) {
	bundle, err := fhir.ParseBundle(body)
	if err != nil {
		writeOperationOutcome(w, http.StatusBadRequest, "invalid Bundle: "+err.Error())
		return
	}

	if err := fhir.ValidateBundleType(bundle); err != nil {
		writeOperationOutcome(w, http.StatusBadRequest, err.Error())
		return
	}

	entryCount := 0
	err = fhir.IterateEntries(bundle, func(entry fhir.BundleEntry) error {
		entryCount++
		if len(entry.Resource) == 0 {
			return nil
		}

		var res struct {
			ResourceType string `json:"resourceType"`
			ID           string `json:"id"`
		}
		if err := json.Unmarshal(entry.Resource, &res); err != nil {
			return fmt.Errorf("invalid entry resource: %w", err)
		}
		if res.ResourceType == "" || res.ID == "" {
			return fmt.Errorf("entry missing resourceType or id")
		}

		if err := h.saveResource(r.Context(), res.ResourceType, res.ID, []byte(entry.Resource)); err != nil {
			return fmt.Errorf("store entry: %w", err)
		}
		return nil
	})
	if err != nil {
		writeOperationOutcome(w, http.StatusBadRequest, "bundle processing failed: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(http.StatusOK)
	h.logger.Info("FHIR Bundle processed",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int("entry_count", entryCount),
		slog.Int("status_code", http.StatusOK),
	)
}

func (h *handler) saveResource(ctx context.Context, resourceType, id string, payload []byte) error {
	messageID := resourceType + "/" + id
	storageID := messageID
	envelope := &payloadref.Envelope{
		ChannelID:  "fhir",
		MessageID:  messageID,
		ReceivedAt: time.Now(),
		Status:     "received",
		Ref: payloadref.PayloadRef{
			StorageID: storageID,
			Location:  "fhirserver",
		},
	}
	return h.store.Save(ctx, envelope, payload)
}

// getPayload retrieves payload bytes from a store using capability interfaces.
// The messagestore package has two implementations with different signatures.
func getPayload(ctx context.Context, store messagestore.Store, storageID string) ([]byte, bool, error) {
	if pg, ok := store.(interface {
		GetPayload(ctx context.Context, storageID string) ([]byte, bool, error)
	}); ok {
		return pg.GetPayload(ctx, storageID)
	}
	if pg, ok := store.(interface {
		GetPayload(storageID string) ([]byte, bool)
	}); ok {
		data, found := pg.GetPayload(storageID)
		return data, found, nil
	}
	return nil, false, fmt.Errorf("store does not support payload retrieval")
}

func writeOperationOutcome(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/fhir+json")
	w.WriteHeader(status)
	oo := fhir.OperationOutcome{
		ResourceType: "OperationOutcome",
		Issue: []fhir.OperationOutcomeIssue{
			{
				Severity:    "error",
				Code:        "invalid",
				Diagnostics: message,
			},
		},
	}
	_ = json.NewEncoder(w).Encode(oo)
}
