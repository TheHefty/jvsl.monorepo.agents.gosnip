package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/TheHefty/jvsl.monorepo.agents.gosnip/gosnip/internal/snippet"
	modernsqlite "modernc.org/sqlite"
)

const (
	currentSchemaVersion = 2
	defaultBusyTimeout   = 5 * time.Second
)

type Options struct {
	Now func() time.Time

	busyTimeout        time.Duration
	afterMigrationStep func(version int) error
	backup             func(context.Context, string, time.Time, time.Duration) (string, error)
}

type OpenReport struct {
	Created         bool
	Migrated        bool
	PreviousVersion int
	CurrentVersion  int
	BackupPath      string
}

type Store struct {
	db  *sql.DB
	now func() time.Time
}

func Open(ctx context.Context, path string, options Options) (_ *Store, report OpenReport, finalErr error) {
	if path == "" {
		return nil, report, fmt.Errorf("open database: path is empty")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, report, fmt.Errorf("open database: resolve %s: %w", path, err)
	}
	path = absPath

	created, err := prepareDatabasePath(path)
	if err != nil {
		return nil, report, err
	}
	report.Created = created

	busyTimeout := options.busyTimeout
	if busyTimeout <= 0 {
		busyTimeout = defaultBusyTimeout
	}
	db, err := sql.Open("sqlite", databaseDSN(path, busyTimeout))
	if err != nil {
		cleanupNewDatabase(path, created)
		return nil, report, fmt.Errorf("open database %s: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	keepOpen := false
	defer func() {
		if !keepOpen {
			_ = db.Close()
			cleanupNewDatabase(path, created && finalErr != nil)
		}
	}()

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, report, classifySQLiteError(fmt.Errorf("connect to database %s: %w", path, err))
	}
	defer conn.Close()

	if err := checkIntegrity(ctx, conn, path); err != nil {
		return nil, report, err
	}
	version, err := schemaVersion(ctx, conn, path, created)
	if err != nil {
		return nil, report, err
	}
	report.PreviousVersion = version
	if version > currentSchemaVersion {
		return nil, report, &FutureSchemaError{Path: path, Found: version, Current: currentSchemaVersion}
	}

	validatedSchema := false
	if version < currentSchemaVersion {
		if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
			return nil, report, classifySQLiteError(fmt.Errorf("lock database for migration: %w", err))
		}
		committed := false
		defer func() {
			if !committed {
				_, _ = conn.ExecContext(context.Background(), "ROLLBACK")
			}
		}()

		lockedVersion, err := schemaVersion(ctx, conn, path, created)
		if err != nil {
			return nil, report, err
		}
		version = lockedVersion
		report.PreviousVersion = version
		if version > currentSchemaVersion {
			return nil, report, &FutureSchemaError{Path: path, Found: version, Current: currentSchemaVersion}
		}
		if !created && version < currentSchemaVersion {
			backup := options.backup
			if backup == nil {
				backup = createBackup
			}
			report.BackupPath, err = backup(ctx, path, now(options), busyTimeout)
			if err != nil {
				return nil, report, fmt.Errorf("back up database before migration: %w", err)
			}
		}
		if err := applyMigrations(ctx, conn, version, options, now(options)); err != nil {
			return nil, report, err
		}
		if err := validateCurrentSchema(ctx, conn, path); err != nil {
			return nil, report, err
		}
		validatedSchema = true
		if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
			return nil, report, classifySQLiteError(fmt.Errorf("commit database migration: %w", err))
		}
		committed = true
		report.Migrated = !created
	}
	if !validatedSchema {
		if err := validateCurrentSchema(ctx, conn, path); err != nil {
			return nil, report, err
		}
	}

	var journalMode string
	if err := conn.QueryRowContext(ctx, "PRAGMA journal_mode=WAL").Scan(&journalMode); err != nil {
		return nil, report, classifySQLiteError(fmt.Errorf("enable WAL for database %s: %w", path, err))
	}
	if !strings.EqualFold(journalMode, "wal") {
		return nil, report, fmt.Errorf("enable WAL for database %s: SQLite selected %q", path, journalMode)
	}
	if err := secureSQLiteFiles(path); err != nil {
		return nil, report, err
	}

	report.CurrentVersion = currentSchemaVersion
	clock := options.Now
	if clock == nil {
		clock = time.Now
	}
	keepOpen = true
	return &Store{db: db, now: clock}, report, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Create(ctx context.Context, input snippet.NewSnippet) (snippet.Snippet, error) {
	timestamp := s.now().UTC()
	encodedTime := timestamp.Format(time.RFC3339Nano)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return snippet.Snippet{}, classifySQLiteError(fmt.Errorf("begin snippet creation: %w", err))
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
        INSERT INTO snippets(name, name_key, code, language, tags, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		input.Name(), input.NameKey(), input.Code(), input.Language(), strings.Join(input.Tags(), ","), encodedTime, encodedTime)
	if err != nil {
		return snippet.Snippet{}, classifyCreateError(input.Name(), err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return snippet.Snippet{}, fmt.Errorf("read new snippet ID: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return snippet.Snippet{}, classifySQLiteError(fmt.Errorf("commit snippet creation: %w", err))
	}
	return snippet.Snippet{
		ID: id, Name: input.Name(), NameKey: input.NameKey(), Code: input.Code(),
		Language: input.Language(), Tags: input.Tags(), CreatedAt: timestamp, UpdatedAt: timestamp,
	}, nil
}

func now(options Options) time.Time {
	if options.Now != nil {
		return options.Now().UTC()
	}
	return time.Now().UTC()
}

func databaseDSN(path string, busyTimeout time.Duration) string {
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(path)}
	query := u.Query()
	query.Set("_busy_timeout", strconv.FormatInt(busyTimeout.Milliseconds(), 10))
	query.Set("_foreign_keys", "on")
	u.RawQuery = query.Encode()
	return u.String()
}

func classifyCreateError(name string, err error) error {
	var sqliteErr *modernsqlite.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code()&0xff == 19 {
		return &ConflictError{Name: name, Cause: err}
	}
	return classifySQLiteError(fmt.Errorf("create snippet: %w", err))
}

func classifySQLiteError(err error) error {
	var sqliteErr *modernsqlite.Error
	if errors.As(err, &sqliteErr) {
		switch sqliteErr.Code() & 0xff {
		case 5, 6:
			return &LockedError{Cause: err}
		}
	}
	return err
}

func cleanupNewDatabase(path string, remove bool) {
	if !remove {
		return
	}
	_ = os.Remove(path + "-wal")
	_ = os.Remove(path + "-shm")
	_ = os.Remove(path)
}
