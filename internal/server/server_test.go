package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sroopra/ghega/internal/alerts"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/payloadref"
)

func newTestAlertStore() alerts.AlertStore {
	return alerts.NewInMemoryAlertStore()
}

func mustSaveMessage(t *testing.T, store messagestore.Store, env *payloadref.Envelope, payload []byte) {
	t.Helper()
	ctx := context.Background()
	if err := store.Save(ctx, env, payload); err != nil {
		t.Fatalf("save message: %v", err)
	}
}

func TestHealthz(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("healthz request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `"status":"ok"`) {
		t.Errorf("body missing ok status: %s", body)
	}
}

func TestChannels(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/channels")
	if err != nil {
		t.Fatalf("channels request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var channels []channelResponse
	if err := json.NewDecoder(resp.Body).Decode(&channels); err != nil {
		t.Fatalf("decode channels: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("len(channels) = %d, want 1", len(channels))
	}
	if channels[0].ID != "adt-a01" {
		t.Errorf("channel id = %q, want %q", channels[0].ID, "adt-a01")
	}
	if channels[0].Name != "ADT A01 MLLP to HTTP" {
		t.Errorf("channel name = %q, want %q", channels[0].Name, "ADT A01 MLLP to HTTP")
	}
}

func TestListMessages(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		env := &payloadref.Envelope{
			ChannelID:  "ch-oru",
			MessageID:  "msg-00" + string(rune('1'+i)),
			ReceivedAt: time.Date(2024, 1, 1, 0, 0, i, 0, time.UTC),
			Status:     "received",
			Ref:        payloadref.PayloadRef{StorageID: "sid-00" + string(rune('1'+i)), Location: "mem://test"},
		}
		if err := store.Save(ctx, env, []byte("payload")); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	resp, err := http.Get(ts.URL + "/api/v1/messages?limit=2&offset=0")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var msgs []messageMetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgs); err != nil {
		t.Fatalf("decode messages: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("len(msgs) = %d, want 2", len(msgs))
	}
}

func TestListMessagesByChannel(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	mustSaveMessage(t, store, &payloadref.Envelope{
		ChannelID:  "ch-adt",
		MessageID:  "msg-adt",
		ReceivedAt: time.Now(),
		Status:     "received",
		Ref:        payloadref.PayloadRef{StorageID: "sid-adt", Location: "mem://test"},
	}, []byte("adt"))
	mustSaveMessage(t, store, &payloadref.Envelope{
		ChannelID:  "ch-oru",
		MessageID:  "msg-oru",
		ReceivedAt: time.Now(),
		Status:     "received",
		Ref:        payloadref.PayloadRef{StorageID: "sid-oru", Location: "mem://test"},
	}, []byte("oru"))

	resp, err := http.Get(ts.URL + "/api/v1/messages?channel_id=ch-adt")
	if err != nil {
		t.Fatalf("list messages by channel: %v", err)
	}
	defer resp.Body.Close()

	var msgs []messageMetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("len(msgs) = %d, want 1", len(msgs))
	}
	if msgs[0].MessageID != "msg-adt" {
		t.Errorf("message_id = %q, want %q", msgs[0].MessageID, "msg-adt")
	}
}

func TestGetMessage(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	mustSaveMessage(t, store, &payloadref.Envelope{
		ChannelID:  "ch-adt",
		MessageID:  "msg-123",
		ReceivedAt: time.Now(),
		Status:     "delivered",
		Ref:        payloadref.PayloadRef{StorageID: "sid-123", Location: "mem://test"},
	}, []byte("payload"))

	resp, err := http.Get(ts.URL + "/api/v1/messages/msg-123")
	if err != nil {
		t.Fatalf("get message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var msg messageMetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if msg.MessageID != "msg-123" {
		t.Errorf("message_id = %q, want %q", msg.MessageID, "msg-123")
	}
	if msg.Status != "delivered" {
		t.Errorf("status = %q, want %q", msg.Status, "delivered")
	}
}

func TestGetMessage_NotFound(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/messages/nonexistent")
	if err != nil {
		t.Fatalf("get message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAuthMiddleware_RejectsInvalidBearer(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/api/v1/channels", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer invalid")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_AllowsValidRequest(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/channels")
	if err != nil {
		t.Fatalf("channels request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRedeliver_Returns501(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/v1/messages/msg-001/redeliver", "application/json", nil)
	if err != nil {
		t.Fatalf("redeliver request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotImplemented)
	}
}

func TestReplay_Returns501(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/v1/messages/msg-001/replay", "application/json", nil)
	if err != nil {
		t.Fatalf("replay request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotImplemented)
	}
}

func TestCORSHeaders(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	req, err := http.NewRequest("OPTIONS", ts.URL+"/api/v1/channels", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "*")
	}
	if got := resp.Header.Get("Access-Control-Allow-Methods"); !strings.Contains(got, "GET") {
		t.Errorf("Access-Control-Allow-Methods missing GET: %q", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Authorization") {
		t.Errorf("Access-Control-Allow-Headers missing Authorization: %q", got)
	}
}

func TestListAlerts(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	alertStore := newTestAlertStore()
	srv := New(store, alertStore)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	_ = alertStore.Create(&alerts.Alert{
		ID:        "alert-1",
		ChannelID: "ch-1",
		Severity:  alerts.SeverityWarning,
		Message:   "test alert",
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	resp, err := http.Get(ts.URL + "/api/v1/alerts")
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result []alerts.Alert
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode alerts: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len(alerts) = %d, want 1", len(result))
	}
	if result[0].ID != "alert-1" {
		t.Errorf("alert id = %q, want %q", result[0].ID, "alert-1")
	}
	if result[0].Severity != alerts.SeverityWarning {
		t.Errorf("severity = %q, want %q", result[0].Severity, alerts.SeverityWarning)
	}
}

func TestRootServesHTML(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("root request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", contentType)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "<!DOCTYPE html>") && !strings.Contains(string(body), "<html") {
		t.Errorf("body does not contain HTML: %s", body)
	}
}

func TestAPIRoutesTakePrecedenceOverStaticFiles(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/channels")
	if err != nil {
		t.Fatalf("channels request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "adt-a01") {
		t.Errorf("body missing expected channel id: %s", body)
	}
}
