package store

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

func Up(ctx context.Context, pool *pgxpool.Pool, source fs.FS) error {
	migrations, err := ListMigrations(source)
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version bigint PRIMARY KEY,
            name text NOT NULL,
            applied_at timestamptz NOT NULL DEFAULT now()
        )`); err != nil {
		return fmt.Errorf("create schema migrations: %w", err)
	}

	for _, migration := range migrations {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		var applied bool
		err = tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, migration.Version).Scan(&applied)
		if err == nil && !applied {
			_, err = tx.Exec(ctx, migration.Up)
		}
		if err == nil && !applied {
			_, err = tx.Exec(ctx, `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, migration.Version, migration.Name)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %d: %w", migration.Version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %d: %w", migration.Version, err)
		}
	}
	return nil
}
