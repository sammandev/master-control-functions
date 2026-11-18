package db

import (
	"embed"
	"fmt"
	"path"
	"sort"
	"strconv"
)

//go:embed migrations/*.sql postgres_migrations/*.sql
var migrationFS embed.FS

// migration is one schema change, loaded from an embedded .sql file whose
// name starts with its version number (0001_functions.sql → version 1).
// SQL files (not code strings) keep the schema readable by tools that
// understand plain SQL, e.g. sqlc.
type migration struct {
	Version int
	Name    string
	SQL     string
}

// loadMigrations reads every migrations/*.sql file, ordered by version
// number parsed from the filename prefix. Duplicate versions fail loudly.
func loadMigrations() ([]migration, error) {
	return loadMigrationsFrom("migrations")
}

func loadMigrationsFrom(directory string) ([]migration, error) {
	entries, err := migrationFS.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var result []migration
	for _, e := range entries {
		version, err := versionFromFile(e.Name())
		if err != nil {
			return nil, err
		}
		sql, err := migrationFS.ReadFile(path.Join(directory, e.Name()))
		if err != nil {
			return nil, err
		}
		result = append(result, migration{
			Version: version,
			Name:    e.Name(),
			SQL:     string(sql),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Version < result[j].Version })
	seen := map[int]bool{}
	for _, m := range result {
		if seen[m.Version] {
			return nil, fmt.Errorf("duplicate migration version %d", m.Version)
		}
		seen[m.Version] = true
	}
	return result, nil
}

// versionFromFile parses "0001_functions.sql" → (1, nil).
func versionFromFile(name string) (int, error) {
	for i := 0; i < len(name); i++ {
		if name[i] == '_' {
			v, err := strconv.Atoi(name[:i])
			if err != nil {
				return 0, fmt.Errorf("migration %s: bad version prefix: %w", name, err)
			}
			return v, nil
		}
	}
	return 0, fmt.Errorf("migration %s: missing version prefix", name)
}
