package fhirsender

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// syntheticPayload simulates payload bytes that must never appear in log output.
const syntheticPayload = `{"resourceType":"Patient","id":" synthetic-patient-001","name":[{"family":"Testpatient"}]}`

func captureLogOutput(f func()) string {
	var buf bytes.Buffer
	oldLogger := slog.Default()
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(oldLogger)
	f()
	return buf.String()
}

func TestSend_Success(t *testing.T) {
	var receivedMethod string
	var receivedBody []byte
	var receivedContentType string
	var receivedAccept string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedContentType = r.Header.Get("Content-Type")
		receivedAccept = r.Header.Get("Accept")
		body, _ := io.ReadAll(r.Body)
		receivedBody = body
		w.Header().Set("Content-Type", "application/fhir+json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"resourceType":"OperationOutcome"}`))
	}))
	defer server.Close()

	sender := &Sender{
		URL:     server.URL,
		Method:  http.MethodPut,
		Headers: map[string]string{"X-Ghega-Test": "contract"},
		Timeout: 5 * time.Second,
		Retries: 1,
	}

	payload := []byte(syntheticPayload)
	resp, err := sender.Send(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodPut {
		t.Errorf("expected method %q, got %q", http.MethodPut, receivedMethod)
	}
	if receivedContentType != "application/fhir+json; fhirVersion=4.0" {
		t.Errorf("expected Content-Type %q, got %q", "application/fhir+json; fhirVersion=4.0", receivedContentType)
	}
	if receivedAccept != "application/fhir+json" {
		t.Errorf("expected Accept %q, got %q", "application/fhir+json", receivedAccept)
	}
	if string(receivedBody) != string(payload) {
		t.Errorf("expected body %q, got %q", string(payload), string(receivedBody))
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if string(resp.Body) != `{"resourceType":"OperationOutcome"}` {
		t.Errorf("expected response body %q, got %q", `{"resourceType":"OperationOutcome"}`, string(resp.Body))
	}
	if resp.Headers.Get("Content-Type") != "application/fhir+json" {
		t.Errorf("expected Content-Type header application/fhir+json, got %q", resp.Headers.Get("Content-Type"))
	}
}

func TestSend_DefaultMethodIsPost(t *testing.T) {
	var receivedMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &Sender{
		URL:     server.URL,
		Timeout: 5 * time.Second,
		Retries: 1,
	}

	_, err := sender.Send(context.Background(), []byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedMethod != http.MethodPost {
		t.Errorf("expected default method %q, got %q", http.MethodPost, receivedMethod)
	}
}

func TestSend_RetriesOnServerError(t *testing.T) {
	var attemptCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attemptCount.Add(1)
		if count < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := &Sender{
		URL:     server.URL,
		Timeout: 5 * time.Second,
		Retries: 5,
	}

	_, err := sender.Send(context.Background(), []byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attemptCount.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attemptCount.Load())
	}
}

func TestSend_FailsAfterMaxRetries(t *testing.T) {
	var attemptCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := &Sender{
		URL:     server.URL,
		Timeout: 5 * time.Second,
		Retries: 2,
	}

	_, err := sender.Send(context.Background(), []byte(`{}`))
	if err == nil {
		t.Fatal("expected error after max retries, got nil")
	}
	if attemptCount.Load() != 2 {
		t.Errorf("expected 2 attempts, got %d", attemptCount.Load())
	}
	if !strings.Contains(err.Error(), "failed after 2 attempts") {
		t.Errorf("expected error to contain 'failed after 2 attempts', got: %v", err)
	}
}

func TestSend_NoRetryOnClientError(t *testing.T) {
	var attemptCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	sender := &Sender{
		URL:     server.URL,
		Timeout: 5 * time.Second,
		Retries: 3,
	}

	_, err := sender.Send(context.Background(), []byte(`{}`))
	if err == nil {
		t.Fatal("expected error for 4xx response, got nil")
	}
	if attemptCount.Load() != 1 {
		t.Errorf("expected 1 attempt for 4xx, got %d", attemptCount.Load())
	}
	if !strings.Contains(err.Error(), "server returned status 400") {
		t.Errorf("expected error to contain 'server returned status 400', got: %v", err)
	}
}

func TestSend_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	sender := &Sender{
		URL:     server.URL,
		Timeout: 5 * time.Second,
		Retries: 1,
	}

	_, err := sender.Send(ctx, []byte(`{}`))
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}

func TestSend_MissingURL(t *testing.T) {
	sender := &Sender{Retries: 1}
	_, err := sender.Send(context.Background(), []byte(`{}`))
	if err == nil {
		t.Fatal("expected error for missing URL, got nil")
	}
	if !strings.Contains(err.Error(), "URL is required") {
		t.Errorf("expected error to contain 'URL is required', got: %v", err)
	}
}

func TestSend_NeverLogsPayloadBytes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	sender := &Sender{
		URL:     server.URL,
		Timeout: 5 * time.Second,
		Retries: 1,
		Logger:  logger,
	}

	payload := []byte(syntheticPayload)
	_, err := sender.Send(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logs := buf.String()
	if strings.Contains(logs, syntheticPayload) {
		t.Errorf("logs contained payload bytes: %q", logs)
	}
	if !strings.Contains(logs, "sending FHIR request") {
		t.Errorf("logs missing 'sending FHIR request': %q", logs)
	}
	if !strings.Contains(logs, "FHIR response received") {
		t.Errorf("logs missing 'FHIR response received': %q", logs)
	}
}

func TestSend_LogsMetadataOnError(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	// Use an invalid URL to force a network error
	sender := &Sender{
		URL:     "http://127.0.0.1:1",
		Timeout: 100 * time.Millisecond,
		Retries: 1,
		Logger:  logger,
	}

	_, err := sender.Send(context.Background(), []byte(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}

	logs := buf.String()
	if strings.Contains(logs, syntheticPayload) {
		t.Errorf("logs contained payload bytes: %q", logs)
	}
	if !strings.Contains(logs, "FHIR request failed") {
		t.Errorf("logs missing 'FHIR request failed': %q", logs)
	}
	if !strings.Contains(logs, "127.0.0.1:1") {
		t.Errorf("logs missing URL: %q", logs)
	}
}

func TestDryRun_ValidPayload(t *testing.T) {
	sender := &Sender{URL: "http://example.com/fhir"}
	err := sender.DryRun([]byte(`{"resourceType":"Patient"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDryRun_NilPayload(t *testing.T) {
	sender := &Sender{URL: "http://example.com/fhir"}
	err := sender.DryRun(nil)
	if err == nil {
		t.Fatal("expected error for nil payload, got nil")
	}
	if !strings.Contains(err.Error(), "payload is nil") {
		t.Errorf("expected error to contain 'payload is nil', got: %v", err)
	}
}

func TestDryRun_MissingURL(t *testing.T) {
	sender := &Sender{}
	err := sender.DryRun([]byte(`{}`))
	if err == nil {
		t.Fatal("expected error for missing URL, got nil")
	}
	if !strings.Contains(err.Error(), "URL is required") {
		t.Errorf("expected error to contain 'URL is required', got: %v", err)
	}
}

func TestSender_Defaults(t *testing.T) {
	sender := &Sender{URL: "http://example.com/fhir"}
	if sender.method() != http.MethodPost {
		t.Errorf("expected default method POST, got %q", sender.method())
	}
	if sender.timeout() != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", sender.timeout())
	}
	if sender.retries() != 3 {
		t.Errorf("expected default retries 3, got %d", sender.retries())
	}
	if sender.logger() != slog.Default() {
		t.Error("expected default logger to be slog.Default()")
	}
}

func TestSender_CustomValues(t *testing.T) {
	sender := &Sender{
		Method:  http.MethodPatch,
		Timeout: 10 * time.Second,
		Retries: 5,
	}
	if sender.method() != http.MethodPatch {
		t.Errorf("expected method PATCH, got %q", sender.method())
	}
	if sender.timeout() != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", sender.timeout())
	}
	if sender.retries() != 5 {
		t.Errorf("expected retries 5, got %d", sender.retries())
	}
}

func TestSend_BundlePayload(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = body
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"resourceType":"Bundle"}`))
	}))
	defer server.Close()

	sender := &Sender{
		URL:     server.URL,
		Timeout: 5 * time.Second,
		Retries: 1,
	}

	bundle := []byte(`{"resourceType":"Bundle","type":"collection","entry":[{"resource":{"resourceType":"Patient"}}]}`)
	resp, err := sender.Send(context.Background(), bundle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
	if string(receivedBody) != string(bundle) {
		t.Errorf("expected body %q, got %q", string(bundle), string(receivedBody))
	}
}
