// Package fhirsender provides a FHIR destination connector for Ghega.
//
// The sender transmits FHIR JSON (individual resources or Bundles) to external
// FHIR servers. It automatically sets the correct FHIR content type headers and
// retries on transport errors or 5xx responses. Only metadata is logged;
// payload bytes are never logged.
package fhirsender

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Response is a structured HTTP response.
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Sender sends FHIR JSON payloads to a configured FHIR server base URL.
// It never logs payload bytes.
type Sender struct {
	URL     string
	Method  string
	Headers map[string]string
	Timeout time.Duration
	Retries int
	Logger  *slog.Logger
}

func (s *Sender) method() string {
	if s.Method != "" {
		return s.Method
	}
	return http.MethodPost
}

func (s *Sender) timeout() time.Duration {
	if s.Timeout > 0 {
		return s.Timeout
	}
	return 30 * time.Second
}

func (s *Sender) retries() int {
	if s.Retries > 0 {
		return s.Retries
	}
	return 3
}

func (s *Sender) logger() *slog.Logger {
	if s.Logger != nil {
		return s.Logger
	}
	return slog.Default()
}

// Send transmits the FHIR payload to the configured URL.
// It retries on transport errors or 5xx status codes up to the configured retry count.
// Only metadata is logged; payload bytes are never logged.
func (s *Sender) Send(ctx context.Context, payload []byte) (*Response, error) {
	if s.URL == "" {
		return nil, fmt.Errorf("fhirsender: URL is required")
	}

	client := &http.Client{Timeout: s.timeout()}
	logger := s.logger()

	var lastErr error
	maxRetries := s.retries()

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, s.method(), s.URL, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("fhirsender: failed to create request: %w", err)
		}

		// Set FHIR-specific headers
		req.Header.Set("Content-Type", "application/fhir+json; fhirVersion=4.0")
		req.Header.Set("Accept", "application/fhir+json")

		// Apply custom headers after defaults so they can override if needed
		for k, v := range s.Headers {
			req.Header.Set(k, v)
		}

		logger.Info("sending FHIR request",
			slog.String("url", s.URL),
			slog.String("method", s.method()),
			slog.Int("attempt", attempt),
			slog.Int("max_retries", maxRetries),
		)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			logger.Error("FHIR request failed",
				slog.String("url", s.URL),
				slog.String("method", s.method()),
				slog.Int("attempt", attempt),
				slog.String("error", err.Error()),
			)
			if attempt < maxRetries {
				time.Sleep(time.Second)
			}
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			logger.Error("failed to read response body",
				slog.String("url", s.URL),
				slog.Int("status_code", resp.StatusCode),
				slog.Int("attempt", attempt),
				slog.String("error", err.Error()),
			)
			if attempt < maxRetries {
				time.Sleep(time.Second)
			}
			continue
		}

		logger.Info("FHIR response received",
			slog.String("url", s.URL),
			slog.Int("status_code", resp.StatusCode),
			slog.Int("attempt", attempt),
		)

		// Retry on 5xx status codes
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			lastErr = fmt.Errorf("fhirsender: server returned status %d", resp.StatusCode)
			if attempt < maxRetries {
				time.Sleep(time.Second)
				continue
			}
			return nil, fmt.Errorf("fhirsender: failed after %d attempts: %w", maxRetries, lastErr)
		}

		// 4xx errors are not retried
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, fmt.Errorf("fhirsender: server returned status %d", resp.StatusCode)
		}

		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header.Clone(),
			Body:       body,
		}, nil
	}

	return nil, fmt.Errorf("fhirsender: failed after %d attempts: %w", maxRetries, lastErr)
}

// DryRun validates the payload without sending it over the network.
// It checks that the sender is properly configured and the payload is non-nil.
func (s *Sender) DryRun(payload []byte) error {
	if s.URL == "" {
		return fmt.Errorf("fhirsender: URL is required")
	}
	if payload == nil {
		return fmt.Errorf("fhirsender: payload is nil")
	}
	return nil
}
