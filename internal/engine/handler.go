// Package engine wires the MLLP listener, message store, mapping engine, and
// HTTP sender into a cohesive message-processing pipeline.
package engine

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sroopra/ghega/pkg/hl7v2"
	"github.com/sroopra/ghega/pkg/httpsender"
	"github.com/sroopra/ghega/pkg/mapping"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/mllp"
	"github.com/sroopra/ghega/pkg/payloadref"
)

// DefaultMappings is the built-in ADT A01 mapping set used when no channel
// configuration is loaded.
var DefaultMappings = []mapping.Mapping{
	{Source: "PID-3.1", Target: "patient_mrn"},
	{Source: "PID-5.1", Target: "patient_last_name"},
	{Source: "PID-5.2", Target: "patient_first_name"},
	{Source: "MSH-9.1", Target: "message_type"},
}

// HandlerConfig holds runtime configuration for the message handler.
type HandlerConfig struct {
	ChannelID       string
	DestinationURL  string
	DestinationMethod string
	Timeout         time.Duration
	Retries         int
	Mappings        []mapping.Mapping
}

// DefaultHandlerConfig returns a sensible default configuration.
func DefaultHandlerConfig() HandlerConfig {
	return HandlerConfig{
		ChannelID:         "adt-a01",
		DestinationURL:    os.Getenv("GHEGA_DESTINATION_URL"),
		DestinationMethod: "POST",
		Timeout:           5 * time.Second,
		Retries:           1,
		Mappings:          DefaultMappings,
	}
}

// NewMLLPHandler creates an MLLP handler that processes HL7 messages through
// the full pipeline: persist → map → send HTTP → update status.
func NewMLLPHandler(store messagestore.Store, cfg HandlerConfig, logger *slog.Logger) mllp.Handler {
	if cfg.DestinationURL == "" {
		cfg.DestinationURL = "http://localhost:8081/webhook"
	}
	if logger == nil {
		logger = slog.Default()
	}

	mapper := mapping.NewEngine(cfg.Mappings)
	sender := &httpsender.Sender{
		URL:     cfg.DestinationURL,
		Method:  cfg.DestinationMethod,
		Timeout: cfg.Timeout,
		Retries: cfg.Retries,
		Logger:  logger,
	}

	return func(msg *hl7v2.Message) ([]byte, error) {
		ctx := context.Background()

		// 1. Generate IDs.
		messageID := uuid.Must(uuid.NewV7()).String()
		storageID := "payload:" + messageID

		// 2. Persist raw payload.
		rawPayload := msg.Serialize()
		env := &payloadref.Envelope{
			ChannelID:  cfg.ChannelID,
			MessageID:  messageID,
			ReceivedAt: time.Now(),
			Status:     "received",
			Ref: payloadref.PayloadRef{
				StorageID: storageID,
				Location:  "sqlite://messages",
			},
		}

		if err := store.Save(ctx, env, rawPayload); err != nil {
			logger.Error("failed to save message",
				slog.String("message_id", messageID),
				slog.String("error", err.Error()),
			)
			// Still return AA so the sender doesn't retry endlessly.
			ack, _ := hl7v2.GenerateACK(msg, "AA")
			return ack, nil
		}

		logger.Info("message received",
			slog.String("message_id", messageID),
			slog.String("channel_id", cfg.ChannelID),
		)

		// 3. Apply mappings.
		mapped, err := mapper.Apply(rawPayload)
		if err != nil {
			logger.Error("mapping failed",
				slog.String("message_id", messageID),
				slog.String("error", err.Error()),
			)
			_ = updateStatus(ctx, store, env, "mapping_failed")
			ack, _ := hl7v2.GenerateACK(msg, "AA")
			return ack, nil
		}

		// 4. Send to HTTP destination.
		jsonPayload, _ := json.Marshal(mapped)
		_, sendErr := sender.Send(ctx, jsonPayload)

		if sendErr != nil {
			logger.Error("destination send failed",
				slog.String("message_id", messageID),
				slog.String("error", sendErr.Error()),
			)
			_ = updateStatus(ctx, store, env, "failed")
		} else {
			logger.Info("message delivered",
				slog.String("message_id", messageID),
				slog.String("destination", cfg.DestinationURL),
			)
			_ = updateStatus(ctx, store, env, "delivered")
		}

		// 5. Return ACK.
		ack, _ := hl7v2.GenerateACK(msg, "AA")
		return ack, nil
	}
}

func updateStatus(ctx context.Context, store messagestore.Store, env *payloadref.Envelope, status string) error {
	return store.UpdateStatus(ctx, env.MessageID, status)
}
