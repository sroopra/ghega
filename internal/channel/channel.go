package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/sroopra/ghega/internal/logging"
	"github.com/sroopra/ghega/pkg/hl7v2"
	"github.com/sroopra/ghega/pkg/httpsender"
	"github.com/sroopra/ghega/pkg/mapping"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/mllp"
	"github.com/sroopra/ghega/pkg/payloadref"
)

// Channel wires together MLLP ingestion, storage, mapping, and HTTP delivery.
type Channel struct {
	Config   Config
	Store    messagestore.Store
	Sender   *httpsender.Sender
	Logger   *logging.Logger
	listener *mllp.Listener
}

// NewChannel creates a Channel from configuration and dependencies.
func NewChannel(cfg Config, store messagestore.Store, sender *httpsender.Sender, logger *logging.Logger) *Channel {
	return &Channel{
		Config: cfg,
		Store:  store,
		Sender: sender,
		Logger: logger,
	}
}

// Run starts the MLLP listener and blocks until the listener accepts its first
// connection. The listener continues to run in the background. Use Stop to shut
// it down.
func (c *Channel) Run() error {
	mllpCfg := mllp.Config{
		Host: c.Config.Source.Host,
		Port: c.Config.Source.Port,
	}

	handler := func(msg *hl7v2.Message) ([]byte, error) {
		return c.handleMessage(msg)
	}

	c.listener = mllp.NewListener(mllpCfg, handler, c.Logger.SLogger())
	if err := c.listener.Start(); err != nil {
		return fmt.Errorf("channel %s: %w", c.Config.Name, err)
	}
	return nil
}

// Stop gracefully shuts down the channel's MLLP listener.
func (c *Channel) Stop() error {
	if c.listener == nil {
		return nil
	}
	return c.listener.Stop()
}

// Addr returns the listener's network address, or nil if not started.
func (c *Channel) Addr() string {
	if c.listener == nil {
		return ""
	}
	addr := c.listener.Addr()
	if addr == nil {
		return ""
	}
	return addr.String()
}

func (c *Channel) handleMessage(msg *hl7v2.Message) ([]byte, error) {
	ctx := context.Background()
	receivedAt := time.Now().UTC()

	// Serialize the parsed message back to raw HL7v2 bytes for storage and mapping.
	raw := msg.Serialize()

	// Derive a message ID from MSH-10 if present; otherwise generate a UUID.
	messageID := c.extractMessageID(msg)

	// Build the payload reference.
	envelope := &payloadref.Envelope{
		ChannelID:  c.Config.Name,
		MessageID:  messageID,
		ReceivedAt: receivedAt,
		Status:     "received",
		Ref: payloadref.PayloadRef{
			StorageID: messageID,
			Location:  "memory",
		},
	}

	// 2. Save metadata + payload reference to the message store.
	if err := c.Store.Save(ctx, envelope, raw); err != nil {
		c.Logger.LogMessageFailed(c.Config.Name, messageID, fmt.Errorf("store save: %w", err))
		ack, _ := hl7v2.GenerateACK(msg, "AE")
		return ack, nil
	}

	c.Logger.LogMessageReceived(c.Config.Name, messageID, receivedAt, envelope.Ref)

	// 3. Apply mappings using the mapping engine.
	engine := c.buildEngine()
	mapped, err := engine.Apply(raw)
	if err != nil {
		c.Logger.LogMessageFailed(c.Config.Name, messageID, fmt.Errorf("mapping apply: %w", err))
		ack, _ := hl7v2.GenerateACK(msg, "AE")
		return ack, nil
	}

	// 4. Send the transformed payload via HTTP sender.
	mappedPayload, err := json.Marshal(mapped)
	if err != nil {
		c.Logger.LogMessageFailed(c.Config.Name, messageID, fmt.Errorf("json marshal: %w", err))
		ack, _ := hl7v2.GenerateACK(msg, "AE")
		return ack, nil
	}

	resp, err := c.Sender.Send(ctx, mappedPayload)
	if err != nil {
		c.Logger.LogMessageFailed(c.Config.Name, messageID, fmt.Errorf("http send: %w", err))
		ack, _ := hl7v2.GenerateACK(msg, "AE")
		return ack, nil
	}

	// 5. Log the result (metadata only).
	c.Logger.LogInfo("message delivered",
		slog.String("channel_id", c.Config.Name),
		slog.String("message_id", messageID),
		slog.Int("http_status", resp.StatusCode),
		slog.String("destination", c.Config.Destination.URL),
	)

	c.Logger.LogMessageProcessed(c.Config.Name, messageID, time.Now().UTC(), envelope.Ref)

	// 6. Generate and send ACK.
	ack, err := hl7v2.GenerateACK(msg, "AA")
	if err != nil {
		c.Logger.LogMessageFailed(c.Config.Name, messageID, fmt.Errorf("ack generation: %w", err))
		// Return a generic AE if we can't generate a proper ACK
		ack, _ = hl7v2.GenerateACK(msg, "AE")
		return ack, nil
	}
	return ack, nil
}

func (c *Channel) extractMessageID(msg *hl7v2.Message) string {
	msh := msg.Segment("MSH")
	if msh != nil {
		if id := msh.Field(10); id != "" {
			return id
		}
	}
	return uuid.NewString()
}

func (c *Channel) buildEngine() *mapping.Engine {
	mappings := c.Config.Mapping.Mappings
	if len(mappings) == 0 {
		// Default identity mapping if none configured.
		mappings = []mapping.Mapping{}
	}
	return mapping.NewEngine(mappings)
}
