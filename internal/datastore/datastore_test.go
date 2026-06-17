package datastore

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "turso.tech/database/tursogo"
)

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		token    string
		expected string
	}{
		{
			name:     "local file without scheme",
			url:      "/path/to/db.sqlite",
			token:    "",
			expected: "/path/to/db.sqlite",
		},
		{
			name:     "local file with file scheme",
			url:      "file:/path/to/db.sqlite",
			token:    "",
			expected: "file:/path/to/db.sqlite",
		},
		{
			name:     "libsql url without token",
			url:      "libsql://my-db.turso.io",
			token:    "",
			expected: "libsql://my-db.turso.io",
		},
		{
			name:     "libsql url with token",
			url:      "libsql://my-db.turso.io",
			token:    "test-token",
			expected: "libsql://my-db.turso.io?authToken=test-token",
		},
		{
			name:     "libsql url with existing query params",
			url:      "libsql://my-db.turso.io?foo=bar",
			token:    "test-token",
			expected: "libsql://my-db.turso.io?foo=bar&authToken=test-token",
		},
		{
			name:     "https url with token",
			url:      "https://my-db.turso.io",
			token:    "test-token",
			expected: "https://my-db.turso.io?authToken=test-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDSN(tt.url, tt.token)
			if result != tt.expected {
				t.Errorf("buildDSN(%q, %q) = %q, want %q", tt.url, tt.token, result, tt.expected)
			}
		})
	}
}

func TestEnsureDBDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Test creating directory for local file
	dbPath := filepath.Join(tmpDir, "subdir", "test.db")
	err := ensureDBDir(dbPath)
	if err != nil {
		t.Fatalf("ensureDBDir(%q) error = %v", dbPath, err)
	}

	// Verify directory was created
	dir := filepath.Dir(dbPath)
	if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
		t.Errorf("directory %q was not created", dir)
	}

	// Test that remote URLs don't create directories
	err = ensureDBDir("libsql://my-db.turso.io")
	if err != nil {
		t.Errorf("ensureDBDir(remote url) error = %v", err)
	}
}

func TestSplitSQL(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected []string
	}{
		{
			name:     "single statement",
			sql:      "CREATE TABLE test (id INTEGER PRIMARY KEY);",
			expected: []string{"CREATE TABLE test (id INTEGER PRIMARY KEY)"},
		},
		{
			name:     "multiple statements",
			sql:      "CREATE TABLE test (id INTEGER);\nINSERT INTO test VALUES (1);",
			expected: []string{"CREATE TABLE test (id INTEGER)", "INSERT INTO test VALUES (1)"},
		},
		{
			name:     "statement with semicolon in string",
			sql:      "INSERT INTO test VALUES ('hello;world');",
			expected: []string{"INSERT INTO test VALUES ('hello;world')"},
		},
		{
			name:     "statement with double quotes",
			sql:      `INSERT INTO test VALUES ("hello;world");`,
			expected: []string{`INSERT INTO test VALUES ("hello;world")`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitSQL(tt.sql)
			if len(result) != len(tt.expected) {
				t.Errorf("splitSQL returned %d statements, want %d", len(result), len(tt.expected))
				return
			}
			for i, stmt := range result {
				stmt = trimSQL(stmt)
				if stmt != tt.expected[i] {
					t.Errorf("statement %d = %q, want %q", i, stmt, tt.expected[i])
				}
			}
		})
	}
}

func trimSQL(s string) string {
	// Simple trim for test comparison
	result := ""
	inSpace := false
	for _, ch := range s {
		if ch == ' ' || ch == '\n' || ch == '\t' {
			if !inSpace {
				result += " "
				inSpace = true
			}
		} else {
			result += string(ch)
			inSpace = false
		}
	}
	// Trim leading/trailing spaces
	if len(result) > 0 && result[0] == ' ' {
		result = result[1:]
	}
	if len(result) > 0 && result[len(result)-1] == ' ' {
		result = result[:len(result)-1]
	}
	return result
}

func TestMigrationRegistry(t *testing.T) {
	// Save and restore original registry
	originalRegistry := migrationRegistry
	defer func() {
		migrationRegistry = originalRegistry
	}()

	// Clear registry
	migrationRegistry = nil

	// Test registering migrations
	Register(MigrationSet{
		Source: "test",
		Migrations: []Migration{
			{Version: 1, Name: "test1", Up: "CREATE TABLE test1 (id INTEGER)"},
			{Version: 2, Name: "test2", Up: "CREATE TABLE test2 (id INTEGER)"},
		},
	})

	sets := GetAllMigrations()
	if len(sets) != 1 {
		t.Fatalf("expected 1 migration set, got %d", len(sets))
	}
	if len(sets[0].Migrations) != 2 {
		t.Fatalf("expected 2 migrations, got %d", len(sets[0].Migrations))
	}
}

func TestOpenAndMigrate(t *testing.T) {
	// Create a temporary directory for the database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	ctx := context.Background()

	// Create the file first
	f, err := os.Create(dbPath)
	if err != nil {
		t.Fatalf("create db file error = %v", err)
	}
	f.Close()

	// Open database directly
	dsn := dbPath
	db, err := sql.Open("turso", dsn)
	if err != nil {
		t.Fatalf("sql.Open error = %v", err)
	}
	// Verify database is working
	if pingErr := db.PingContext(ctx); pingErr != nil {
		t.Fatalf("database ping error = %v", pingErr)
	}

	// Run migrations
	if migrateErr := RunMigrations(ctx, db); migrateErr != nil {
		t.Fatalf("RunMigrations error = %v", migrateErr)
	}

	// Verify migrations were applied
	var tableCount int
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='app_migrations'").Scan(&tableCount)
	if err != nil {
		t.Errorf("query migration table error = %v", err)
	}
	if tableCount != 1 {
		t.Errorf("expected app_migrations table to exist, got count %d", tableCount)
	}

	// Verify session table exists
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='sessions'").Scan(&tableCount)
	if err != nil {
		t.Errorf("query sessions table error = %v", err)
	}
	if tableCount != 1 {
		t.Errorf("expected sessions table to exist, got count %d", tableCount)
	}
}
