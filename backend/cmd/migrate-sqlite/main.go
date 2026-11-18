//go:build sqlite_migration

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"webprogfndb/internal/db"
	"webprogfndb/internal/migration"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "migration failed:", err)
		os.Exit(1)
	}
}

func run() error {
	sourcePath := flag.String("source", "", "existing SQLite backup file; opened read-only")
	confirmed := flag.Bool("confirm-writes-stopped", false, "confirm both services are stopped for cutover")
	verifyOnly := flag.Bool("verify-only", false, "compare all application rows without changing either database")
	flag.Parse()
	if *sourcePath == "" || !*confirmed || flag.NArg() != 0 {
		return fmt.Errorf("supply --source PATH and --confirm-writes-stopped; DATABASE_URL selects PostgreSQL")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	source, err := migration.OpenSource(*sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	target, err := db.OpenPostgres(os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer target.Close()
	return migration.Migrate(ctx, source, target, *verifyOnly)
}
