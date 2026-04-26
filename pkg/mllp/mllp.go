// Package mllp implements the Minimal Lower Layer Protocol (MLLP) for Ghega.
// It provides TCP-based listeners that accept HL7v2 messages framed with
// MLLP start/end block bytes and respond with ACK messages.
package mllp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sroopra/ghega/pkg/hl7v2"
)

// Frame boundaries per the MLLP specification.
const (
	StartBlock = 0x0B // vertical tab
	EndBlock   = 0x1C // file separator
	CarriageReturn = 0x0D // carriage return
)

// EncodeFrame wraps a raw HL7 payload in MLLP framing bytes.
func EncodeFrame(payload []byte) []byte {
	frame := make([]byte, 0, len(payload)+3)
	frame = append(frame, StartBlock)
	frame = append(frame, payload...)
	frame = append(frame, EndBlock, CarriageReturn)
	return frame
}

// DecodeFrame extracts the HL7 payload from an MLLP-framed message.
// It returns the payload and the number of bytes consumed from the input.
// If a complete frame is not present, it returns io.ErrShortBuffer.
func DecodeFrame(data []byte) ([]byte, int, error) {
	if len(data) < 3 {
		return nil, 0, io.ErrShortBuffer
	}
	if data[0] != StartBlock {
		return nil, 0, fmt.Errorf("missing start block (0x0B)")
	}
	for i := 1; i < len(data)-1; i++ {
		if data[i] == EndBlock && data[i+1] == CarriageReturn {
			payload := data[1:i]
			return payload, i + 2, nil
		}
	}
	return nil, 0, io.ErrShortBuffer
}

// Config holds MLLP listener configuration.
type Config struct {
	Host string
	Port int
}

// DefaultConfig returns the default MLLP configuration.
func DefaultConfig() Config {
	return Config{
		Host: "0.0.0.0",
		Port: 2575,
	}
}

// ConfigFromEnv builds Config from environment variables.
func ConfigFromEnv() Config {
	cfg := DefaultConfig()
	if h := os.Getenv("GHEGA_MLLP_HOST"); h != "" {
		cfg.Host = h
	}
	if p := os.Getenv("GHEGA_MLLP_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			cfg.Port = v
		}
	}
	return cfg
}

// Handler is called for each received HL7 message. It must return the ACK
// payload to send back to the client, or an error.
type Handler func(msg *hl7v2.Message) ([]byte, error)

// DefaultHandler parses the message and returns an AA ACK.
func DefaultHandler(msg *hl7v2.Message) ([]byte, error) {
	return hl7v2.GenerateACK(msg, "AA")
}

// Listener binds to a TCP address, accepts connections, and handles
// MLLP-framed HL7v2 messages.
type Listener struct {
	cfg      Config
	listener net.Listener
	handler  Handler
	logger   *slog.Logger

	mu     sync.Mutex
	wg     sync.WaitGroup
	closed bool
	quit   chan struct{}
}

// NewListener creates a new MLLP Listener with the given configuration.
// If handler is nil, DefaultHandler is used.
func NewListener(cfg Config, handler Handler, logger *slog.Logger) *Listener {
	if handler == nil {
		handler = DefaultHandler
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Listener{
		cfg:     cfg,
		handler: handler,
		logger:  logger,
		quit:    make(chan struct{}),
	}
}

// Addr returns the listener's network address, or nil if not started.
func (l *Listener) Addr() net.Addr {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.listener == nil {
		return nil
	}
	return l.listener.Addr()
}

// Start binds the listener and begins accepting connections.
func (l *Listener) Start() error {
	addr := fmt.Sprintf("%s:%d", l.cfg.Host, l.cfg.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("mllp listen on %s: %w", addr, err)
	}

	l.mu.Lock()
	l.listener = ln
	l.closed = false
	l.mu.Unlock()

	l.logger.Info("ghega mllp listener started", slog.String("addr", ln.Addr().String()))

	go l.acceptLoop()
	return nil
}

// Stop gracefully shuts down the listener.
func (l *Listener) Stop() error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return nil
	}
	l.closed = true
	close(l.quit)
	ln := l.listener
	l.mu.Unlock()

	if ln != nil {
		_ = ln.Close()
	}

	l.wg.Wait()
	l.logger.Info("ghega mllp listener stopped")
	return nil
}

func (l *Listener) acceptLoop() {
	for {
		l.mu.Lock()
		ln := l.listener
		l.mu.Unlock()
		if ln == nil {
			return
		}

		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-l.quit:
				return
			default:
				l.logger.Error("mllp accept error", slog.String("error", err.Error()))
				continue
			}
		}

		l.wg.Add(1)
		go func() {
			defer l.wg.Done()
			l.handleConnection(conn)
		}()
	}
}

func (l *Listener) handleConnection(conn net.Conn) {
	defer conn.Close()

	l.logger.Info("mllp connection established", slog.String("remote", conn.RemoteAddr().String()))
	defer l.logger.Info("mllp connection closed", slog.String("remote", conn.RemoteAddr().String()))

	reader := bufio.NewReader(conn)
	for {
		select {
		case <-l.quit:
			return
		default:
		}

		if err := l.readAndProcess(reader, conn); err != nil {
			if err == io.EOF {
				return
			}
			select {
			case <-l.quit:
				return
			default:
				l.logger.Error("mllp read/process error",
					slog.String("remote", conn.RemoteAddr().String()),
					slog.String("error", err.Error()))
				return
			}
		}
	}
}

func (l *Listener) readAndProcess(reader *bufio.Reader, conn net.Conn) error {
	// Read start block
	b, err := reader.ReadByte()
	if err != nil {
		return err
	}
	if b != StartBlock {
		return fmt.Errorf("expected start block 0x0B, got 0x%02X", b)
	}

	var payload bytes.Buffer
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return err
		}
		if b == EndBlock {
			next, err := reader.ReadByte()
			if err != nil {
				return err
			}
			if next == CarriageReturn {
				break
			}
			payload.WriteByte(b)
			payload.WriteByte(next)
		} else {
			payload.WriteByte(b)
		}
	}

	msg, err := hl7v2.Parse(payload.Bytes())
	if err != nil {
		l.logger.Error("hl7v2 parse error",
			slog.String("remote", conn.RemoteAddr().String()),
			slog.String("error", err.Error()))
		// Send NAK-like response on parse failure
		nak := EncodeFrame([]byte("MSH|^~\\&|||||||ACK|||P|2.5\rMSA|AR|UNKNOWN\r"))
		_ = l.writeWithTimeout(conn, nak)
		return nil
	}

	ackPayload, err := l.handler(msg)
	if err != nil {
		l.logger.Error("mllp handler error",
			slog.String("remote", conn.RemoteAddr().String()),
			slog.String("error", err.Error()))
		// Send AE on handler error
		ackPayload, _ = hl7v2.GenerateACK(msg, "AE")
	}

	frame := EncodeFrame(ackPayload)
	if err := l.writeWithTimeout(conn, frame); err != nil {
		l.logger.Error("mllp write error",
			slog.String("remote", conn.RemoteAddr().String()),
			slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (l *Listener) writeWithTimeout(conn net.Conn, data []byte) error {
	if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	defer conn.SetWriteDeadline(time.Time{}) // clear deadline
	_, err := conn.Write(data)
	return err
}
