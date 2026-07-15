package store

import (
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
)

var migrationName = regexp.MustCompile(`^(\d+)_([a-z0-9_]+)\.(up|down)\.sql$`)

type Migration struct {
	Version int64
	Name    string
	Up      string
	Down    string
}

func ListMigrations(source fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(source, ".")
	if err != nil {
		return nil, err
	}

	byVersion := make(map[int64]*Migration)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		match := migrationName.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		version, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("migration %q: %w", entry.Name(), err)
		}
		body, err := fs.ReadFile(source, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", entry.Name(), err)
		}

		migration := byVersion[version]
		if migration == nil {
			migration = &Migration{Version: version, Name: match[2]}
			byVersion[version] = migration
		}
		if migration.Name != match[2] {
			return nil, fmt.Errorf("migration %d has conflicting names", version)
		}
		if match[3] == "up" {
			if migration.Up != "" {
				return nil, fmt.Errorf("migration %d has duplicate up files", version)
			}
			migration.Up = string(body)
		} else {
			if migration.Down != "" {
				return nil, fmt.Errorf("migration %d has duplicate down files", version)
			}
			migration.Down = string(body)
		}
	}

	migrations := make([]Migration, 0, len(byVersion))
	for _, migration := range byVersion {
		if migration.Up == "" || migration.Down == "" {
			return nil, fmt.Errorf("migration %d must have up and down files", migration.Version)
		}
		migrations = append(migrations, *migration)
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	return migrations, nil
}
