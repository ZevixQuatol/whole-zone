package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ZevixQuatol/whole-zone/api/internal/account"
	"github.com/ZevixQuatol/whole-zone/api/internal/config"
	"github.com/ZevixQuatol/whole-zone/api/internal/store"
	"github.com/ZevixQuatol/whole-zone/api/migrations"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load(config.Path(os.Getenv))
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer pool.Close()
	if err := store.Up(ctx, pool, migrations.FS); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	repo := account.NewPG(pool)
	service := account.NewService(repo, cfg.SessionTTL, time.Now)
	if cfg.SMTPHost != "" {
		service.SetMailer(account.NewSMTPMailer(
			cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom, cfg.AppURL,
		), cfg.ResetTTL)
	}
	handler := account.NewHandler(service, cfg.AppURL, cfg.CookieName, cfg.SecureCookie)
	handler.RegisterRoutes(router.Group("/api/v1"))

	if err := router.Run(cfg.Addr); err != nil {
		log.Fatalf("run api: %v", err)
	}
}
