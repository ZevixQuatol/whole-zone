package config

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Config struct {
	Addr         string
	AppURL       string
	DatabaseURL  string
	CookieName   string
	SecureCookie bool
	SessionTTL   time.Duration
	ResetTTL     time.Duration
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
}

func Load(getenv func(string) string) (Config, error) {
	databaseURL := getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	secure, err := boolValue(getenv("COOKIE_SECURE"), false)
	if err != nil {
		return Config{}, fmt.Errorf("COOKIE_SECURE: %w", err)
	}
	sessionTTL, err := durationValue(getenv("SESSION_TTL"), 30*24*time.Hour)
	if err != nil {
		return Config{}, fmt.Errorf("SESSION_TTL: %w", err)
	}
	resetTTL, err := durationValue(getenv("RESET_TTL"), 30*time.Minute)
	if err != nil {
		return Config{}, fmt.Errorf("RESET_TTL: %w", err)
	}

	return Config{
		Addr:         value(getenv("HTTP_ADDR"), ":8080"),
		AppURL:       value(getenv("APP_URL"), "http://localhost:3000"),
		DatabaseURL:  databaseURL,
		CookieName:   value(getenv("COOKIE_NAME"), "hy_session"),
		SecureCookie: secure,
		SessionTTL:   sessionTTL,
		ResetTTL:     resetTTL,
		SMTPHost:     getenv("SMTP_HOST"),
		SMTPPort:     value(getenv("SMTP_PORT"), "587"),
		SMTPUser:     getenv("SMTP_USER"),
		SMTPPassword: getenv("SMTP_PASSWORD"),
		SMTPFrom:     value(getenv("SMTP_FROM"), "no-reply@huanyu.local"),
	}, nil
}

func value(current, fallback string) string {
	if current == "" {
		return fallback
	}
	return current
}

func boolValue(current string, fallback bool) (bool, error) {
	if current == "" {
		return fallback, nil
	}
	return strconv.ParseBool(current)
}

func durationValue(current string, fallback time.Duration) (time.Duration, error) {
	if current == "" {
		return fallback, nil
	}
	return time.ParseDuration(current)
}
