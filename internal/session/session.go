// Package session implements cookie-based session management for the Ghega BFF.
package session

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated user session.
type Session struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Store persists and queries sessions.
type Store interface {
	Get(id string) *Session
	Save(s *Session)
	Delete(id string)
}

// MemoryStore is a thread-safe in-memory implementation of Store.
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewMemoryStore creates a new empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]*Session),
	}
}

// Get retrieves a session by ID. Returns nil if not found.
func (m *MemoryStore) Get(id string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.sessions[id]
	if !ok {
		return nil
	}
	cp := *s
	return &cp
}

// Save persists a session.
func (m *MemoryStore) Save(s *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cp := *s
	m.sessions[s.ID] = &cp
}

// Delete removes a session by ID.
func (m *MemoryStore) Delete(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, id)
}

const (
	cookieName    = "ghega_session"
	sessionMaxAge = 24 * time.Hour
)

// serializeSession serializes a session to a signed cookie value.
// Format: base64(sessionJSON).base64(signature)
func serializeSession(s *Session, secret string) (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString(data)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encoded))
	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	return encoded + "." + sig, nil
}

// parseSession verifies and deserializes a signed cookie value.
func parseSession(value string, secret string) (*Session, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid cookie format")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0]))

	expectedSig, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	if !hmac.Equal(mac.Sum(nil), expectedSig) {
		return nil, errors.New("invalid signature")
	}

	data, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Manager ties a Store and signing secret together to create, read, and destroy sessions.
type Manager struct {
	store  Store
	secret string
}

// NewManager creates a new Manager with the given store and secret.
func NewManager(store Store, secret string) *Manager {
	return &Manager{
		store:  store,
		secret: secret,
	}
}

// CreateSession creates a new session, persists it, and writes the signed cookie.
func (m *Manager) CreateSession(w http.ResponseWriter, email, name string, roles []string) {
	now := time.Now().UTC()
	s := &Session{
		ID:        uuid.NewString(),
		Email:     email,
		Name:      name,
		Roles:     append([]string(nil), roles...),
		CreatedAt: now,
		ExpiresAt: now.Add(sessionMaxAge),
	}

	m.store.Save(s)

	value, err := serializeSession(s, m.secret)
	if err != nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(sessionMaxAge.Seconds()),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

// ReadSession reads and verifies the session cookie from the request.
// Returns nil if the cookie is missing, tampered with, expired, or not found in the store.
func (m *Manager) ReadSession(r *http.Request) *Session {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return nil
	}

	s, err := parseSession(c.Value, m.secret)
	if err != nil {
		return nil
	}

	stored := m.store.Get(s.ID)
	if stored == nil {
		return nil
	}

	if time.Now().UTC().After(stored.ExpiresAt) {
		m.store.Delete(stored.ID)
		return nil
	}

	return stored
}

// DestroySession clears the session cookie.
func (m *Manager) DestroySession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

// contextKey is the private type used for context values.
type contextKey struct{}

var sessionContextKey = &contextKey{}

// ContextWithSession returns a new context with the session attached.
func ContextWithSession(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionContextKey, s)
}

// SessionFromContext retrieves the session from the context.
// Returns nil if no session is present.
func SessionFromContext(ctx context.Context) *Session {
	s, _ := ctx.Value(sessionContextKey).(*Session)
	return s
}
