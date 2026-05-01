package fhirserver

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sroopra/ghega/pkg/fhir"
	"github.com/sroopra/ghega/pkg/messagestore"
)

func TestPostPatient(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, WithBasePath("/fhir/R4"))

	patient := map[string]any{
		"resourceType": "Patient",
		"name": []map[string]any{
			{"family": "TESTPATIENT", "given": []string{"ONE"}},
		},
	}
	body, _ := json.Marshal(patient)

	req := httptest.NewRequest(http.MethodPost, "/fhir/R4/Patient", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	if !strings.HasPrefix(location, "/fhir/R4/Patient/") {
		t.Fatalf("unexpected Location header: %s", location)
	}

	// Verify metadata stored in messagestore.
	ctx := context.Background()
	msgs, err := store.ListAll(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message in store, got %d", len(msgs))
	}
	if msgs[0].Status != "received" {
		t.Fatalf("expected status received, got %s", msgs[0].Status)
	}
	if msgs[0].ChannelID != "fhir" {
		t.Fatalf("expected channel_id fhir, got %s", msgs[0].ChannelID)
	}

	// Verify payload stored.
	payload, ok := store.GetPayload(msgs[0].Ref.StorageID)
	if !ok {
		t.Fatal("expected payload to be stored")
	}
	if !bytes.Contains(payload, []byte("TESTPATIENT")) {
		t.Fatalf("stored payload missing expected data")
	}
}

func TestPostInvalidJSON(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, WithBasePath("/fhir/R4"))

	req := httptest.NewRequest(http.MethodPost, "/fhir/R4/Patient", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}

	var outcome fhir.OperationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse OperationOutcome: %v", err)
	}
	if outcome.ResourceType != "OperationOutcome" {
		t.Fatalf("expected ResourceType OperationOutcome, got %s", outcome.ResourceType)
	}
	if len(outcome.Issue) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(outcome.Issue))
	}
	if outcome.Issue[0].Severity != "error" {
		t.Fatalf("expected severity error, got %s", outcome.Issue[0].Severity)
	}
}

func TestPostBundle(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, WithBasePath("/fhir/R4"))

	bundle := map[string]any{
		"resourceType": "Bundle",
		"type":         "collection",
		"entry": []map[string]any{
			{
				"resource": map[string]any{
					"resourceType": "Patient",
					"id":           "patient-1",
					"name": []map[string]any{
						{"family": "TESTPATIENT", "given": []string{"ONE"}},
					},
				},
			},
			{
				"resource": map[string]any{
					"resourceType": "Observation",
					"id":           "obs-1",
					"status":       "final",
				},
			},
		},
	}
	body, _ := json.Marshal(bundle)

	req := httptest.NewRequest(http.MethodPost, "/fhir/R4", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Fatalf("expected 200/201, got %d", rec.Code)
	}

	// Verify both entries stored.
	ctx := context.Background()
	msgs, err := store.ListAll(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages in store, got %d", len(msgs))
	}
}

func TestGetResource(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, WithBasePath("/fhir/R4"))

	patient := map[string]any{
		"resourceType": "Patient",
		"id":           "test-patient-1",
		"name": []map[string]any{
			{"family": "TESTPATIENT", "given": []string{"ONE"}},
		},
	}
	body, _ := json.Marshal(patient)

	req := httptest.NewRequest(http.MethodPost, "/fhir/R4/Patient", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	// GET the resource back.
	req = httptest.NewRequest(http.MethodGet, "/fhir/R4/Patient/test-patient-1", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/fhir+json") {
		t.Fatalf("unexpected Content-Type: %s", ct)
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got["resourceType"] != "Patient" {
		t.Fatalf("expected resourceType Patient, got %v", got["resourceType"])
	}
	if got["id"] != "test-patient-1" {
		t.Fatalf("expected id test-patient-1, got %v", got["id"])
	}
}

func TestWrongContentType(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, WithBasePath("/fhir/R4"))

	req := httptest.NewRequest(http.MethodPost, "/fhir/R4/Patient", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}

	var outcome fhir.OperationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse OperationOutcome: %v", err)
	}
	if outcome.ResourceType != "OperationOutcome" {
		t.Fatalf("expected ResourceType OperationOutcome, got %s", outcome.ResourceType)
	}
}

func TestDeleteResource(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, WithBasePath("/fhir/R4"))

	patient := map[string]any{
		"resourceType": "Patient",
		"id":           "del-patient-1",
	}
	body, _ := json.Marshal(patient)

	req := httptest.NewRequest(http.MethodPost, "/fhir/R4/Patient", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	// DELETE the resource.
	req = httptest.NewRequest(http.MethodDelete, "/fhir/R4/Patient/del-patient-1", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %d", rec.Code)
	}

	// GET should now return 404.
	req = httptest.NewRequest(http.MethodGet, "/fhir/R4/Patient/del-patient-1", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec.Code)
	}
}

func TestPutResource(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, WithBasePath("/fhir/R4"))

	patient := map[string]any{
		"resourceType": "Patient",
		"id":           "put-patient-1",
		"name": []map[string]any{
			{"family": "OLDNAME"},
		},
	}
	body, _ := json.Marshal(patient)

	req := httptest.NewRequest(http.MethodPut, "/fhir/R4/Patient/put-patient-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for PUT, got %d", rec.Code)
	}

	// Verify stored.
	ctx := context.Background()
	msgs, _ := store.ListAll(ctx, 10, 0)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message in store, got %d", len(msgs))
	}

	// Verify GET returns it.
	req = httptest.NewRequest(http.MethodGet, "/fhir/R4/Patient/put-patient-1", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on GET, got %d", rec.Code)
	}
}

func TestMissingResourceType(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, WithBasePath("/fhir/R4"))

	req := httptest.NewRequest(http.MethodPost, "/fhir/R4/Patient", strings.NewReader(`{"name": []}`))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var outcome fhir.OperationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &outcome); err != nil {
		t.Fatalf("failed to parse OperationOutcome: %v", err)
	}
	if outcome.Issue[0].Diagnostics != "Missing resourceType" {
		t.Fatalf("unexpected diagnostics: %s", outcome.Issue[0].Diagnostics)
	}
}

func TestGetNotFound(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, WithBasePath("/fhir/R4"))

	req := httptest.NewRequest(http.MethodGet, "/fhir/R4/Patient/does-not-exist", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, WithBasePath("/fhir/R4"))

	req := httptest.NewRequest(http.MethodPatch, "/fhir/R4/Patient/1", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestPHISafeLogging(t *testing.T) {
	// This test verifies that the server does not log payload bytes.
	// We use a custom logger and check its output.
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	store := messagestore.NewInMemoryStore()
	srv := New(store, WithBasePath("/fhir/R4"), WithLogger(logger))

	patient := map[string]any{
		"resourceType": "Patient",
		"name": []map[string]any{
			{"family": "TESTPATIENT", "given": []string{"ONE"}},
		},
	}
	body, _ := json.Marshal(patient)

	req := httptest.NewRequest(http.MethodPost, "/fhir/R4/Patient", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/fhir+json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	logOutput := buf.String()
	if strings.Contains(logOutput, "TESTPATIENT") {
		t.Fatalf("log output leaked payload bytes: %s", logOutput)
	}
	if !strings.Contains(logOutput, "FHIR resource stored") {
		t.Fatalf("log output missing expected event: %s", logOutput)
	}
	if !strings.Contains(logOutput, "resource_type") {
		t.Fatalf("log output missing resource_type: %s", logOutput)
	}
}
