package cli

import (
	"strings"
	"testing"
)

func TestMessageRedeliver_ReturnsError(t *testing.T) {
	err := runMessageRedeliver([]string{"msg-001", "--destination", "http://example.com"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "redeliver not yet implemented") {
		t.Errorf("expected error to contain 'redeliver not yet implemented', got: %s", err.Error())
	}
}

func TestMessageReplay_ReturnsError(t *testing.T) {
	err := runMessageReplay([]string{"msg-001", "--as-new"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "replay not yet implemented") {
		t.Errorf("expected error to contain 'replay not yet implemented', got: %s", err.Error())
	}
}

func TestMessageReplayPreview_ReturnsError(t *testing.T) {
	err := runMessageReplayPreview([]string{"msg-001"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "replay-preview not yet implemented") {
		t.Errorf("expected error to contain 'replay-preview not yet implemented', got: %s", err.Error())
	}
}
