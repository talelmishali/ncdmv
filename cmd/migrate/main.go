package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/file"
)

const (
	dbPathEnvVar = "NCDMV_DB_PATH"
)

var (
	down           = flag.Bool("down", false, "up or down")
	count          = flag.Int("count", 0, "number of migrations")
	dbPathFlag     = flag.String("db_path", "./database/ncdmv.db", "path to SQLite DB file")
	migrationsPath = flag.String("migrations_path", "./database/migrations", "path to migrations directory")
)

func main() {
	dbPath := os.Getenv(dbPathEnvVar)
	if dbPath == "" {
		if *dbPathFlag != "" {
			dbPath = *dbPathFlag
		} else {
			log.Fatalf("No DB path specified")
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	instance, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		log.Fatal(err)
	}

	fileSource, err := (&file.File{}).Open(*migrationsPath)
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithInstance("file", fileSource, "sqlite", instance)
	if err != nil {
		log.Fatal(err)
	}

	if !*down {
		if *count == 0 {
			if err := m.Up(); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := m.Steps(*count); err != nil {
				log.Fatal(err)
			}
		}
	} else {
		if *count == 0 {
			if err := m.Down(); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := m.Steps(*count * -1); err != nil {
				log.Fatal(err)
			}
		}
	}
}
