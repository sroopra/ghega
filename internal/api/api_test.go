package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/payloadref"
)

func newTestEnvelope(channelID, messageID, status string) *payloadref.Envelope {
	return &payloadref.Envelope{
		ChannelID:  channelID,
		MessageID:  messageID,
		ReceivedAt: time.Date(2024, 3, 10, 8, 15, 0, 0, time.UTC),
		Status:     status,
		Ref: payloadref.PayloadRef{
			StorageID: "store-" + messageID,
			Location:  "mem://ghega/" + messageID,
		},
	}
}

func setupTestHandler() (*Handler, *messagestore.InMemoryStore, *InMemoryChannelStore) {
	msgStore := messagestore.NewInMemoryStore()
	chStore := NewInMemoryChannelStore()
	chStore.Register(Channel{
		ID:        "ch-adt",
		Name:      "ADT Channel",
		Type:      "mllp-to-http",
		Status:    "active",
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	chStore.Register(Channel{
		ID:        "ch-oru",
		Name:      "ORU Channel",
		Type:      "mllp-to-http",
		Status:    "active",
		CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	})
	return NewHandler(msgStore, chStore), msgStore, chStore
}

func TestHealthz(t *testing.T) {
	handler, _, _ := setupTestHandler()
	router := NewRouter(handler, false)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"status":"ok"`) {
		t.Errorf("expected body to contain status ok, got: %s", body)
	}
}

func TestListMessages(t *testing.T) {
	handler, msgStore, _ := setupTestHandler()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		env := newTestEnvelope("ch-oru", "msg-oru-00"+strconv.Itoa(i), "received")
		if err := msgStore.Save(ctx, env, []byte("payload-"+strconv.Itoa(i))); err != nil {
			t.Fatalf("save failed: %v", err)
		}
	}
	for i := 0; i < 3; i++ {
		env := newTestEnvelope("ch-adt", "msg-adt-00"+strconv.Itoa(i), "processed")
		if err := msgStore.Save(ctx, env, []byte("adt-payload-"+strconv.Itoa(i))); err != nil {
			t.Fatalf("save failed: %v", err)
		}
	}

	router := NewRouter(handler, true) // dev auth bypass

	t.Run("list all messages", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/messages?limit=10&offset=0", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		var resp []MessageMetadataResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if len(resp) != 8 {
			t.Errorf("expected 8 messages, got %d", len(resp))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/messages?limit=3&offset=0", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		var resp []MessageMetadataResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if len(resp) != 3 {
			t.Errorf("expected 3 messages, got %d", len(resp))
		}
	})

	t.Run("filter by channel", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/messages?channel_id=ch-adt", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		var resp []MessageMetadataResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if len(resp) != 3 {
			t.Errorf("expected 3 messages, got %d", len(resp))
		}
		for _, m := range resp {
			if m.ChannelID != "ch-adt" {
				t.Errorf("expected channel ch-adt, got %s", m.ChannelID)
			}
		}
	})
}

func TestGetMessage(t *testing.T) {
	handler, msgStore, _ := setupTestHandler()
	ctx := context.Background()

	env := newTestEnvelope("ch-oru", "msg-001", "received")
	if err := msgStore.Save(ctx, env, []byte("synthetic-payload-data")); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	router := NewRouter(handler, true)

	t.Run("existing message", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/msg-001", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		var resp MessageMetadataResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if resp.MessageID != "msg-001" {
			t.Errorf("expected msg-001, got %s", resp.MessageID)
		}
		if resp.ChannelID != "ch-oru" {
			t.Errorf("expected ch-oru, got %s", resp.ChannelID)
		}
		if resp.Status != "received" {
			t.Errorf("expected received, got %s", resp.Status)
		}
	})

	t.Run("missing message", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/nonexistent", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rr.Code)
		}
	})
}

func TestListChannels(t *testing.T) {
	handler, _, _ := setupTestHandler()
	router := NewRouter(handler, true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp []Channel
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 channels, got %d", len(resp))
	}
}

func TestGetChannel(t *testing.T) {
	handler, _, _ := setupTestHandler()
	router := NewRouter(handler, true)

	t.Run("existing channel", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/ch-adt", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		var resp Channel
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if resp.ID != "ch-adt" {
			t.Errorf("expected ch-adt, got %s", resp.ID)
		}
	})

	t.Run("missing channel", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/nonexistent", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rr.Code)
		}
	})
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	handler, _, _ := setupTestHandler()
	router := NewRouter(handler, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	handler, _, _ := setupTestHandler()
	router := NewRouter(handler, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	handler, _, _ := setupTestHandler()
	router := NewRouter(handler, false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	handler, _, _ := setupTestHandler()
	router := NewRouter(handler, false)

	// A JWT-like token has three base64url parts separated by dots
	validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDevAuthBypass(t *testing.T) {
	handler, _, _ := setupTestHandler()
	router := NewRouter(handler, true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	// No Authorization header
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 with dev-auth, got %d", rr.Code)
	}
}

func TestNoPayloadBytesInResponses(t *testing.T) {
	handler, msgStore, _ := setupTestHandler()
	ctx := context.Background()

	// Use a payload that would be easy to spot if leaked
	payload := "PATIENT:DOE,JANE|MRN:87654321|SSN:111-11-1111|DOB:1990-02-02"
	env := newTestEnvelope("ch-oru", "msg-leak-test", "received")
	if err := msgStore.Save(ctx, env, []byte(payload)); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	router := NewRouter(handler, true)

	t.Run("list messages", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		body := rr.Body.String()
		if strings.Contains(body, payload) {
			t.Errorf("list messages leaked payload bytes: %s", body)
		}
	})

	t.Run("get message", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/msg-leak-test", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		body := rr.Body.String()
		if strings.Contains(body, payload) {
			t.Errorf("get message leaked payload bytes: %s", body)
		}
	})
}

func TestAuthMiddleware_AttachesUser(t *testing.T) {
	var capturedUser *User
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = ContextUser(r)
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(false)
	wrapped := middleware(handler)

	validToken := "header.payload.signature"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if capturedUser == nil {
		t.Fatal("expected user to be attached to context")
	}
	if capturedUser.Role != "operator" {
		t.Errorf("expected role operator, got %s", capturedUser.Role)
	}
	if capturedUser.Token != validToken {
		t.Errorf("expected token to match")
	}
}

func TestAuthMiddleware_DevAuthUser(t *testing.T) {
	var capturedUser *User
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = ContextUser(r)
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(true)
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if capturedUser == nil {
		t.Fatal("expected user to be attached to context in dev mode")
	}
	if capturedUser.ID != "dev-user" {
		t.Errorf("expected dev-user, got %s", capturedUser.ID)
	}
	if capturedUser.Role != "admin" {
		t.Errorf("expected admin role, got %s", capturedUser.Role)
	}
}
