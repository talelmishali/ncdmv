package models

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed testdata/migrations/*.sql
var migrations embed.FS

// RunMigrations runs all migrations in the given path against the provided SQLite DB.
//
// If count is set to 0, all migrations will be run. If down is set to true, down
// migrations will be run.
func RunMigrations(dbPath string, count int, down bool) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open DB %q: %w", dbPath, err)
	}
	defer db.Close()

	instance, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("invalid SQLite DB file: %w", err)
	}

	d, err := iofs.New(migrations, "testdata/migrations")
	if err != nil {
		return fmt.Errorf("failed to open migrations: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "sqlite", instance)
	if err != nil {
		log.Fatal(err)
	}

	f := func() error {
		if count == 0 {
			if !down {
				if err := m.Up(); err != nil {
					return fmt.Errorf("failed to run all up migrations: %w", err)
				}
			} else {
				if err := m.Down(); err != nil {
					return fmt.Errorf("failed to run all down migrations: %w", err)
				}
			}
		} else {
			if !down {
				if err := m.Steps(count); err != nil {
					return fmt.Errorf("failed to run %d up migrations: %w", count, err)
				}
			} else {
				if err := m.Steps(count * -1); err != nil {
					return fmt.Errorf("failed to run %d down migrations: %w", count, err)
				}
			}
		}
		return nil
	}
	if err := f(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
