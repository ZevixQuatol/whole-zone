package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadRequiresConfigFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.yml"))
	if err == nil || !strings.Contains(err.Error(), "config.example.yml") {
		t.Fatalf("error = %v, want missing-file guidance", err)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	path := writeConfig(t, "server:\n  addr: ':8080'\n")
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "database.url") {
		t.Fatalf("error = %v, want database.url validation", err)
	}
}

func TestLoadRejectsMalformedYAML(t *testing.T) {
	path := writeConfig(t, "server:\n  addr: [\n")
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "解析配置文件") {
		t.Fatalf("error = %v, want YAML parse error", err)
	}
}

func TestLoadUsesYAMLValuesAndDefaults(t *testing.T) {
	path := writeConfig(t, `
server:
  addr: ":9090"
  app_url: https://huanyu.example
database:
  url: postgres://localhost/huanyu
redis:
  url: redis://localhost:6379/0
cookie:
  name: session
  secure: true
account:
  session_ttl: 48h
  reset_ttl: 15m
smtp:
  host: smtp.example.com
  user: mailer
  password: mail-password
admin:
  email: admin@example.com
  handle: admin
  password: long-admin-password
  display_name: 平台管理员
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Addr != ":9090" || cfg.AppURL != "https://huanyu.example" {
		t.Fatalf("unexpected server config: %+v", cfg)
	}
	if cfg.DatabaseURL != "postgres://localhost/huanyu" || cfg.RedisURL != "redis://localhost:6379/0" {
		t.Fatalf("unexpected service config: %+v", cfg)
	}
	if cfg.CookieName != "session" || !cfg.SecureCookie {
		t.Fatalf("unexpected cookie config: %+v", cfg)
	}
	if cfg.SessionTTL != 48*time.Hour || cfg.ResetTTL != 15*time.Minute {
		t.Fatalf("unexpected account duration: %+v", cfg)
	}
	if cfg.SMTPPort != "587" || cfg.SMTPFrom != "no-reply@huanyu.local" {
		t.Fatalf("smtp defaults were not applied: %+v", cfg)
	}
	if cfg.Admin.Email != "admin@example.com" || cfg.Admin.DisplayName != "平台管理员" {
		t.Fatalf("unexpected admin config: %+v", cfg.Admin)
	}
}

func TestLoadRejectsInvalidDurations(t *testing.T) {
	for _, value := range []string{"tomorrow", "0s", "-1h"} {
		t.Run(value, func(t *testing.T) {
			path := writeConfig(t, "database:\n  url: postgres://localhost/huanyu\naccount:\n  session_ttl: \""+value+"\"\n")
			if _, err := Load(path); err == nil {
				t.Fatalf("expected duration %q to fail", value)
			}
		})
	}
}

func TestPathUsesOptionalOverride(t *testing.T) {
	if path := Path(func(string) string { return "" }); path != DefaultPath {
		t.Fatalf("path = %q, want %q", path, DefaultPath)
	}
	if path := Path(func(string) string { return " D:/config/huanyu.yml " }); path != "D:/config/huanyu.yml" {
		t.Fatalf("override path = %q", path)
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
