package messagestore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sroopra/ghega/pkg/payloadref"
	_ "modernc.org/sqlite"
)

// SQLiteStore persists message metadata in SQLite and raw payloads in a
// separate payloads table. It uses modernc.org/sqlite (pure Go, no CGO).
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens a SQLite database at the given DSN and initializes
// the required tables. A typical DSN is a file path such as "ghega.db".
func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("migrate sqlite: %w", err)
	}
	return store, nil
}

// migrate creates the messages and payloads tables if they do not exist.
func (s *SQLiteStore) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			channel_id TEXT NOT NULL,
			message_id TEXT NOT NULL UNIQUE,
			received_at TEXT NOT NULL,
			status TEXT NOT NULL,
			storage_id TEXT NOT NULL,
			location TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_messages_channel_id ON messages(channel_id);
		CREATE INDEX IF NOT EXISTS idx_messages_message_id ON messages(message_id);

		CREATE TABLE IF NOT EXISTS payloads (
			storage_id TEXT PRIMARY KEY,
			data BLOB NOT NULL
		);
	`)
	return err
}

// Save persists the message metadata and raw payload.
func (s *SQLiteStore) Save(ctx context.Context, envelope *payloadref.Envelope, rawPayload []byte) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO messages (channel_id, message_id, received_at, status, storage_id, location)
		VALUES (?, ?, ?, ?, ?, ?)
	`, envelope.ChannelID, envelope.MessageID, envelope.ReceivedAt.Format("2006-01-02T15:04:05.999999999Z07:00"), envelope.Status, envelope.Ref.StorageID, envelope.Ref.Location)
	if err != nil {
		return fmt.Errorf("insert message metadata: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO payloads (storage_id, data) VALUES (?, ?)
		ON CONFLICT(storage_id) DO UPDATE SET data = excluded.data
	`, envelope.Ref.StorageID, rawPayload)
	if err != nil {
		return fmt.Errorf("insert payload: %w", err)
	}

	return tx.Commit()
}

// GetMetadata retrieves message metadata by message ID.
func (s *SQLiteStore) GetMetadata(ctx context.Context, messageID string) (*payloadref.Envelope, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT channel_id, message_id, received_at, status, storage_id, location
		FROM messages
		WHERE message_id = ?
	`, messageID)

	var env payloadref.Envelope
	var receivedAtStr string
	err := row.Scan(&env.ChannelID, &env.MessageID, &receivedAtStr, &env.Status, &env.Ref.StorageID, &env.Ref.Location)
	if err == sql.ErrNoRows {
		return nil, &ErrNotFound{MessageID: messageID}
	}
	if err != nil {
		return nil, fmt.Errorf("scan message metadata: %w", err)
	}

	env.ReceivedAt, err = parseTime(receivedAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse received_at: %w", err)
	}
	return &env, nil
}

// ListByChannel returns message metadata for a channel, paginated.
func (s *SQLiteStore) ListByChannel(ctx context.Context, channelID string, limit, offset int) ([]*payloadref.Envelope, error) {
	if limit <= 0 {
		limit = -1
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT channel_id, message_id, received_at, status, storage_id, location
		FROM messages
		WHERE channel_id = ?
		ORDER BY received_at DESC
		LIMIT ? OFFSET ?
	`, channelID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	var out []*payloadref.Envelope
	for rows.Next() {
		var env payloadref.Envelope
		var receivedAtStr string
		if err := rows.Scan(&env.ChannelID, &env.MessageID, &receivedAtStr, &env.Status, &env.Ref.StorageID, &env.Ref.Location); err != nil {
			return nil, fmt.Errorf("scan message row: %w", err)
		}
		env.ReceivedAt, err = parseTime(receivedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse received_at: %w", err)
		}
		out = append(out, &env)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}
	return out, nil
}

// GetPayload retrieves raw payload bytes by storage ID.
// This is intended for testing and internal use only.
func (s *SQLiteStore) GetPayload(ctx context.Context, storageID string) ([]byte, bool, error) {
	row := s.db.QueryRowContext(ctx, `SELECT data FROM payloads WHERE storage_id = ?`, storageID)
	var data []byte
	err := row.Scan(&data)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("scan payload: %w", err)
	}
	return data, true, nil
}

// ListAll returns message metadata across all channels, paginated.
func (s *SQLiteStore) ListAll(ctx context.Context, limit, offset int) ([]*payloadref.Envelope, error) {
	if limit <= 0 {
		limit = -1
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT channel_id, message_id, received_at, status, storage_id, location
		FROM messages
		ORDER BY received_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	var out []*payloadref.Envelope
	for rows.Next() {
		var env payloadref.Envelope
		var receivedAtStr string
		if err := rows.Scan(&env.ChannelID, &env.MessageID, &receivedAtStr, &env.Status, &env.Ref.StorageID, &env.Ref.Location); err != nil {
			return nil, fmt.Errorf("scan message row: %w", err)
		}
		env.ReceivedAt, err = parseTime(receivedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse received_at: %w", err)
		}
		out = append(out, &env)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}
	return out, nil
}

// UpdateStatus updates the status of a message by its message ID.
func (s *SQLiteStore) UpdateStatus(ctx context.Context, messageID, status string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE messages SET status = ? WHERE message_id = ?
	`, status, messageID)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

// Delete removes a message and its payload by message ID.
func (s *SQLiteStore) Delete(ctx context.Context, messageID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var storageID string
	err = tx.QueryRowContext(ctx, `SELECT storage_id FROM messages WHERE message_id = ?`, messageID).Scan(&storageID)
	if err == sql.ErrNoRows {
		return &ErrNotFound{MessageID: messageID}
	}
	if err != nil {
		return fmt.Errorf("select message: %w", err)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM messages WHERE message_id = ?`, messageID)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM payloads WHERE storage_id = ?`, storageID)
	if err != nil {
		return fmt.Errorf("delete payload: %w", err)
	}
	return tx.Commit()
}

// Close closes the underlying database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}
