package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/payloadref"
)

func setupTestServer(devAuth bool) (*Server, *messagestore.InMemoryStore, *InMemoryChannelRegistry) {
	store := messagestore.NewInMemoryStore()
	registry := NewInMemoryChannelRegistry()
	server := NewServer(store, registry)
	return server, store, registry
}

func TestHealthz(t *testing.T) {
	server, _, _ := setupTestServer(false)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	server.Healthz(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"status":"ok"`) {
		t.Errorf("expected body to contain status ok, got: %s", body)
	}
}

func TestListMessages(t *testing.T) {
	server, store, _ := setupTestServer(false)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		env := &payloadref.Envelope{
			ChannelID:  "ch-test",
			MessageID:  fmt.Sprintf("msg-%03d", i),
			ReceivedAt: time.Date(2024, 1, 1, 0, 0, i, 0, time.UTC),
			Status:     "received",
			Ref: payloadref.PayloadRef{
				StorageID: fmt.Sprintf("store-%03d", i),
				Location:  fmt.Sprintf("mem://ghega/%03d", i),
			},
		}
		if err := store.Save(ctx, env, []byte(fmt.Sprintf("payload-%d", i))); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	rec := httptest.NewRecorder()
	server.ListMessages(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp PaginatedMessagesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp.Messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(resp.Messages))
	}
	for _, msg := range resp.Messages {
		if msg.MessageID == "" {
			t.Error("expected non-empty message_id")
		}
		if msg.StorageID == "" {
			t.Error("expected non-empty storage_id")
		}
	}
}

func TestListMessages_Pagination(t *testing.T) {
	server, store, _ := setupTestServer(false)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		env := &payloadref.Envelope{
			ChannelID:  "ch-test",
			MessageID:  fmt.Sprintf("msg-%03d", i),
			ReceivedAt: time.Date(2024, 1, 1, 0, 0, i, 0, time.UTC),
			Status:     "received",
			Ref: payloadref.PayloadRef{
				StorageID: fmt.Sprintf("store-%03d", i),
				Location:  fmt.Sprintf("mem://ghega/%03d", i),
			},
		}
		if err := store.Save(ctx, env, []byte(fmt.Sprintf("payload-%d", i))); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages?limit=2&offset=1", nil)
	rec := httptest.NewRecorder()
	server.ListMessages(rec, req)

	var resp PaginatedMessagesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Limit != 2 {
		t.Errorf("expected limit 2, got %d", resp.Limit)
	}
	if resp.Offset != 1 {
		t.Errorf("expected offset 1, got %d", resp.Offset)
	}
	if len(resp.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(resp.Messages))
	}
}

func TestGetMessage(t *testing.T) {
	server, store, _ := setupTestServer(false)
	ctx := context.Background()

	env := &payloadref.Envelope{
		ChannelID:  "ch-test",
		MessageID:  "msg-001",
		ReceivedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:     "received",
		Ref: payloadref.PayloadRef{
			StorageID: "store-001",
			Location:  "mem://ghega/001",
		},
	}
	if err := store.Save(ctx, env, []byte("synthetic-payload-data")); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/msg-001", nil)
	rec := httptest.NewRecorder()
	server.GetMessage(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var msg MessageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &msg); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if msg.MessageID != "msg-001" {
		t.Errorf("expected message_id msg-001, got %s", msg.MessageID)
	}
	if msg.ChannelID != "ch-test" {
		t.Errorf("expected channel_id ch-test, got %s", msg.ChannelID)
	}
	if msg.Status != "received" {
		t.Errorf("expected status received, got %s", msg.Status)
	}
}

func TestGetMessage_NotFound(t *testing.T) {
	server, _, _ := setupTestServer(false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/nonexistent", nil)
	rec := httptest.NewRecorder()
	server.GetMessage(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestGetMessage_MissingID(t *testing.T) {
	server, _, _ := setupTestServer(false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/", nil)
	rec := httptest.NewRecorder()
	server.GetMessage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestListChannels(t *testing.T) {
	server, _, registry := setupTestServer(false)

	registry.Register(&ChannelConfig{
		Name: "ch-1",
		Source: SourceConfig{
			Type: "mllp",
			Host: "0.0.0.0",
			Port: 2575,
		},
		Destination: DestinationConfig{
			Type: "http",
			URL:  "http://example.com/webhook",
		},
		Mapping: MappingConfig{
			MessageType: "ADT_A01",
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels", nil)
	rec := httptest.NewRecorder()
	server.ListChannels(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var channels []*ChannelConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &channels); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(channels) != 1 {
		t.Errorf("expected 1 channel, got %d", len(channels))
	}
	if channels[0].Name != "ch-1" {
		t.Errorf("expected channel name ch-1, got %s", channels[0].Name)
	}
}

func TestGetChannel(t *testing.T) {
	server, _, registry := setupTestServer(false)

	registry.Register(&ChannelConfig{
		Name: "ch-1",
		Source: SourceConfig{
			Type: "mllp",
			Host: "0.0.0.0",
			Port: 2575,
		},
		Destination: DestinationConfig{
			Type: "http",
			URL:  "http://example.com/webhook",
		},
		Mapping: MappingConfig{
			MessageType: "ADT_A01",
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/ch-1", nil)
	rec := httptest.NewRecorder()
	server.GetChannel(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var ch ChannelConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &ch); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if ch.Name != "ch-1" {
		t.Errorf("expected channel name ch-1, got %s", ch.Name)
	}
}

func TestGetChannel_NotFound(t *testing.T) {
	server, _, _ := setupTestServer(false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/nonexistent", nil)
	rec := httptest.NewRecorder()
	server.GetChannel(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestGetChannel_MissingID(t *testing.T) {
	server, _, _ := setupTestServer(false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/", nil)
	rec := httptest.NewRecorder()
	server.GetChannel(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestAuthMiddleware_AcceptsValidToken(t *testing.T) {
	middleware := AuthMiddleware(false)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			t.Error("expected user in context")
		}
		if user.Role != "operator" {
			t.Errorf("expected role operator, got %s", user.Role)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestAuthMiddleware_RejectsMissingToken(t *testing.T) {
	middleware := AuthMiddleware(false)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_RejectsInvalidFormat(t *testing.T) {
	middleware := AuthMiddleware(false)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_RejectsMalformedJWT(t *testing.T) {
	middleware := AuthMiddleware(false)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with malformed JWT")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-jwt")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_DevAuthBypass(t *testing.T) {
	middleware := AuthMiddleware(true)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			t.Error("expected user in context")
		}
		if user.Subject != "dev-user" {
			t.Errorf("expected subject dev-user, got %s", user.Subject)
		}
		if user.Role != "admin" {
			t.Errorf("expected role admin, got %s", user.Role)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestRouter_WithAuth(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	registry := NewInMemoryChannelRegistry()
	router := NewRouter(store, registry, false)

	// Healthz should work without auth.
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("healthz expected status 200, got %d", rec.Code)
	}

	// API without auth should return 401.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("api without auth expected status 401, got %d", rec.Code)
	}

	// API with valid auth should work.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("api with auth expected status 200, got %d", rec.Code)
	}
}

func TestRouter_DevAuthBypass(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	registry := NewInMemoryChannelRegistry()
	router := NewRouter(store, registry, true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("api with dev-auth expected status 200, got %d", rec.Code)
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	registry := NewInMemoryChannelRegistry()
	router := NewRouter(store, registry, true)

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	paths := []string{"/api/v1/messages", "/api/v1/messages/123", "/api/v1/channels", "/api/v1/channels/123"}

	for _, path := range paths {
		for _, method := range methods {
			req := httptest.NewRequest(method, path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s %s expected status 405, got %d", method, path, rec.Code)
			}
		}
	}
}

func TestAPIResponses_NeverContainPayloadBytes(t *testing.T) {
	server, store, _ := setupTestServer(false)
	ctx := context.Background()

	payload := "PATIENT:DOE,JANE|MRN:87654321|SSN:111-11-1111|DOB:1990-02-02"
	env := &payloadref.Envelope{
		ChannelID:  "ch-test",
		MessageID:  "msg-phi",
		ReceivedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:     "received",
		Ref: payloadref.PayloadRef{
			StorageID: "store-phi",
			Location:  "mem://ghega/phi",
		},
	}
	if err := store.Save(ctx, env, []byte(payload)); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test ListMessages
	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	rec := httptest.NewRecorder()
	server.ListMessages(rec, req)
	body := rec.Body.String()
	if strings.Contains(body, payload) {
		t.Errorf("ListMessages response leaked payload bytes: %s", body)
	}
	if !strings.Contains(body, "msg-phi") {
		t.Errorf("ListMessages response missing message_id: %s", body)
	}

	// Test GetMessage
	req = httptest.NewRequest(http.MethodGet, "/api/v1/messages/msg-phi", nil)
	rec = httptest.NewRecorder()
	server.GetMessage(rec, req)
	body = rec.Body.String()
	if strings.Contains(body, payload) {
		t.Errorf("GetMessage response leaked payload bytes: %s", body)
	}
	if !strings.Contains(body, "msg-phi") {
		t.Errorf("GetMessage response missing message_id: %s", body)
	}
}

func TestListMessages_EmptyStore(t *testing.T) {
	server, _, _ := setupTestServer(false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	rec := httptest.NewRecorder()
	server.ListMessages(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp PaginatedMessagesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(resp.Messages))
	}
}

func TestListChannels_EmptyRegistry(t *testing.T) {
	server, _, _ := setupTestServer(false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels", nil)
	rec := httptest.NewRecorder()
	server.ListChannels(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var channels []*ChannelConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &channels); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(channels) != 0 {
		t.Errorf("expected 0 channels, got %d", len(channels))
	}
}

func TestServe_DevAuthFlag(t *testing.T) {
	// This test verifies the --dev-auth flag is wired correctly by checking
	// that the router created with devAuth=true allows unauthenticated requests.
	store := messagestore.NewInMemoryStore()
	registry := NewInMemoryChannelRegistry()
	router := NewRouter(store, registry, true)

	srv := httptest.NewServer(router)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/messages")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status 200 with dev-auth, got %d: %s", resp.StatusCode, string(body))
	}
}
