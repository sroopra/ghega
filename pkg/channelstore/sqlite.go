package channelstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore persists channel revisions and deployment audits in SQLite.
// It uses modernc.org/sqlite (pure Go, no CGO).
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

func (s *SQLiteStore) migrate() error {
	// Migrate old schema (PRIMARY KEY (name, hash)) to new schema
	// (surrogate auto-increment key) so that rollback can insert a
	// new row with the same hash.
	var tableExists int
	if err := s.db.QueryRow(`
		SELECT count(*) FROM sqlite_master WHERE type='table' AND name='channels'
	`).Scan(&tableExists); err != nil {
		return fmt.Errorf("check channels table: %w", err)
	}

	if tableExists > 0 {
		var idColExists int
		if err := s.db.QueryRow(`
			SELECT count(*) FROM pragma_table_info('channels') WHERE name='id'
		`).Scan(&idColExists); err != nil {
			return fmt.Errorf("check id column: %w", err)
		}
		if idColExists == 0 {
			_, err := s.db.Exec(`
				ALTER TABLE channels RENAME TO channels_old;
				CREATE TABLE channels (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL,
					hash TEXT NOT NULL,
					yaml BLOB NOT NULL,
					revision INTEGER NOT NULL,
					deployed_at TEXT NOT NULL
				);
				CREATE INDEX idx_channels_name ON channels(name);
				INSERT INTO channels (name, hash, yaml, revision, deployed_at)
				SELECT name, hash, yaml, revision, deployed_at FROM channels_old;
				DROP TABLE channels_old;
			`)
			if err != nil {
				return fmt.Errorf("migrate channels table: %w", err)
			}
		}
	} else {
		_, err := s.db.Exec(`
			CREATE TABLE channels (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				hash TEXT NOT NULL,
				yaml BLOB NOT NULL,
				revision INTEGER NOT NULL,
				deployed_at TEXT NOT NULL
			);
			CREATE INDEX idx_channels_name ON channels(name);
		`)
		if err != nil {
			return fmt.Errorf("create channels table: %w", err)
		}
	}

	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS deployments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			channel_name TEXT NOT NULL,
			hash TEXT NOT NULL,
			action TEXT NOT NULL,
			timestamp TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_deployments_channel ON deployments(channel_name);
	`)
	return err
}

// SaveChannel persists a channel revision. If revision <= 0 it is auto-incremented
// as max(existing) + 1.
func (s *SQLiteStore) SaveChannel(ctx context.Context, name, hash string, yamlBytes []byte, revision int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if revision <= 0 {
		var maxRev sql.NullInt64
		row := tx.QueryRowContext(ctx, `SELECT MAX(revision) FROM channels WHERE name = ?`, name)
		if err := row.Scan(&maxRev); err != nil {
			return fmt.Errorf("query max revision: %w", err)
		}
		if maxRev.Valid {
			revision = int(maxRev.Int64) + 1
		} else {
			revision = 1
		}
	}

	deployedAt := time.Now().UTC()
	res, err := tx.ExecContext(ctx, `
		UPDATE channels SET yaml = ?, revision = ?, deployed_at = ?
		WHERE name = ? AND hash = ?
	`, yamlBytes, revision, deployedAt.Format(time.RFC3339Nano), name, hash)
	if err != nil {
		return fmt.Errorf("update channel: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO channels (name, hash, yaml, revision, deployed_at)
			VALUES (?, ?, ?, ?, ?)
		`, name, hash, yamlBytes, revision, deployedAt.Format(time.RFC3339Nano))
		if err != nil {
			return fmt.Errorf("insert channel: %w", err)
		}
	}

	return tx.Commit()
}

// GetChannel returns the latest revision for a channel name.
func (s *SQLiteStore) GetChannel(ctx context.Context, name string) (*ChannelRecord, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT name, hash, yaml, revision, deployed_at
		FROM channels
		WHERE name = ?
		ORDER BY revision DESC
		LIMIT 1
	`, name)

	var rec ChannelRecord
	var deployedAtStr string
	err := row.Scan(&rec.Name, &rec.Hash, &rec.YAML, &rec.Revision, &deployedAtStr)
	if err == sql.ErrNoRows {
		return nil, &ErrNotFound{Name: name}
	}
	if err != nil {
		return nil, fmt.Errorf("scan channel: %w", err)
	}
	rec.DeployedAt, err = time.Parse(time.RFC3339Nano, deployedAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse deployed_at: %w", err)
	}
	return &rec, nil
}

// GetChannelRevision returns a specific revision by hash.
// If multiple rows share the same hash, the latest revision is returned.
func (s *SQLiteStore) GetChannelRevision(ctx context.Context, name, hash string) (*ChannelRecord, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT name, hash, yaml, revision, deployed_at
		FROM channels
		WHERE name = ? AND hash = ?
		ORDER BY revision DESC
		LIMIT 1
	`, name, hash)

	var rec ChannelRecord
	var deployedAtStr string
	err := row.Scan(&rec.Name, &rec.Hash, &rec.YAML, &rec.Revision, &deployedAtStr)
	if err == sql.ErrNoRows {
		return nil, &ErrNotFound{Name: name, Hash: hash}
	}
	if err != nil {
		return nil, fmt.Errorf("scan channel revision: %w", err)
	}
	rec.DeployedAt, err = time.Parse(time.RFC3339Nano, deployedAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse deployed_at: %w", err)
	}
	return &rec, nil
}

// ListChannelRevisions returns all revisions for a channel ordered by revision desc.
func (s *SQLiteStore) ListChannelRevisions(ctx context.Context, name string) ([]ChannelRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT name, hash, yaml, revision, deployed_at
		FROM channels
		WHERE name = ?
		ORDER BY revision DESC
	`, name)
	if err != nil {
		return nil, fmt.Errorf("query revisions: %w", err)
	}
	defer rows.Close()

	var out []ChannelRecord
	for rows.Next() {
		var rec ChannelRecord
		var deployedAtStr string
		if err := rows.Scan(&rec.Name, &rec.Hash, &rec.YAML, &rec.Revision, &deployedAtStr); err != nil {
			return nil, fmt.Errorf("scan revision row: %w", err)
		}
		rec.DeployedAt, err = time.Parse(time.RFC3339Nano, deployedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse deployed_at: %w", err)
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}
	return out, nil
}

// RollbackChannel verifies the hash exists, creates a new current revision
// pointing to the rolled-back hash, and records a rollback audit entry.
func (s *SQLiteStore) RollbackChannel(ctx context.Context, name, hash string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var yamlBytes []byte
	err = tx.QueryRowContext(ctx, `
		SELECT yaml FROM channels WHERE name = ? AND hash = ?
		ORDER BY revision DESC LIMIT 1
	`, name, hash).Scan(&yamlBytes)
	if err == sql.ErrNoRows {
		return &ErrNotFound{Name: name, Hash: hash}
	}
	if err != nil {
		return fmt.Errorf("verify hash exists: %w", err)
	}

	var maxRev sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MAX(revision) FROM channels WHERE name = ?`, name).Scan(&maxRev); err != nil {
		return fmt.Errorf("query max revision: %w", err)
	}
	newRev := 1
	if maxRev.Valid {
		newRev = int(maxRev.Int64) + 1
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO channels (name, hash, yaml, revision, deployed_at)
		VALUES (?, ?, ?, ?, ?)
	`, name, hash, yamlBytes, newRev, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("insert rolled-back revision: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO deployments (channel_name, hash, action, timestamp)
		VALUES (?, ?, ?, ?)
	`, name, hash, "rollback", time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}

	return tx.Commit()
}

// SaveDeploymentAudit records a deployment action.
func (s *SQLiteStore) SaveDeploymentAudit(ctx context.Context, channelName, hash, action string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO deployments (channel_name, hash, action, timestamp)
		VALUES (?, ?, ?, ?)
	`, channelName, hash, action, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}
	return nil
}

// ListDeploymentAudit returns audit entries for a channel ordered by timestamp desc.
func (s *SQLiteStore) ListDeploymentAudit(ctx context.Context, channelName string) ([]AuditRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT channel_name, hash, action, timestamp
		FROM deployments
		WHERE channel_name = ?
		ORDER BY timestamp DESC
	`, channelName)
	if err != nil {
		return nil, fmt.Errorf("query audits: %w", err)
	}
	defer rows.Close()

	var out []AuditRecord
	for rows.Next() {
		var rec AuditRecord
		var tsStr string
		if err := rows.Scan(&rec.ChannelName, &rec.Hash, &rec.Action, &tsStr); err != nil {
			return nil, fmt.Errorf("scan audit row: %w", err)
		}
		rec.Timestamp, err = time.Parse(time.RFC3339Nano, tsStr)
		if err != nil {
			return nil, fmt.Errorf("parse timestamp: %w", err)
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}
	return out, nil
}

// Close closes the underlying database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
