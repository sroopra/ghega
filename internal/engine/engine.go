// Package engine wires the MLLP listener, message store, mapping engine,
// and HTTP sender into a single ADT A01 processing pipeline for Ghega.
package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sroopra/ghega/pkg/hl7v2"
	"github.com/sroopra/ghega/pkg/httpsender"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/mapping"
	"github.com/sroopra/ghega/pkg/mllp"
	"github.com/sroopra/ghega/pkg/payloadref"
)

// Config holds engine configuration.
type Config struct {
	DestinationURL string
}

// DefaultConfig returns the default engine configuration.
func DefaultConfig() Config {
	return Config{
		DestinationURL: "http://localhost:8081/webhook",
	}
}

// ConfigFromEnv builds Config from environment variables.
func ConfigFromEnv() Config {
	cfg := DefaultConfig()
	if u := os.Getenv("GHEGA_DESTINATION_URL"); u != "" {
		cfg.DestinationURL = u
	}
	return cfg
}

// Engine wires the MLLP listener, message store, mapping engine, and HTTP
// sender into a cohesive ADT A01 processing pipeline.
type Engine struct {
	Store          messagestore.Store
	Sender         *httpsender.Sender
	MappingEngine  *mapping.Engine
	Logger         *slog.Logger
	DestinationURL string
}

// NewEngine creates a new Engine with the given store and logger.
// If logger is nil, a discarding logger is used.
func NewEngine(store messagestore.Store, logger *slog.Logger) *Engine {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	}

	cfg := ConfigFromEnv()

	mapEngine := mapping.NewEngine([]mapping.Mapping{
		{Source: "PID-3.1", Target: "patient_mrn", Transform: mapping.TransformCopy},
		{Source: "PID-5.1", Target: "patient_last_name", Transform: mapping.TransformCopy},
		{Source: "PID-5.2", Target: "patient_first_name", Transform: mapping.TransformCopy},
		{Source: "MSH-9.1", Target: "message_type", Transform: mapping.TransformCopy},
	})

	sender := &httpsender.Sender{
		URL:     cfg.DestinationURL,
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/json"},
		Timeout: 5 * time.Second,
		Retries: 1,
		Logger:  logger,
	}

	return &Engine{
		Store:          store,
		Sender:         sender,
		MappingEngine:  mapEngine,
		Logger:         logger,
		DestinationURL: cfg.DestinationURL,
	}
}

// deriveChannelID extracts a channel identifier from the HL7 message type.
func deriveChannelID(msg *hl7v2.Message) string {
	msh := msg.Segment("MSH")
	if msh == nil {
		return "unknown"
	}
	msgType := msh.Field(9)
	if msgType != "" {
		parts := strings.Split(msgType, string(msg.ComponentSep))
		if len(parts) >= 2 {
			return fmt.Sprintf("%s-%s", parts[0], parts[1])
		}
		return msgType
	}
	return "unknown"
}

// HandleMessage processes a single HL7v2 message through the pipeline:
// persist, map, send via HTTP, update status, and return ACK.
// If any step fails, the error is logged and the status is set to "failed",
// but an ACK is still returned so the sender is not blocked.
func (e *Engine) HandleMessage(msg *hl7v2.Message) ([]byte, error) {
	msgID := uuid.NewString()
	channelID := deriveChannelID(msg)

	// Serialize the raw message for storage and mapping.
	raw := msg.Serialize()

	storageID := uuid.NewString()
	envelope := &payloadref.Envelope{
		ChannelID:  channelID,
		MessageID:  msgID,
		ReceivedAt: time.Now(),
		Status:     "received",
		Ref: payloadref.PayloadRef{
			StorageID: storageID,
			Location:  "sqlite://messages",
		},
	}

	ctx := context.Background()

	// Step 1: persist raw payload.
	if err := e.Store.Save(ctx, envelope, raw); err != nil {
		e.Logger.Error("failed to save message",
			slog.String("message_id", msgID),
			slog.String("channel_id", channelID),
			slog.String("error", err.Error()),
		)
		_ = e.Store.UpdateStatus(ctx, msgID, "failed")
		return hl7v2.GenerateACK(msg, "AA")
	}

	e.Logger.Info("message received",
		slog.String("message_id", msgID),
		slog.String("channel_id", channelID),
		slog.String("storage_id", storageID),
	)

	// Step 2: run mapping engine.
	mapped, err := e.MappingEngine.Apply(raw)
	if err != nil {
		e.Logger.Error("mapping failed",
			slog.String("message_id", msgID),
			slog.String("channel_id", channelID),
			slog.String("error", err.Error()),
		)
		_ = e.Store.UpdateStatus(ctx, msgID, "failed")
		return hl7v2.GenerateACK(msg, "AA")
	}

	// Step 3: serialize mapped output to JSON.
	jsonPayload, err := json.Marshal(mapped)
	if err != nil {
		e.Logger.Error("json marshal failed",
			slog.String("message_id", msgID),
			slog.String("channel_id", channelID),
			slog.String("error", err.Error()),
		)
		_ = e.Store.UpdateStatus(ctx, msgID, "failed")
		return hl7v2.GenerateACK(msg, "AA")
	}

	// Step 4: send JSON payload via HTTP.
	sendCtx, cancel := context.WithTimeout(ctx, e.Sender.Timeout)
	defer cancel()

	resp, err := e.Sender.Send(sendCtx, jsonPayload)
	if err != nil {
		e.Logger.Error("http send failed",
			slog.String("message_id", msgID),
			slog.String("channel_id", channelID),
			slog.String("error", err.Error()),
		)
		_ = e.Store.UpdateStatus(ctx, msgID, "failed")
		return hl7v2.GenerateACK(msg, "AA")
	}

	e.Logger.Info("message delivered",
		slog.String("message_id", msgID),
		slog.String("channel_id", channelID),
		slog.Int("status_code", resp.StatusCode),
	)

	// Step 5: update status to delivered.
	if err := e.Store.UpdateStatus(ctx, msgID, "delivered"); err != nil {
		e.Logger.Error("failed to update status",
			slog.String("message_id", msgID),
			slog.String("channel_id", channelID),
			slog.String("error", err.Error()),
		)
	}

	return hl7v2.GenerateACK(msg, "AA")
}

// MLLPHandler returns an mllp.Handler that delegates to Engine.HandleMessage.
func (e *Engine) MLLPHandler() mllp.Handler {
	return func(msg *hl7v2.Message) ([]byte, error) {
		return e.HandleMessage(msg)
	}
}
