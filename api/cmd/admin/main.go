package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/ZevixQuatol/whole-zone/api/internal/account"
	"github.com/ZevixQuatol/whole-zone/api/internal/config"
	"github.com/ZevixQuatol/whole-zone/api/internal/store"
	"github.com/ZevixQuatol/whole-zone/api/migrations"
	"github.com/google/uuid"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "bootstrap" {
		log.Fatal("usage: go run ./cmd/admin bootstrap")
	}
	cfg, err := config.Load(config.Path(os.Getenv))
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	password := cfg.Admin.Password
	email, handle, displayName, err := account.NormalizeRegistration(account.RegisterInput{
		Email:       cfg.Admin.Email,
		Handle:      cfg.Admin.Handle,
		Password:    password,
		DisplayName: cfg.Admin.DisplayName,
	})
	if err != nil {
		log.Fatalf("validate admin account: %v", err)
	}
	hash, err := account.HashPassword(password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer pool.Close()
	if err := store.Up(ctx, pool, migrations.FS); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	now := time.Now().UTC()
	user := account.User{
		ID: uuid.New(), Email: email, Handle: handle, PasswordHash: hash,
		Role: account.RoleAdmin, Status: account.StatusActive, DisplayName: displayName,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := account.NewPG(pool).BootstrapAdmin(ctx, user); err != nil {
		log.Fatalf("bootstrap admin: %v", err)
	}
	log.Printf("platform admin %s is ready", email)
}
