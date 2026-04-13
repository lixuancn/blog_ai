package config

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const defaultPasswordHash = "58cd2c961705ac711493b038c04bb98c7c926ed0"
const defaultAuthToken = "change-this-token-in-production"
const defaultCookieName = "admin_token"

type Config struct {
	App    AppConfig
	Server ServerConfig
	Auth   AuthConfig
	Store  StoreConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type ServerConfig struct {
	Addr string
}

type AuthConfig struct {
	Username     string
	PasswordHash string
	Token        string
	CookieName   string
}

type StoreConfig struct {
	MySQLDSN string
}

func Load() (Config, error) {
	loadEnvFiles()

	cfg := Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "LaneBlog Admin API"),
			Env:  getEnv("APP_ENV", "local"),
		},
		Server: ServerConfig{
			Addr: getEnv("SERVER_ADDR", ":8080"),
		},
		Auth: AuthConfig{
			Username:     getEnv("ADMIN_USERNAME", "lane"),
			PasswordHash: getEnv("ADMIN_PASSWORD_HASH", defaultPasswordHash),
			Token:        getEnv("AUTH_TOKEN", defaultAuthToken),
			CookieName:   defaultCookieName,
		},
		Store: StoreConfig{
			MySQLDSN: getEnv("ADMIN_MYSQL_DSN", ""),
		},
	}

	if cfg.Server.Addr == "" {
		return Config{}, errors.New("SERVER_ADDR cannot be empty")
	}

	if cfg.Auth.Username == "" {
		return Config{}, errors.New("ADMIN_USERNAME cannot be empty")
	}

	if cfg.Auth.PasswordHash == "" {
		return Config{}, errors.New("ADMIN_PASSWORD_HASH cannot be empty")
	}

	if cfg.Auth.Token == "" {
		return Config{}, errors.New("AUTH_TOKEN cannot be empty")
	}

	if cfg.Store.MySQLDSN == "" {
		return Config{}, errors.New("ADMIN_MYSQL_DSN cannot be empty")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func loadEnvFiles() {
	candidates := []string{
		".env",
		"configs/.env",
		filepath.Join("..", ".env"),
		filepath.Join("..", "configs", ".env"),
		filepath.Join("..", "..", ".env"),
		filepath.Join("..", "..", "configs", ".env"),
	}

	for _, path := range candidates {
		_ = loadEnvFile(path)
	}
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = cleanEnvValue(value)
		if key == "" || value == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}

	return scanner.Err()
}

func cleanEnvValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") && len(value) >= 2 {
		return strings.TrimSpace(value[1 : len(value)-1])
	}

	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") && len(value) >= 2 {
		return strings.TrimSpace(value[1 : len(value)-1])
	}

	if index := strings.Index(value, " #"); index >= 0 {
		value = value[:index]
	}

	return strings.TrimSpace(value)
}
