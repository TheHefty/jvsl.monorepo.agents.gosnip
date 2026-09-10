package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	modernsqlite "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

var migrationNames = []string{
	"migrations/001_create_snippets.sql",
	"migrations/002_unique_name_key.sql",
}

func checkIntegrity(ctx context.Context, conn *sql.Conn, path string) error {
	rows, err := conn.QueryContext(ctx, "PRAGMA quick_check")
	if err != nil {
		return &CorruptError{Path: path, Cause: err}
	}
	defer rows.Close()
	result := ""
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return &CorruptError{Path: path, Cause: err}
		}
		if line != "ok" {
			result = line
			break
		}
	}
	if err := rows.Err(); err != nil {
		return &CorruptError{Path: path, Cause: err}
	}
	if result != "" {
		return &CorruptError{Path: path, Cause: errors.New(result)}
	}
	return nil
}

func schemaVersion(ctx context.Context, conn *sql.Conn, path string, created bool) (int, error) {
	var migrationTableCount int
	if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'`).Scan(&migrationTableCount); err != nil {
		return 0, &CorruptError{Path: path, Cause: err}
	}
	if migrationTableCount == 0 {
		var userTableCount int
		if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`).Scan(&userTableCount); err != nil {
			return 0, &CorruptError{Path: path, Cause: err}
		}
		if created && userTableCount == 0 {
			return 0, nil
		}
		return 0, &UnsupportedSchemaError{Path: path}
	}
	var count int
	var minimum, maximum sql.NullInt64
	if err := conn.QueryRowContext(ctx, `SELECT count(*), min(version), max(version) FROM schema_migrations`).Scan(&count, &minimum, &maximum); err != nil {
		return 0, &CorruptError{Path: path, Cause: err}
	}
	if maximum.Valid && maximum.Int64 > currentSchemaVersion {
		return int(maximum.Int64), nil
	}
	if !minimum.Valid || !maximum.Valid || minimum.Int64 != 1 || maximum.Int64 < 1 || count != int(maximum.Int64) {
		return 0, &UnsupportedSchemaError{Path: path}
	}
	return int(maximum.Int64), nil
}

func validateCurrentSchema(ctx context.Context, conn *sql.Conn, path string) error {
	rows, err := conn.QueryContext(ctx, `PRAGMA table_info(snippets)`)
	if err != nil {
		return &CorruptError{Path: path, Cause: err}
	}
	defer rows.Close()
	wantColumns := []string{"id", "name", "name_key", "code", "language", "tags", "created_at", "updated_at"}
	gotColumns := make([]string, 0, len(wantColumns))
	for rows.Next() {
		var position, notNull, primaryKey int
		var name, dataType string
		var defaultValue any
		if err := rows.Scan(&position, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
			return &CorruptError{Path: path, Cause: err}
		}
		gotColumns = append(gotColumns, name)
	}
	if err := rows.Err(); err != nil {
		return &CorruptError{Path: path, Cause: err}
	}
	if len(gotColumns) != len(wantColumns) {
		return &CorruptError{Path: path, Cause: fmt.Errorf("snippets table has columns %v, want %v", gotColumns, wantColumns)}
	}
	for index := range wantColumns {
		if gotColumns[index] != wantColumns[index] {
			return &CorruptError{Path: path, Cause: fmt.Errorf("snippets table has columns %v, want %v", gotColumns, wantColumns)}
		}
	}
	var indexCount int
	if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE type='index' AND name='snippets_name_key_unique'`).Scan(&indexCount); err != nil {
		return &CorruptError{Path: path, Cause: err}
	}
	if indexCount != 1 {
		return &CorruptError{Path: path, Cause: errors.New("schema v2 is missing the unique name-key index")}
	}
	return nil
}

func applyMigrations(ctx context.Context, conn *sql.Conn, from int, options Options, appliedAt time.Time) error {
	for version := from + 1; version <= currentSchemaVersion; version++ {
		contents, err := migrationFiles.ReadFile(migrationNames[version-1])
		if err != nil {
			return fmt.Errorf("read schema migration %d: %w", version, err)
		}
		if _, err := conn.ExecContext(ctx, string(contents)); err != nil {
			return classifySQLiteError(fmt.Errorf("apply schema migration %d: %w", version, err))
		}
		if options.afterMigrationStep != nil {
			if err := options.afterMigrationStep(version); err != nil {
				return fmt.Errorf("apply schema migration %d: %w", version, err)
			}
		}
		if _, err := conn.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES (?, ?)`, version, appliedAt.UTC().Format(time.RFC3339Nano)); err != nil {
			return classifySQLiteError(fmt.Errorf("record schema migration %d: %w", version, err))
		}
	}
	return nil
}

func createBackup(ctx context.Context, path string, timestamp time.Time, busyTimeout time.Duration) (string, error) {
	base := path + ".backup-" + timestamp.UTC().Format("20060102T150405.000000000Z")
	backupPath := ""
	for attempt := 0; ; attempt++ {
		candidate := base
		if attempt > 0 {
			candidate += "-" + strconv.Itoa(attempt)
		}
		file, err := os.OpenFile(candidate, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		if err := file.Close(); err != nil {
			_ = os.Remove(candidate)
			return "", err
		}
		backupPath = candidate
		break
	}

	complete := false
	defer func() {
		if !complete {
			_ = os.Remove(backupPath)
		}
	}()
	sourceDB, err := sql.Open("sqlite", databaseDSN(path, busyTimeout))
	if err != nil {
		return "", err
	}
	sourceDB.SetMaxOpenConns(1)
	defer sourceDB.Close()
	sourceConn, err := sourceDB.Conn(ctx)
	if err != nil {
		return "", err
	}
	defer sourceConn.Close()

	err = sourceConn.Raw(func(driverConn any) error {
		backuper, ok := driverConn.(interface {
			NewBackup(string) (*modernsqlite.Backup, error)
		})
		if !ok {
			return fmt.Errorf("SQLite driver does not support online backup")
		}
		backup, err := backuper.NewBackup(backupPath)
		if err != nil {
			return err
		}
		_, stepErr := backup.Step(-1)
		finishErr := backup.Finish()
		if stepErr != nil {
			return stepErr
		}
		return finishErr
	})
	if err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(backupPath, 0o600); err != nil {
			return "", err
		}
	}
	complete = true
	return backupPath, nil
}
