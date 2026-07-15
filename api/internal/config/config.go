package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

const DefaultPath = "config.yml"

type Config struct {
	Addr         string
	AppURL       string
	DatabaseURL  string
	RedisURL     string
	CookieName   string
	SecureCookie bool
	SessionTTL   time.Duration
	ResetTTL     time.Duration
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	Admin        Admin
}

type Admin struct {
	Email       string
	Handle      string
	Password    string
	DisplayName string
}

type fileConfig struct {
	Server struct {
		Addr   string `yaml:"addr"`
		AppURL string `yaml:"app_url"`
	} `yaml:"server"`
	Database struct {
		URL string `yaml:"url"`
	} `yaml:"database"`
	Redis struct {
		URL string `yaml:"url"`
	} `yaml:"redis"`
	Cookie struct {
		Name   string `yaml:"name"`
		Secure bool   `yaml:"secure"`
	} `yaml:"cookie"`
	Account struct {
		SessionTTL string `yaml:"session_ttl"`
		ResetTTL   string `yaml:"reset_ttl"`
	} `yaml:"account"`
	SMTP struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		From     string `yaml:"from"`
	} `yaml:"smtp"`
	Admin struct {
		Email       string `yaml:"email"`
		Handle      string `yaml:"handle"`
		Password    string `yaml:"password"`
		DisplayName string `yaml:"display_name"`
	} `yaml:"admin"`
}

// Path 优先使用显式路径，并兼容 GoLand 从项目根目录、api 或 cmd 子目录直接启动。
func Path(getenv func(string) string) string {
	if path := strings.TrimSpace(getenv("HUANYU_CONFIG")); path != "" {
		return path
	}
	for _, path := range []string{
		DefaultPath,
		filepath.Join("api", DefaultPath),
		filepath.Join("..", "..", DefaultPath),
	} {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return DefaultPath
}

func Load(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("配置文件 %s 不存在，请先复制 config.example.yml", path)
		}
		return Config{}, fmt.Errorf("读取配置文件 %s: %w", path, err)
	}

	var file fileConfig
	if err := yaml.Unmarshal(content, &file); err != nil {
		return Config{}, fmt.Errorf("解析配置文件 %s: %w", path, err)
	}

	sessionTTL, err := duration(file.Account.SessionTTL, 30*24*time.Hour)
	if err != nil {
		return Config{}, fmt.Errorf("account.session_ttl: %w", err)
	}
	resetTTL, err := duration(file.Account.ResetTTL, 30*time.Minute)
	if err != nil {
		return Config{}, fmt.Errorf("account.reset_ttl: %w", err)
	}
	databaseURL := strings.TrimSpace(file.Database.URL)
	if databaseURL == "" {
		return Config{}, errors.New("database.url 不能为空")
	}

	return Config{
		Addr:         fallback(file.Server.Addr, ":8080"),
		AppURL:       fallback(file.Server.AppURL, "http://localhost:3000"),
		DatabaseURL:  databaseURL,
		RedisURL:     strings.TrimSpace(file.Redis.URL),
		CookieName:   fallback(file.Cookie.Name, "hy_session"),
		SecureCookie: file.Cookie.Secure,
		SessionTTL:   sessionTTL,
		ResetTTL:     resetTTL,
		SMTPHost:     strings.TrimSpace(file.SMTP.Host),
		SMTPPort:     fallback(file.SMTP.Port, "587"),
		SMTPUser:     strings.TrimSpace(file.SMTP.User),
		SMTPPassword: file.SMTP.Password,
		SMTPFrom:     fallback(file.SMTP.From, "no-reply@huanyu.local"),
		Admin: Admin{
			Email:       strings.TrimSpace(file.Admin.Email),
			Handle:      strings.TrimSpace(file.Admin.Handle),
			Password:    file.Admin.Password,
			DisplayName: strings.TrimSpace(file.Admin.DisplayName),
		},
	}, nil
}

func fallback(current, defaultValue string) string {
	if current = strings.TrimSpace(current); current != "" {
		return current
	}
	return defaultValue
}

func duration(current string, defaultValue time.Duration) (time.Duration, error) {
	if current = strings.TrimSpace(current); current == "" {
		return defaultValue, nil
	}
	value, err := time.ParseDuration(current)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, errors.New("必须大于 0")
	}
	return value, nil
}
