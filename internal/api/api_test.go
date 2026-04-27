package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRedeliverEndpoint_Returns501(t *testing.T) {
	h := Handler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/msg-001/redeliver", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "not yet implemented") {
		t.Errorf("expected body to contain 'not yet implemented', got: %s", body)
	}
}

func TestReplayEndpoint_Returns501(t *testing.T) {
	h := Handler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/msg-001/replay", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "not yet implemented") {
		t.Errorf("expected body to contain 'not yet implemented', got: %s", body)
	}
}

func TestMessagesEndpoint_GetMethodNotAllowed(t *testing.T) {
	h := Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/msg-001/redeliver", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}
