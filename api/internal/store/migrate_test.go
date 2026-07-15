package store

import (
	"os"
	"reflect"
	"testing"
)

func TestListMigrationsRequiresOrderedPairs(t *testing.T) {
	migrations, err := ListMigrations(os.DirFS("../../migrations"))
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}

	versions := make([]int64, 0, len(migrations))
	for _, migration := range migrations {
		versions = append(versions, migration.Version)
		if migration.Up == "" || migration.Down == "" {
			t.Fatalf("migration %d is missing a direction", migration.Version)
		}
	}

	if !reflect.DeepEqual(versions, []int64{1}) {
		t.Fatalf("versions = %v, want [1]", versions)
	}
}
