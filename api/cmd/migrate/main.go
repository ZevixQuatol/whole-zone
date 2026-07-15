package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/ZevixQuatol/whole-zone/api/internal/config"
	"github.com/ZevixQuatol/whole-zone/api/internal/store"
	"github.com/ZevixQuatol/whole-zone/api/migrations"
)

func main() {
	cfg, err := config.Load(config.Path(os.Getenv))
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer pool.Close()
	if err := store.Up(ctx, pool, migrations.FS); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	log.Print("database migrations applied")
}
