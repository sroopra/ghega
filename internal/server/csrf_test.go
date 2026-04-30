package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sroopra/ghega/internal/config"
	"github.com/sroopra/ghega/internal/session"
	"github.com/sroopra/ghega/pkg/messagestore"
	"golang.org/x/oauth2"
)

func TestCSRF_GET_PassesWithoutToken(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/channels")
	if err != nil {
		t.Fatalf("get request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestCSRF_POST_MatchingToken_Passes(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	cookie, token := getCSRFTokens(t, ts)

	req, _ := http.NewRequest("POST", ts.URL+"/api/v1/messages/msg-001/redeliver", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(csrfHeaderName, token)
	req.AddCookie(cookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotImplemented)
	}
}

func TestCSRF_POST_MissingToken_Fails403(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/api/v1/messages/msg-001/redeliver", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `"error":"csrf token mismatch"`) {
		t.Errorf("body = %q, want csrf token mismatch error", body)
	}
}

func TestCSRF_POST_MismatchedToken_Fails403(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	cookie, _ := getCSRFTokens(t, ts)

	req, _ := http.NewRequest("POST", ts.URL+"/api/v1/messages/msg-001/redeliver", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(csrfHeaderName, "bad-token")
	req.AddCookie(cookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `"error":"csrf token mismatch"`) {
		t.Errorf("body = %q, want csrf token mismatch error", body)
	}
}

func TestCSRF_TokenReturnedInResponseHeader(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/channels")
	if err != nil {
		t.Fatalf("get request: %v", err)
	}
	defer resp.Body.Close()

	token := resp.Header.Get(csrfHeaderName)
	if token == "" {
		t.Error("missing X-CSRF-Token header in response")
	}
}

func TestCSRF_CookieSetOnEveryRequest(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/channels")
	if err != nil {
		t.Fatalf("get request: %v", err)
	}
	defer resp.Body.Close()

	var csrfCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == csrfCookieName {
			csrfCookie = c
			break
		}
	}
	if csrfCookie == nil {
		t.Fatal("missing csrf cookie")
	}
	if csrfCookie.Value == "" {
		t.Error("csrf cookie value is empty")
	}
	if csrfCookie.Path != "/" {
		t.Errorf("cookie path = %q, want %q", csrfCookie.Path, "/")
	}
	if csrfCookie.HttpOnly {
		t.Error("expected csrf cookie HttpOnly = false")
	}
	if csrfCookie.Secure {
		t.Error("expected csrf cookie Secure = false")
	}
	if csrfCookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("cookie SameSite = %v, want Strict", csrfCookie.SameSite)
	}
}

func TestCSRF_SkipsAuthCallback(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	alertStore := newTestAlertStore()
	mgr := session.NewManager(session.NewMemoryStore(), "test-secret")
	op := &OIDCProvider{
		config: oauth2.Config{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			RedirectURL:  "http://localhost:8080/auth/callback",
			Endpoint: oauth2.Endpoint{
				AuthURL:  "http://localhost:9999/auth",
				TokenURL: "http://localhost:9999/token",
			},
			Scopes: []string{"openid", "profile", "email"},
		},
		mgr: mgr,
	}
	cfg := config.AuthConfig{Enabled: true, SessionSecret: "test-secret"}
	srv := New(store, alertStore, WithAuthConfig(cfg), WithSessionManager(mgr), WithOIDCProvider(op))
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/auth/callback?state=xxx&code=yyy", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("callback request: %v", err)
	}
	defer resp.Body.Close()

	// Should get 400 (missing state cookie), not 403 (csrf mismatch)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestCSRF_SkipsHealthz(t *testing.T) {
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
}

func TestCSRF_ReusesExistingCookie(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	cookie, token := getCSRFTokens(t, ts)

	// Make another GET request with the existing cookie
	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/channels", nil)
	req.AddCookie(cookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("second get request: %v", err)
	}
	defer resp.Body.Close()

	// The response header should contain the same token
	newToken := resp.Header.Get(csrfHeaderName)
	if newToken != token {
		t.Errorf("token changed: got %q, want %q", newToken, token)
	}

	// The cookie should be the same
	var newCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == csrfCookieName {
			newCookie = c
			break
		}
	}
	if newCookie == nil {
		t.Fatal("missing csrf cookie in second response")
	}
	if newCookie.Value != cookie.Value {
		t.Errorf("cookie value changed: got %q, want %q", newCookie.Value, cookie.Value)
	}
}

func TestCSRF_PUT_Delete_Patch_AlsoRequireToken(t *testing.T) {
	store := messagestore.NewInMemoryStore()
	srv := New(store, newTestAlertStore())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	for _, method := range []string{"PUT", "DELETE", "PATCH"} {
		req, _ := http.NewRequest(method, ts.URL+"/api/v1/messages/msg-001", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s request: %v", method, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("%s status = %d, want %d", method, resp.StatusCode, http.StatusForbidden)
		}
	}
}
