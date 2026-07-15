package config

import (
	"testing"
	"time"
)

func TestLoadRequiresDatabaseURL(t *testing.T) {
	_, err := Load(func(string) string { return "" })
	if err == nil {
		t.Fatal("expected missing DATABASE_URL to fail")
	}
}

func TestLoadUsesDevelopmentDefaults(t *testing.T) {
	cfg, err := Load(func(key string) string {
		if key == "DATABASE_URL" {
			return "postgres://localhost/huanyu"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Addr != ":8080" {
		t.Fatalf("Addr = %q, want :8080", cfg.Addr)
	}
	if cfg.AppURL != "http://localhost:3000" {
		t.Fatalf("AppURL = %q, want http://localhost:3000", cfg.AppURL)
	}
	if cfg.SecureCookie {
		t.Fatal("SecureCookie = true, want false")
	}
	if cfg.SessionTTL != 30*24*time.Hour {
		t.Fatalf("SessionTTL = %s, want 720h", cfg.SessionTTL)
	}
	if cfg.ResetTTL != 30*time.Minute {
		t.Fatalf("ResetTTL = %s, want 30m", cfg.ResetTTL)
	}
	if cfg.CookieName != "hy_session" {
		t.Fatalf("CookieName = %q, want hy_session", cfg.CookieName)
	}
}

func TestLoadReadsOverrides(t *testing.T) {
	env := map[string]string{
		"DATABASE_URL":  "postgres://localhost/huanyu",
		"HTTP_ADDR":     ":9090",
		"APP_URL":       "https://huanyu.example",
		"COOKIE_NAME":   "session",
		"COOKIE_SECURE": "true",
		"SESSION_TTL":   "48h",
		"RESET_TTL":     "15m",
	}

	cfg, err := Load(func(key string) string { return env[key] })
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Addr != ":9090" || cfg.AppURL != "https://huanyu.example" {
		t.Fatalf("unexpected address overrides: %+v", cfg)
	}
	if cfg.CookieName != "session" || !cfg.SecureCookie {
		t.Fatalf("unexpected cookie overrides: %+v", cfg)
	}
	if cfg.SessionTTL != 48*time.Hour || cfg.ResetTTL != 15*time.Minute {
		t.Fatalf("unexpected duration overrides: %+v", cfg)
	}
}

func TestLoadRejectsInvalidOverrides(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "secure cookie", key: "COOKIE_SECURE", value: "sometimes"},
		{name: "session ttl", key: "SESSION_TTL", value: "tomorrow"},
		{name: "reset ttl", key: "RESET_TTL", value: "soon"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(func(key string) string {
				if key == "DATABASE_URL" {
					return "postgres://localhost/huanyu"
				}
				if key == tt.key {
					return tt.value
				}
				return ""
			})
			if err == nil {
				t.Fatalf("expected %s to reject %q", tt.key, tt.value)
			}
		})
	}
}
