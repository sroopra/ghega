package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSerializeParse_RoundTrip(t *testing.T) {
	t.Helper()

	original := &Session{
		ID:        "sess-123",
		Email:     "test@example.com",
		Name:      "Test User",
		Roles:     []string{"admin", "user"},
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	secret := "super-secret-key"
	encoded, err := serializeSession(original, secret)
	if err != nil {
		t.Fatalf("serializeSession failed: %v", err)
	}

	decoded, err := parseSession(encoded, secret)
	if err != nil {
		t.Fatalf("parseSession failed: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Email != original.Email {
		t.Errorf("Email = %q, want %q", decoded.Email, original.Email)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if len(decoded.Roles) != len(original.Roles) {
		t.Fatalf("len(Roles) = %d, want %d", len(decoded.Roles), len(original.Roles))
	}
	for i, r := range decoded.Roles {
		if r != original.Roles[i] {
			t.Errorf("Roles[%d] = %q, want %q", i, r, original.Roles[i])
		}
	}
	if !decoded.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", decoded.CreatedAt, original.CreatedAt)
	}
	if !decoded.ExpiresAt.Equal(original.ExpiresAt) {
		t.Errorf("ExpiresAt = %v, want %v", decoded.ExpiresAt, original.ExpiresAt)
	}
}

func TestParseSession_TamperedCookie(t *testing.T) {
	t.Helper()

	original := &Session{
		ID:        "sess-456",
		Email:     "alice@example.com",
		Name:      "Alice",
		Roles:     []string{"user"},
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	secret := "correct-secret"
	encoded, err := serializeSession(original, secret)
	if err != nil {
		t.Fatalf("serializeSession failed: %v", err)
	}

	cases := []struct {
		name  string
		value string
	}{
		{
			name:  "wrong secret",
			value: encoded,
		},
		{
			name: "modified payload",
			value: func() string {
				parts := strings.Split(encoded, ".")
				// Flip a character in the payload so the signature no longer matches.
				if len(parts[0]) > 0 {
					b := []byte(parts[0])
					if b[0] == 'A' {
						b[0] = 'B'
					} else {
						b[0] = 'A'
					}
					parts[0] = string(b)
				}
				return strings.Join(parts, ".")
			}(),
		},
		{
			name:  "missing dot",
			value: strings.Replace(encoded, ".", "", 1),
		},
		{
			name:  "garbage",
			value: "not.a.cookie",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			secretToUse := secret
			if tc.name == "wrong secret" {
				secretToUse = "wrong-secret"
			}

			_, err := parseSession(tc.value, secretToUse)
			if err == nil {
				t.Error("expected error for tampered cookie, got nil")
			}
		})
	}
}

func TestManager_ExpiredSession(t *testing.T) {
	t.Helper()

	store := NewMemoryStore()
	mgr := NewManager(store, "test-secret")

	// Manually insert an expired session into the store and create a valid cookie for it.
	expired := &Session{
		ID:        "expired-sess",
		Email:     "old@example.com",
		Name:      "Old User",
		Roles:     []string{"user"},
		CreatedAt: time.Now().UTC().Add(-48 * time.Hour),
		ExpiresAt: time.Now().UTC().Add(-24 * time.Hour),
	}
	store.Save(expired)

	cookieValue, err := serializeSession(expired, "test-secret")
	if err != nil {
		t.Fatalf("serializeSession failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: cookieValue,
	})

	session := mgr.ReadSession(req)
	if session != nil {
		t.Error("expected nil session for expired cookie")
	}

	// The expired session should have been deleted from the store.
	if store.Get("expired-sess") != nil {
		t.Error("expected expired session to be removed from store")
	}
}

func TestContextWithSession(t *testing.T) {
	t.Helper()

	s := &Session{
		ID:    "ctx-sess",
		Email: "ctx@example.com",
		Name:  "Context User",
		Roles: []string{"admin"},
	}

	ctx := context.Background()
	if got := SessionFromContext(ctx); got != nil {
		t.Error("expected nil session from empty context")
	}

	ctx = ContextWithSession(ctx, s)
	got := SessionFromContext(ctx)
	if got == nil {
		t.Fatal("expected session from context, got nil")
	}
	if got.ID != s.ID {
		t.Errorf("ID = %q, want %q", got.ID, s.ID)
	}
	if got.Email != s.Email {
		t.Errorf("Email = %q, want %q", got.Email, s.Email)
	}
}

func TestMemoryStore_Concurrency(t *testing.T) {
	t.Helper()

	store := NewMemoryStore()
	var wg sync.WaitGroup

	// Writers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			store.Save(&Session{
				ID:        "sess-" + string(rune('0'+idx%10)),
				Email:     "user@example.com",
				Name:      "User",
				Roles:     []string{"user"},
				CreatedAt: time.Now().UTC(),
				ExpiresAt: time.Now().UTC().Add(time.Hour),
			})
		}(i)
	}

	// Readers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = store.Get("sess-" + string(rune('0'+idx%10)))
		}(i)
	}

	// Deleters
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			store.Delete("sess-" + string(rune('0'+idx%10)))
		}(i)
	}

	wg.Wait()

	// The test passes if no data race is detected (run with -race).
}

func TestManager_CreateAndReadSession(t *testing.T) {
	t.Helper()

	store := NewMemoryStore()
	mgr := NewManager(store, "mgr-secret")

	// Create session via Manager
	w := httptest.NewRecorder()
	mgr.CreateSession(w, "bob@example.com", "Bob", []string{"editor"})

	resp := w.Result()
	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != cookieName {
		t.Errorf("cookie name = %q, want %q", c.Name, cookieName)
	}
	if c.HttpOnly != true {
		t.Error("expected HttpOnly cookie")
	}
	if c.Secure != false {
		t.Error("expected Secure=false")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want %v", c.SameSite, http.SameSiteLaxMode)
	}
	if c.Path != "/" {
		t.Errorf("Path = %q, want %q", c.Path, "/")
	}
	if c.MaxAge != int(sessionMaxAge.Seconds()) {
		t.Errorf("MaxAge = %d, want %d", c.MaxAge, int(sessionMaxAge.Seconds()))
	}

	// Read session back
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(c)

	session := mgr.ReadSession(req)
	if session == nil {
		t.Fatal("expected session, got nil")
	}
	if session.Email != "bob@example.com" {
		t.Errorf("Email = %q, want %q", session.Email, "bob@example.com")
	}
	if session.Name != "Bob" {
		t.Errorf("Name = %q, want %q", session.Name, "Bob")
	}
	if len(session.Roles) != 1 || session.Roles[0] != "editor" {
		t.Errorf("Roles = %v, want [editor]", session.Roles)
	}
}

func TestManager_DestroySession(t *testing.T) {
	t.Helper()

	store := NewMemoryStore()
	mgr := NewManager(store, "mgr-secret")

	w := httptest.NewRecorder()
	mgr.CreateSession(w, "alice@example.com", "Alice", []string{"admin"})

	resp := w.Result()
	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	// Destroy
	w2 := httptest.NewRecorder()
	mgr.DestroySession(w2)

	resp2 := w2.Result()
	cookies2 := resp2.Cookies()
	if len(cookies2) != 1 {
		t.Fatalf("expected 1 cookie after destroy, got %d", len(cookies2))
	}
	c := cookies2[0]
	if c.Name != cookieName {
		t.Errorf("cookie name = %q, want %q", c.Name, cookieName)
	}
	if c.Value != "" {
		t.Errorf("cookie value = %q, want empty", c.Value)
	}
	if c.MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", c.MaxAge)
	}
}

func TestManager_ReadSession_MissingCookie(t *testing.T) {
	t.Helper()

	store := NewMemoryStore()
	mgr := NewManager(store, "mgr-secret")

	req := httptest.NewRequest("GET", "/", nil)
	if got := mgr.ReadSession(req); got != nil {
		t.Error("expected nil for missing cookie")
	}
}

func TestManager_ReadSession_UnknownSessionInStore(t *testing.T) {
	t.Helper()

	store := NewMemoryStore()
	mgr := NewManager(store, "mgr-secret")

	s := &Session{
		ID:        "orphan",
		Email:     "orphan@example.com",
		Name:      "Orphan",
		Roles:     []string{},
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	cookieValue, err := serializeSession(s, "mgr-secret")
	if err != nil {
		t.Fatalf("serializeSession failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: cookieValue})

	// Session is not in the store
	if got := mgr.ReadSession(req); got != nil {
		t.Error("expected nil for session not in store")
	}
}
