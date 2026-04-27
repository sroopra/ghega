package alerts

import (
	"sync"
	"time"
)

// Severity represents the severity level of an alert.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Alert represents a system alert triggered by a channel event.
// No PHI is stored in Alert values.
type Alert struct {
	ID          string     `json:"id"`
	ChannelID   string     `json:"channel_id"`
	Severity    Severity   `json:"severity"`
	Message     string     `json:"message"`
	CreatedAt   time.Time  `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
}

// AlertStore persists and queries alerts.
type AlertStore interface {
	Create(a *Alert) error
	List() ([]*Alert, error)
	Acknowledge(id string) error
	Resolve(id string) error
}

// InMemoryAlertStore is a thread-safe in-memory implementation of AlertStore.
type InMemoryAlertStore struct {
	mu     sync.RWMutex
	alerts map[string]*Alert
}

// NewInMemoryAlertStore creates a new empty InMemoryAlertStore.
func NewInMemoryAlertStore() *InMemoryAlertStore {
	return &InMemoryAlertStore{
		alerts: make(map[string]*Alert),
	}
}

// Create persists a new alert.
func (s *InMemoryAlertStore) Create(a *Alert) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := *a
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = time.Now().UTC()
	}
	s.alerts[cp.ID] = &cp
	return nil
}

// List returns all alerts in insertion order.
func (s *InMemoryAlertStore) List() ([]*Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*Alert, 0, len(s.alerts))
	for _, a := range s.alerts {
		cp := *a
		out = append(out, &cp)
	}
	return out, nil
}

// Acknowledge marks an alert as acknowledged.
func (s *InMemoryAlertStore) Acknowledge(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.alerts[id]
	if !ok {
		return &ErrAlertNotFound{ID: id}
	}
	now := time.Now().UTC()
	a.AcknowledgedAt = &now
	return nil
}

// Resolve marks an alert as resolved.
func (s *InMemoryAlertStore) Resolve(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.alerts[id]
	if !ok {
		return &ErrAlertNotFound{ID: id}
	}
	now := time.Now().UTC()
	a.ResolvedAt = &now
	return nil
}

// ErrAlertNotFound is returned when an alert ID does not exist in the store.
type ErrAlertNotFound struct {
	ID string
}

func (e *ErrAlertNotFound) Error() string {
	return "alert not found: " + e.ID
}

// TriggerLog logs a placeholder alert when a message fails.
// This is a no-op implementation that writes to the provided logger.
func TriggerLog(log func(string, ...any), channelID string, messageID string, err error) {
	log("[Ghega Alert] channel=%s message=%s error=%v", channelID, messageID, err)
}
