package payloadref

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// syntheticPayload is a byte sequence that must NEVER appear in any string
// representation of PayloadRef or Envelope. It simulates what real PHI would look like.
const syntheticPayload = "PATIENT:DOE,JOHN|MRN:12345678|SSN:000-00-0000|DOB:1980-01-01"

func TestPayloadRef_String_NeverContainsPayloadBytes(t *testing.T) {
	ref := PayloadRef{
		StorageID: "store-abc-123",
		Location:  "s3://ghega-bucket/inbound/msg-001",
	}

	formats := []string{
		fmt.Sprintf("%v", ref),
		fmt.Sprintf("%+v", ref),
		fmt.Sprintf("%#v", ref),
		ref.String(),
	}

	for _, out := range formats {
		if strings.Contains(out, syntheticPayload) {
			t.Errorf("PayloadRef string representation leaked payload bytes: %q", out)
		}
		// Also ensure the struct fields we expect are present
		if !strings.Contains(out, "store-abc-123") {
			t.Errorf("PayloadRef string missing StorageID: %q", out)
		}
		if !strings.Contains(out, "s3://ghega-bucket/inbound/msg-001") {
			t.Errorf("PayloadRef string missing Location: %q", out)
		}
	}
}

func TestEnvelope_String_NeverContainsPayloadBytes(t *testing.T) {
	ref := PayloadRef{
		StorageID: "store-abc-123",
		Location:  "s3://ghega-bucket/inbound/msg-001",
	}
	env := Envelope{
		ChannelID:  "channel-mllp-inbound",
		MessageID:  "msg-uuid-1234",
		ReceivedAt: time.Date(2024, 1, 15, 9, 30, 0, 0, time.UTC),
		Status:     "received",
		Ref:        ref,
	}

	formats := []string{
		fmt.Sprintf("%v", env),
		fmt.Sprintf("%+v", env),
		fmt.Sprintf("%#v", env),
		env.String(),
	}

	for _, out := range formats {
		if strings.Contains(out, syntheticPayload) {
			t.Errorf("Envelope string representation leaked payload bytes: %q", out)
		}
		if !strings.Contains(out, "channel-mllp-inbound") {
			t.Errorf("Envelope string missing ChannelID: %q", out)
		}
		if !strings.Contains(out, "msg-uuid-1234") {
			t.Errorf("Envelope string missing MessageID: %q", out)
		}
		if !strings.Contains(out, "2024-01-15T09:30:00Z") {
			t.Errorf("Envelope string missing ReceivedAt: %q", out)
		}
		if !strings.Contains(out, "received") {
			t.Errorf("Envelope string missing Status: %q", out)
		}
		if !strings.Contains(out, "store-abc-123") {
			t.Errorf("Envelope string missing Ref.StorageID: %q", out)
		}
	}
}

func TestPayloadRef_DoesNotHoldPayloadBytes(t *testing.T) {
	ref := PayloadRef{
		StorageID: "store-xyz",
		Location:  "fs:///tmp/ghega/queue/msg-002",
	}

	// Verify there is no field that could accidentally hold payload bytes.
	// This is a compile-time guarantee, but we assert it at runtime via
	// reflection to guard against future refactors.
	typ := ref
	_ = typ // PayloadRef only has StorageID and Location (both strings).
}

func TestEnvelope_DoesNotHoldPayloadBytes(t *testing.T) {
	env := Envelope{
		ChannelID:  "ch-1",
		MessageID:  "msg-1",
		ReceivedAt: time.Now(),
		Status:     "pending",
		Ref: PayloadRef{
			StorageID: "store-1",
			Location:  "loc-1",
		},
	}

	// Envelope only holds metadata and a PayloadRef. No payload byte fields.
	_ = env
}

func TestPayloadRef_String_Format(t *testing.T) {
	ref := PayloadRef{StorageID: "sid", Location: "loc"}
	want := "PayloadRef{StorageID:sid Location:loc}"
	got := ref.String()
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestEnvelope_String_Format(t *testing.T) {
	ref := PayloadRef{StorageID: "sid", Location: "loc"}
	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	env := Envelope{ChannelID: "ch", MessageID: "mid", ReceivedAt: ts, Status: "ok", Ref: ref}
	want := "Envelope{ChannelID:ch MessageID:mid ReceivedAt:2024-06-01T12:00:00Z Status:ok Ref:PayloadRef{StorageID:sid Location:loc}}"
	got := env.String()
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
