package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadStoreConfig(t *testing.T) {
	t.Setenv("SERVER_ADDR", ":8080")
	t.Setenv("ADMIN_USERNAME", "lane")
	t.Setenv("ADMIN_PASSWORD_HASH", "hash")
	t.Setenv("AUTH_TOKEN", "token")

	t.Run("mysql dsn is required", func(t *testing.T) {
		t.Setenv("ADMIN_MYSQL_DSN", "")

		if _, err := Load(); err == nil {
			t.Fatal("expected mysql dsn to be required")
		}
	})

	t.Run("mysql dsn loads successfully", func(t *testing.T) {
		t.Setenv("ADMIN_MYSQL_DSN", "lane:password@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&loc=Local")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("load config: %v", err)
		}
		if cfg.Store.MySQLDSN == "" {
			t.Fatal("expected mysql dsn to be loaded")
		}
	})
}

func TestLoadEnvFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, ".env")
	content := "APP_NAME=LaneBlog Admin API\nADMIN_MYSQL_DSN=root:pass@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&loc=Local\nADMIN_PASSWORD_HASH=123456 # comment\n"
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	oldAppName, hadAppName := os.LookupEnv("APP_NAME")
	oldDSN, hadDSN := os.LookupEnv("ADMIN_MYSQL_DSN")
	oldHash, hadHash := os.LookupEnv("ADMIN_PASSWORD_HASH")
	_ = os.Unsetenv("APP_NAME")
	_ = os.Unsetenv("ADMIN_MYSQL_DSN")
	_ = os.Unsetenv("ADMIN_PASSWORD_HASH")
	t.Cleanup(func() {
		if hadAppName {
			_ = os.Setenv("APP_NAME", oldAppName)
		} else {
			_ = os.Unsetenv("APP_NAME")
		}
		if hadDSN {
			_ = os.Setenv("ADMIN_MYSQL_DSN", oldDSN)
		} else {
			_ = os.Unsetenv("ADMIN_MYSQL_DSN")
		}
		if hadHash {
			_ = os.Setenv("ADMIN_PASSWORD_HASH", oldHash)
		} else {
			_ = os.Unsetenv("ADMIN_PASSWORD_HASH")
		}
	})

	if err := loadEnvFile(file); err != nil {
		t.Fatalf("load env file: %v", err)
	}

	if got := os.Getenv("APP_NAME"); got != "LaneBlog Admin API" {
		t.Fatalf("unexpected APP_NAME: %s", got)
	}
	if got := os.Getenv("ADMIN_MYSQL_DSN"); got == "" {
		t.Fatal("expected ADMIN_MYSQL_DSN to be loaded")
	}
	if got := os.Getenv("ADMIN_PASSWORD_HASH"); got != "123456" {
		t.Fatalf("unexpected ADMIN_PASSWORD_HASH: %s", got)
	}
}
