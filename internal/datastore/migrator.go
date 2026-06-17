package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// Migration represents a single database migration.
type Migration struct {
	// Version is the unique version number for this migration.
	Version int64
	// Name is a human-readable name for the migration.
	Name string
	// Up is the SQL to apply the migration.
	Up string
	// Down is the SQL to rollback the migration (optional).
	Down string
}

// MigrationSet represents a collection of migrations from a single source.
type MigrationSet struct {
	// Source identifies where these migrations come from (e.g., "core", "snip").
	Source string
	// Migrations is the list of migrations in this set.
	Migrations []Migration
}

// Register adds a migration set to the global registry.
func Register(set MigrationSet) {
	migrationRegistry = append(migrationRegistry, set)
}

// migrationRegistry holds all registered migration sets.
var migrationRegistry []MigrationSet

// GetAllMigrations returns all registered migrations sorted by version.
func GetAllMigrations() []MigrationSet {
	return migrationRegistry
}

// RunMigrations applies all pending migrations to the database.
func RunMigrations(ctx context.Context, db *sql.DB) error {
	if len(migrationRegistry) == 0 {
		return nil
	}

	// Merge all migrations from all sources into a flat list
	var allMigrations []Migration
	for _, set := range migrationRegistry {
		allMigrations = append(allMigrations, set.Migrations...)
	}

	// Sort by version to ensure correct order
	sort.Slice(allMigrations, func(i, j int) bool {
		return allMigrations[i].Version < allMigrations[j].Version
	})

	// Ensure migration table exists
	if err := ensureMigrationTable(ctx, db); err != nil {
		return fmt.Errorf("ensure migration table: %w", err)
	}

	// Get current version
	currentVersion, err := getCurrentVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	// Apply pending migrations
	for _, m := range allMigrations {
		if m.Version > currentVersion {
			fmt.Printf("datastore: applying migration %d: %s\n", m.Version, m.Name)
			if err := runMigration(ctx, db, m); err != nil {
				return fmt.Errorf("migration %d (%s): %w", m.Version, m.Name, err)
			}
		}
	}

	return nil
}

func ensureMigrationTable(ctx context.Context, db *sql.DB) error {
	sql := `
		CREATE TABLE IF NOT EXISTS app_migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version INTEGER NOT NULL,
			name TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE UNIQUE INDEX IF NOT EXISTS idx_app_migrations_version 
		ON app_migrations(version);
	`
	_, err := db.ExecContext(ctx, sql)
	return err
}

func getCurrentVersion(ctx context.Context, db *sql.DB) (int64, error) {
	var version int64
	err := db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM app_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func runMigration(ctx context.Context, db *sql.DB, m Migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	// Split and execute the migration SQL (may contain multiple statements)
	statements := splitSQL(m.Up)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("execute statement: %w\nSQL: %s", err, stmt)
		}
	}

	// Record the migration
	insertSQL := "INSERT INTO app_migrations (version, name) VALUES (?, ?)"
	if _, err := tx.ExecContext(ctx, insertSQL, m.Version, m.Name); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	return tx.Commit()
}

// splitSQL splits a SQL string into individual statements.
// This is a simple implementation that handles semicolons.
func splitSQL(sql string) []string {
	var statements []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false

	for i := 0; i < len(sql); i++ {
		ch := sql[i]

		switch {
		case ch == '\'' && !inDoubleQuote:
			inSingleQuote = !inSingleQuote
			current.WriteByte(ch)
		case ch == '"' && !inSingleQuote:
			inDoubleQuote = !inDoubleQuote
			current.WriteByte(ch)
		case ch == ';' && !inSingleQuote && !inDoubleQuote:
			statements = append(statements, current.String())
			current.Reset()
		default:
			current.WriteByte(ch)
		}
	}

	// Add the last statement if there's content
	if current.Len() > 0 {
		statements = append(statements, current.String())
	}

	return statements
}
