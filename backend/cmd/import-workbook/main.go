package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"webprogfndb/internal/db"
	"webprogfndb/internal/importer"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: import-workbook <path-to-xlsx>")
	}
	path := os.Args[1]
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("open file: %v", err)
	}
	defer f.Close()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	d, err := db.OpenPostgres(dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer d.Close()

	if err := db.MigratePostgres(d); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	version, err := db.CurrentVersion(d)
	if err != nil {
		log.Fatalf("read schema version: %v", err)
	}
	log.Printf("PostgreSQL database migrated to schema version %d", version)

	log.Printf("Parsing workbook: %s", path)
	wb, err := importer.ParseWorkbook(f)
	if err != nil {
		log.Fatalf("parse workbook: %v", err)
	}

	log.Printf("Importing workbook into database...")
	res, err := importer.Import(context.Background(), d, wb, filepath.Base(path), true)
	if err != nil {
		log.Fatalf("import workbook: %v", err)
	}

	fmt.Printf("\n=== Import Successful ===\n")
	fmt.Printf("Run ID: %d\n", res.RunID)
	for entity, count := range res.Counts {
		fmt.Printf("  %s: %d\n", entity, count)
	}
	if len(res.Skipped) > 0 {
		fmt.Printf("Skipped rows: %d (first 10):\n", len(res.Skipped))
		for i, s := range res.Skipped {
			if i >= 10 {
				break
			}
			fmt.Printf("  [%s] %s: %s\n", s.Table, s.RowKey, s.Reason)
		}
	}
	if len(res.Normalizations) > 0 {
		fmt.Printf("Normalizations logged: %d\n", len(res.Normalizations))
	}
}
