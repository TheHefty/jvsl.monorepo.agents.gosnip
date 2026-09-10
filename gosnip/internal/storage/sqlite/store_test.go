package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/TheHefty/jvsl.monorepo.agents.gosnip/gosnip/internal/snippet"
	_ "modernc.org/sqlite"
)

func TestOpenCreatesCurrentPrivateDatabaseAndCreatePreservesValues(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "gosnip")
	path := filepath.Join(dir, "gosnip.db")
	now := time.Date(2026, 9, 10, 7, 30, 1, 123456789, time.FixedZone("local", 3600))
	store, report, err := Open(context.Background(), path, Options{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	if !report.Created || report.CurrentVersion != currentSchemaVersion || report.BackupPath != "" {
		t.Errorf("OpenReport = %+v", report)
	}
	if runtime.GOOS != "windows" {
		assertMode(t, dir, 0o700)
		assertMode(t, path, 0o600)
	}

	input := mustNormalize(t, " Example ", []byte("line one\r\nline two\n"))
	created, err := store.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID <= 0 || created.Name != "Example" || created.NameKey != "example" {
		t.Errorf("created identity = %+v", created)
	}
	if !bytes.Equal(created.Code, input.Code()) {
		t.Errorf("created code = %q, want %q", created.Code, input.Code())
	}
	if !created.CreatedAt.Equal(now.UTC()) || !created.UpdatedAt.Equal(now.UTC()) {
		t.Errorf("created timestamps = %s / %s", created.CreatedAt, created.UpdatedAt)
	}
	var storedCode []byte
	var storedTags, storedCreated, storedUpdated string
	if err := store.db.QueryRow(`SELECT code, tags, created_at, updated_at FROM snippets WHERE id = ?`, created.ID).
		Scan(&storedCode, &storedTags, &storedCreated, &storedUpdated); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(storedCode, input.Code()) || storedTags != "cli" {
		t.Errorf("stored code/tags = %q / %q", storedCode, storedTags)
	}
	if storedCreated != "2026-09-10T06:30:01.123456789Z" || storedUpdated != storedCreated {
		t.Errorf("stored timestamps = %q / %q", storedCreated, storedUpdated)
	}
}

func TestOpenMigratesV1OnceAndAvoidsBackupNameCollisions(t *testing.T) {
	t.Parallel()

	path := filepath.Join(privateTempDir(t), "gosnip.db")
	createV1Fixture(t, path, "Existing", "existing")
	now := time.Date(2026, 9, 10, 8, 15, 0, 7, time.UTC)
	collision := path + ".backup-20260910T081500.000000007Z"
	if err := os.WriteFile(collision, []byte("occupied"), 0o600); err != nil {
		t.Fatal(err)
	}

	store, report, err := Open(context.Background(), path, Options{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if !report.Migrated || report.PreviousVersion != 1 || report.CurrentVersion != 2 {
		t.Errorf("OpenReport = %+v", report)
	}
	if report.BackupPath != collision+"-1" {
		t.Errorf("backup path = %q, want collision suffix", report.BackupPath)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	assertSchemaAndSnippet(t, path, 2, "Existing")
	assertSchemaAndSnippet(t, report.BackupPath, 1, "Existing")

	store, secondReport, err := Open(context.Background(), path, Options{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if secondReport.Migrated || secondReport.BackupPath != "" || secondReport.PreviousVersion != 2 {
		t.Errorf("second OpenReport = %+v", secondReport)
	}
}

func TestOpenPreservesFutureAndCorruptDatabases(t *testing.T) {
	t.Parallel()

	t.Run("future schema", func(t *testing.T) {
		path := filepath.Join(privateTempDir(t), "future.db")
		createV1Fixture(t, path, "Existing", "existing")
		db, err := sql.Open("sqlite", path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES (99, '2026-09-10T08:00:00Z')`); err != nil {
			t.Fatal(err)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		before, _ := os.ReadFile(path)

		_, _, err = Open(context.Background(), path, Options{})
		var futureErr *FutureSchemaError
		if !errors.As(err, &futureErr) {
			t.Fatalf("Open() error = %v, want FutureSchemaError", err)
		}
		after, _ := os.ReadFile(path)
		if !bytes.Equal(before, after) {
			t.Error("future database bytes changed")
		}
	})

	t.Run("corrupt database", func(t *testing.T) {
		path := filepath.Join(privateTempDir(t), "corrupt.db")
		original := []byte("not a SQLite database")
		if err := os.WriteFile(path, original, 0o600); err != nil {
			t.Fatal(err)
		}
		_, _, err := Open(context.Background(), path, Options{})
		var corruptErr *CorruptError
		if !errors.As(err, &corruptErr) {
			t.Fatalf("Open() error = %v, want CorruptError", err)
		}
		after, _ := os.ReadFile(path)
		if !bytes.Equal(original, after) {
			t.Error("corrupt database bytes changed")
		}
	})

	t.Run("incomplete current schema", func(t *testing.T) {
		path := filepath.Join(privateTempDir(t), "incomplete.db")
		db, err := sql.Open("sqlite", path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL); INSERT INTO schema_migrations VALUES (1, '2026-09-10T08:00:00Z'), (2, '2026-09-10T08:00:00Z')`); err != nil {
			t.Fatal(err)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		if runtime.GOOS != "windows" {
			if err := os.Chmod(path, 0o600); err != nil {
				t.Fatal(err)
			}
		}

		_, _, err = Open(context.Background(), path, Options{})
		var corruptErr *CorruptError
		if !errors.As(err, &corruptErr) {
			t.Fatalf("Open() error = %v, want CorruptError", err)
		}
	})
}

func TestMigrationFailureRollsBackAndKeepsReadableBackup(t *testing.T) {
	t.Parallel()

	path := filepath.Join(privateTempDir(t), "gosnip.db")
	createV1Fixture(t, path, "Existing", "existing")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	injected := errors.New("injected migration failure")
	_, report, err := Open(context.Background(), path, Options{
		Now: func() time.Time { return time.Date(2026, 9, 10, 8, 0, 0, 42, time.UTC) },
		afterMigrationStep: func(version int) error {
			if version == 2 {
				return injected
			}
			return nil
		},
	})
	if !errors.Is(err, injected) {
		t.Fatalf("Open() error = %v, want injected failure", err)
	}
	if report.BackupPath == "" {
		t.Fatal("OpenReport.BackupPath is empty after completed backup")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, before) {
		t.Error("failed migration changed the main database bytes")
	}
	assertSchemaAndSnippet(t, path, 1, "Existing")
	assertSchemaAndSnippet(t, report.BackupPath, 1, "Existing")
}

func TestBackupFailurePreventsMigration(t *testing.T) {
	t.Parallel()

	path := filepath.Join(privateTempDir(t), "gosnip.db")
	createV1Fixture(t, path, "Existing", "existing")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	injected := errors.New("injected backup failure")
	_, report, err := Open(context.Background(), path, Options{
		backup: func(context.Context, string, time.Time, time.Duration) (string, error) {
			return "", injected
		},
	})
	if !errors.Is(err, injected) {
		t.Fatalf("Open() error = %v, want injected failure", err)
	}
	if report.BackupPath != "" || report.CurrentVersion != 0 {
		t.Errorf("OpenReport = %+v", report)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("backup failure changed the main database bytes")
	}
	assertSchemaAndSnippet(t, path, 1, "Existing")
}

func TestOpenRejectsUnsafePathsWithoutChangingThem(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX mode bits are not the Windows ACL model")
	}

	t.Run("unsafe database mode", func(t *testing.T) {
		dir := privateTempDir(t)
		path := filepath.Join(dir, "gosnip.db")
		createV1Fixture(t, path, "Existing", "existing")
		if err := os.Chmod(path, 0o644); err != nil {
			t.Fatal(err)
		}
		before, _ := os.ReadFile(path)

		_, _, err := Open(context.Background(), path, Options{})
		var permissionErr *PermissionError
		if !errors.As(err, &permissionErr) {
			t.Fatalf("Open() error = %v, want PermissionError", err)
		}
		after, _ := os.ReadFile(path)
		if !bytes.Equal(after, before) {
			t.Error("unsafe database was changed")
		}
		assertMode(t, path, 0o644)
	})

	t.Run("database symlink", func(t *testing.T) {
		dir := privateTempDir(t)
		target := filepath.Join(dir, "target.db")
		createV1Fixture(t, target, "Existing", "existing")
		link := filepath.Join(dir, "gosnip.db")
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}

		_, _, err := Open(context.Background(), link, Options{})
		var pathErr *UnsafePathError
		if !errors.As(err, &pathErr) {
			t.Fatalf("Open() error = %v, want UnsafePathError", err)
		}
		assertSchemaAndSnippet(t, target, 1, "Existing")
	})

	t.Run("unsafe data directory", func(t *testing.T) {
		dir := privateTempDir(t)
		path := filepath.Join(dir, "gosnip.db")
		createV1Fixture(t, path, "Existing", "existing")
		if err := os.Chmod(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		_, _, err := Open(context.Background(), path, Options{})
		var permissionErr *PermissionError
		if !errors.As(err, &permissionErr) || permissionErr.Path != dir {
			t.Fatalf("Open() error = %v, want directory PermissionError", err)
		}
		assertMode(t, dir, 0o755)
	})

	t.Run("unsafe WAL sidecar", func(t *testing.T) {
		dir := privateTempDir(t)
		path := filepath.Join(dir, "gosnip.db")
		createV1Fixture(t, path, "Existing", "existing")
		walPath := path + "-wal"
		if err := os.WriteFile(walPath, []byte("unsafe sidecar"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, _, err := Open(context.Background(), path, Options{})
		var permissionErr *PermissionError
		if !errors.As(err, &permissionErr) || permissionErr.Path != walPath {
			t.Fatalf("Open() error = %v, want WAL PermissionError", err)
		}
		assertMode(t, walPath, 0o644)
	})
}

func TestCompetingWriterTimesOutAndEquivalentNamesStayUnique(t *testing.T) {
	path := filepath.Join(privateTempDir(t), "gosnip.db")
	first, _, err := Open(context.Background(), path, Options{busyTimeout: 75 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = first.Close() })
	second, _, err := Open(context.Background(), path, Options{busyTimeout: 75 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Close() })

	locker, err := first.db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer locker.Close()
	if _, err := locker.ExecContext(context.Background(), "BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	_, err = second.Create(context.Background(), mustNormalize(t, "Locked", []byte("code")))
	if _, rollbackErr := locker.ExecContext(context.Background(), "ROLLBACK"); rollbackErr != nil {
		t.Fatal(rollbackErr)
	}
	if err := locker.Close(); err != nil {
		t.Fatal(err)
	}
	var lockedErr *LockedError
	if !errors.As(err, &lockedErr) {
		t.Fatalf("Create() error = %v, want LockedError", err)
	}

	inputs := []snippet.NewSnippet{
		mustNormalize(t, "Straße", []byte("first")),
		mustNormalize(t, "STRASSE", []byte("second")),
	}
	errs := make(chan error, len(inputs))
	var wg sync.WaitGroup
	for i, store := range []*Store{first, second} {
		wg.Add(1)
		go func(store *Store, input snippet.NewSnippet) {
			defer wg.Done()
			_, createErr := store.Create(context.Background(), input)
			errs <- createErr
		}(store, inputs[i])
	}
	wg.Wait()
	close(errs)

	successes, conflicts := 0, 0
	for createErr := range errs {
		if createErr == nil {
			successes++
			continue
		}
		var conflictErr *ConflictError
		if errors.As(createErr, &conflictErr) {
			conflicts++
			continue
		}
		t.Fatalf("unexpected concurrent Create() error = %v", createErr)
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent results: successes=%d conflicts=%d", successes, conflicts)
	}
}

func mustNormalize(t *testing.T, name string, code []byte) snippet.NewSnippet {
	t.Helper()
	value, err := snippet.Normalize(snippet.Draft{Name: name, Code: code, Language: "Go", Tags: []string{"CLI"}})
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func createV1Fixture(t *testing.T, path, name, nameKey string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	statements := []string{
		`PRAGMA journal_mode=WAL`,
		`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`,
		`INSERT INTO schema_migrations(version, applied_at) VALUES (1, '2026-09-10T07:00:00Z')`,
		`CREATE TABLE snippets (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, name_key TEXT NOT NULL, code BLOB NOT NULL, language TEXT NOT NULL, tags TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`INSERT INTO snippets(name, name_key, code, language, tags, created_at, updated_at) VALUES (?, ?, ?, 'go', 'fixture', '2026-09-10T07:00:00Z', '2026-09-10T07:00:00Z')`,
	}
	for i, statement := range statements {
		var execErr error
		if i == len(statements)-1 {
			_, execErr = db.Exec(statement, name, nameKey, []byte("fixture\n"))
		} else {
			_, execErr = db.Exec(statement)
		}
		if execErr != nil {
			t.Fatalf("fixture statement %d: %v", i, execErr)
		}
	}
	if _, err := db.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(path, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func assertSchemaAndSnippet(t *testing.T, path string, wantVersion int, wantName string) {
	t.Helper()
	db, err := sql.Open("sqlite", readOnlyDSN(path))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var version int
	if err := db.QueryRow(`SELECT max(version) FROM schema_migrations`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != wantVersion {
		t.Errorf("schema version = %d, want %d", version, wantVersion)
	}
	var name string
	if err := db.QueryRow(`SELECT name FROM snippets`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != wantName {
		t.Errorf("snippet name = %q, want %q", name, wantName)
	}
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Errorf("%s mode = %#o, want %#o", path, got, want)
	}
}

func privateTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}
