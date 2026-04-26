package channel

import (
	"context"
	"encoding/json"
	"fmt"
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

// Channel represents a configured integration channel that wires together
// the MLLP listener, message store, mapping engine, and HTTP sender.
type Channel struct {
	cfg      *Config
	store    messagestore.Store
	logger   *logging.Logger
	listener *mllp.Listener
}

// NewChannel creates a Channel from configuration and dependencies.
func NewChannel(cfg *Config, store messagestore.Store, logger *logging.Logger) *Channel {
	return &Channel{
		cfg:    cfg,
		store:  store,
		logger: logger,
	}
}

// Run starts the MLLP listener and begins processing messages.
// It blocks until the listener is stopped or encounters a fatal error.
func (ch *Channel) Run() error {
	engine := mapping.NewEngine(ch.cfg.Mapping.Fields)

	sender := &httpsender.Sender{
		URL:     ch.cfg.Destination.URL,
		Method:  ch.cfg.Destination.Method,
		Timeout: time.Duration(ch.cfg.Destination.Timeout) * time.Second,
		Retries: 3,
		Logger:  ch.logger.Inner(),
	}

	handler := func(msg *hl7v2.Message) ([]byte, error) {
		return ch.handleMessage(msg, engine, sender)
	}

	mllpCfg := mllp.Config{
		Host: ch.cfg.Source.Host,
		Port: ch.cfg.Source.Port,
	}

	ch.listener = mllp.NewListener(mllpCfg, handler, ch.logger.Inner())
	if err := ch.listener.Start(); err != nil {
		return fmt.Errorf("channel %q: %w", ch.cfg.Name, err)
	}

	return nil
}

// Stop gracefully shuts down the channel.
func (ch *Channel) Stop() error {
	if ch.listener == nil {
		return nil
	}
	return ch.listener.Stop()
}

// Addr returns the listener's network address, or nil if not started.
func (ch *Channel) Addr() string {
	if ch.listener == nil {
		return ""
	}
	addr := ch.listener.Addr()
	if addr == nil {
		return ""
	}
	return addr.String()
}

func (ch *Channel) handleMessage(msg *hl7v2.Message, engine *mapping.Engine, sender *httpsender.Sender) ([]byte, error) {
	ctx := context.Background()
	messageID := uuid.NewString()
	receivedAt := time.Now().UTC()
	storageID := uuid.NewString()

	rawPayload := msg.Serialize()

	ref := payloadref.PayloadRef{
		StorageID: storageID,
		Location:  "memory",
	}

	env := &payloadref.Envelope{
		ChannelID:  ch.cfg.Name,
		MessageID:  messageID,
		ReceivedAt: receivedAt,
		Status:     "received",
		Ref:        ref,
	}

	if err := ch.store.Save(ctx, env, rawPayload); err != nil {
		ch.logger.LogMessageFailed(ch.cfg.Name, messageID, err)
		return hl7v2.GenerateACK(msg, "AE")
	}

	ch.logger.LogMessageReceived(ch.cfg.Name, messageID, receivedAt, ref)

	mapped, err := engine.Apply(rawPayload)
	if err != nil {
		ch.logger.LogMessageFailed(ch.cfg.Name, messageID, err)
		return hl7v2.GenerateACK(msg, "AE")
	}

	mappedPayload, err := json.Marshal(mapped)
	if err != nil {
		ch.logger.LogMessageFailed(ch.cfg.Name, messageID, err)
		return hl7v2.GenerateACK(msg, "AE")
	}

	_, err = sender.Send(ctx, mappedPayload)
	if err != nil {
		ch.logger.LogMessageFailed(ch.cfg.Name, messageID, err)
		return hl7v2.GenerateACK(msg, "AE")
	}

	ch.logger.LogMessageProcessed(ch.cfg.Name, messageID, time.Now().UTC(), ref)

	return hl7v2.GenerateACK(msg, "AA")
}

