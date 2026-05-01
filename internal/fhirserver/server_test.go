package fhirserver

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/sroopra/ghega/pkg/fhir"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/payloadref"
)

func newTestServer(t *testing.T) (*httptest.Server, *messagestore.InMemoryStore) {
	t.Helper()
	store := messagestore.NewInMemoryStore()
	handler := NewHandler(Config{
		BasePath: "/fhir",
		Store:    store,
		Logger:   slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})),
	})
	return httptest.NewServer(handler), store
}

func TestPostPatient(t *testing.T) {
	ts, store := newTestServer(t)
	defer ts.Close()

	patient := `{
		"resourceType": "Patient",
		"id": "test-patient-001",
		"name": [{"family": "TestPatient", "given": ["John"]}],
		"gender": "male"
	}`

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/fhir/Patient/test-patient-001", bytes.NewReader([]byte(patient)))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/fhir+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc == "" {
		t.Fatal("expected Location header")
	}

	// Verify metadata stored.
	ctx := t.Context()
	env, err := store.GetMetadata(ctx, "Patient/test-patient-001")
	if err != nil {
		t.Fatalf("get metadata: %v", err)
	}
	if env.ChannelID != "fhir" {
		t.Fatalf("expected channel id fhir, got %s", env.ChannelID)
	}

	// Verify payload stored.
	storedPayload, ok := store.GetPayload(env.Ref.StorageID)
	if !ok {
		t.Fatal("payload not found")
	}
	var stored map[string]any
	if err := json.Unmarshal(storedPayload, &stored); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if stored["resourceType"] != "Patient" {
		t.Fatalf("expected resourceType Patient, got %v", stored["resourceType"])
	}
}

func TestPostInvalidJSON(t *testing.T) {
	ts, _ := newTestServer(t)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/fhir", bytes.NewReader([]byte("not json")))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/fhir+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var oo fhir.OperationOutcome
	if err := json.Unmarshal(body, &oo); err != nil {
		t.Fatalf("expected OperationOutcome JSON: %v", err)
	}
	if len(oo.Issue) == 0 {
		t.Fatal("expected at least one OperationOutcomeIssue")
	}
	if oo.Issue[0].Severity != "error" {
		t.Fatalf("expected severity error, got %s", oo.Issue[0].Severity)
	}
}

func TestPostBundle(t *testing.T) {
	ts, store := newTestServer(t)
	defer ts.Close()

	bundle := `{
		"resourceType": "Bundle",
		"type": "transaction",
		"entry": [
			{
				"resource": {
					"resourceType": "Patient",
					"id": "bundle-patient-001",
					"gender": "female"
				}
			},
			{
				"resource": {
					"resourceType": "Observation",
					"id": "bundle-obs-001",
					"status": "final"
				}
			}
		]
	}`

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/fhir", bytes.NewReader([]byte(bundle)))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/fhir+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	ctx := t.Context()

	// Verify both resources stored.
	env1, err := store.GetMetadata(ctx, "Patient/bundle-patient-001")
	if err != nil {
		t.Fatalf("get patient metadata: %v", err)
	}
	payload1, ok := store.GetPayload(env1.Ref.StorageID)
	if !ok {
		t.Fatal("patient payload not found")
	}
	var p1 map[string]any
	if err := json.Unmarshal(payload1, &p1); err != nil {
		t.Fatalf("unmarshal patient: %v", err)
	}
	if p1["resourceType"] != "Patient" {
		t.Fatalf("expected Patient, got %v", p1["resourceType"])
	}

	env2, err := store.GetMetadata(ctx, "Observation/bundle-obs-001")
	if err != nil {
		t.Fatalf("get observation metadata: %v", err)
	}
	payload2, ok := store.GetPayload(env2.Ref.StorageID)
	if !ok {
		t.Fatal("observation payload not found")
	}
	var p2 map[string]any
	if err := json.Unmarshal(payload2, &p2); err != nil {
		t.Fatalf("unmarshal observation: %v", err)
	}
	if p2["resourceType"] != "Observation" {
		t.Fatalf("expected Observation, got %v", p2["resourceType"])
	}
}

func TestGetStoredResource(t *testing.T) {
	ts, store := newTestServer(t)
	defer ts.Close()

	// Seed store directly.
	ctx := t.Context()
	patient := []byte(`{"resourceType":"Patient","id":"get-patient-001","gender":"unknown"}`)
	env := &payloadref.Envelope{
		ChannelID:  "fhir",
		MessageID:  "Patient/get-patient-001",
		ReceivedAt: time.Now(),
		Status:     "received",
		Ref: payloadref.PayloadRef{
			StorageID: "Patient/get-patient-001",
			Location:  "fhirserver",
		},
	}
	if err := store.Save(ctx, env, patient); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/fhir/Patient/get-patient-001", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/fhir+json" {
		t.Fatalf("expected Content-Type application/fhir+json, got %s", ct)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != string(patient) {
		t.Fatalf("expected body %s, got %s", patient, body)
	}
}

func TestWrongContentType(t *testing.T) {
	ts, _ := newTestServer(t)
	defer ts.Close()

	patient := `{"resourceType":"Patient","id":"ct-patient-001"}`
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/fhir", bytes.NewReader([]byte(patient)))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var oo fhir.OperationOutcome
	if err := json.Unmarshal(body, &oo); err != nil {
		t.Fatalf("expected OperationOutcome JSON: %v", err)
	}
	if len(oo.Issue) == 0 {
		t.Fatal("expected at least one issue")
	}
}

func TestPutResource(t *testing.T) {
	ts, store := newTestServer(t)
	defer ts.Close()

	patient := `{"resourceType":"Patient","id":"put-patient-001","gender":"male"}`
	req, err := http.NewRequest(http.MethodPut, ts.URL+"/fhir/Patient/put-patient-001", bytes.NewReader([]byte(patient)))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/fhir+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	ctx := t.Context()
	env, err := store.GetMetadata(ctx, "Patient/put-patient-001")
	if err != nil {
		t.Fatalf("get metadata: %v", err)
	}
	payload, ok := store.GetPayload(env.Ref.StorageID)
	if !ok {
		t.Fatal("payload not found")
	}
	if string(payload) != patient {
		t.Fatalf("expected payload %s, got %s", patient, payload)
	}
}

func TestDeleteResource(t *testing.T) {
	ts, store := newTestServer(t)
	defer ts.Close()

	// Seed store directly.
	ctx := t.Context()
	patient := []byte(`{"resourceType":"Patient","id":"del-patient-001","gender":"female"}`)
	env := &payloadref.Envelope{
		ChannelID:  "fhir",
		MessageID:  "Patient/del-patient-001",
		ReceivedAt: time.Now(),
		Status:     "received",
		Ref: payloadref.PayloadRef{
			StorageID: "Patient/del-patient-001",
			Location:  "fhirserver",
		},
	}
	if err := store.Save(ctx, env, patient); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	req, err := http.NewRequest(http.MethodDelete, ts.URL+"/fhir/Patient/del-patient-001", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}

	_, err = store.GetMetadata(ctx, "Patient/del-patient-001")
	if err == nil {
		t.Fatal("expected resource to be deleted")
	}
}
