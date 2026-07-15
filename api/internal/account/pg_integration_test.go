package account_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/ZevixQuatol/whole-zone/api/internal/account"
	"github.com/ZevixQuatol/whole-zone/api/internal/store"
	"github.com/ZevixQuatol/whole-zone/api/migrations"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresRegisterAndLogin(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open admin pool: %v", err)
	}
	defer admin.Close()
	schema := "account_test_" + uuid.NewString()
	schema = strings.ReplaceAll(schema, "-", "_")
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	defer func() { _, _ = admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE") }()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse test database url: %v", err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("open test pool: %v", err)
	}
	defer pool.Close()
	if err := store.Up(ctx, pool, migrations.FS); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	assertSchemaComments(t, ctx, pool)

	repo := account.NewPG(pool)
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	inviteToken, err := account.NewToken()
	if err != nil {
		t.Fatal(err)
	}
	invite := account.Invite{ID: uuid.New(), CodeHash: inviteToken.Hash, MaxUses: 1, ExpiresAt: now.Add(time.Hour), CreatedAt: now}
	if err := repo.CreateInvite(ctx, invite); err != nil {
		t.Fatalf("create invite: %v", err)
	}
	service := account.NewService(repo, 24*time.Hour, func() time.Time { return now })
	result, err := service.Register(ctx, account.RegisterInput{
		InviteCode: inviteToken.Raw, Email: "dev@example.com", Handle: "dev_user",
		Password: "long-enough-password", DisplayName: "开发者",
	}, "integration-test")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if result.User.Handle != "dev_user" {
		t.Fatalf("registered handle = %q", result.User.Handle)
	}
	if _, err := service.Login(ctx, account.LoginInput{Login: "dev@example.com", Password: "long-enough-password"}, "integration-test"); err != nil {
		t.Fatalf("login: %v", err)
	}
	if _, err := service.Register(ctx, account.RegisterInput{
		InviteCode: inviteToken.Raw, Email: "second@example.com", Handle: "second_user",
		Password: "long-enough-password", DisplayName: "第二个用户",
	}, "integration-test"); err != account.ErrInviteInvalid {
		t.Fatalf("second registration error = %v, want ErrInviteInvalid", err)
	}
}

func assertSchemaComments(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	tables := []string{
		"schema_migrations", "users", "invites", "sessions",
		"password_resets", "system_notifications", "account_events",
	}
	rows, err := pool.Query(ctx, `
        SELECT c.relname, ''::text AS field, COALESCE(obj_description(c.oid, 'pg_class'), '')
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = current_schema() AND c.relname = ANY($1::text[])
        UNION ALL
        SELECT c.relname, a.attname, COALESCE(col_description(c.oid, a.attnum), '')
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum > 0 AND NOT a.attisdropped
        WHERE n.nspname = current_schema() AND c.relname = ANY($1::text[])
        ORDER BY 1, 2`, tables)
	if err != nil {
		t.Fatalf("query schema comments: %v", err)
	}
	defer rows.Close()

	seenTables := make(map[string]bool, len(tables))
	for rows.Next() {
		var table, field, comment string
		if err := rows.Scan(&table, &field, &comment); err != nil {
			t.Fatalf("scan schema comment: %v", err)
		}
		seenTables[table] = true
		name := table
		if field != "" {
			name += "." + field
		}
		if strings.TrimSpace(comment) == "" {
			t.Errorf("%s is missing a comment", name)
			continue
		}
		if !strings.ContainsFunc(comment, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
			t.Errorf("%s comment %q is not Chinese", name, comment)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate schema comments: %v", err)
	}
	for _, table := range tables {
		if !seenTables[table] {
			t.Errorf("table %s was not found", table)
		}
	}
}
