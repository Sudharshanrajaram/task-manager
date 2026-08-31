package config

import (
	"os"
	"testing"
)

func TestConfigLoad(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("ENV", "test")
	os.Setenv("DB_HOST", "127.0.0.1")
	os.Setenv("DB_NAME", "testdb")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("ENV")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_NAME")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Server.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Server.Port)
	}
	if cfg.Server.Env != "test" {
		t.Errorf("expected env test, got %s", cfg.Server.Env)
	}
	if cfg.Database.Host != "127.0.0.1" {
		t.Errorf("expected db host 127.0.0.1, got %s", cfg.Database.Host)
	}
	if cfg.Database.DBName != "testdb" {
		t.Errorf("expected db name testdb, got %s", cfg.Database.DBName)
	}
	if cfg.Database.DSN() == "" {
		t.Errorf("expected non-empty DSN")
	}
	if cfg.Redis.Addr() == "" {
		t.Errorf("expected non-empty Redis Addr")
	}
}
