package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/siddhantmadhur/pequod/internal/config"
)

func TestWrite(t *testing.T) {
	var cfg config.Config
	tempDir := t.TempDir()
	cfg.PersistentDir = tempDir + "/data"
	cfg.Port = 8080
	cfg.CacheDir = tempDir + "/cache"

	if err := cfg.Write(); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	f, err := os.ReadFile(tempDir + "/data/config.toml")
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	expectedOutput := fmt.Sprintf(`port = 8080
secret_key = ""
finished_wizard = false
persistent_dir = "%s/data"
cache_dir = "%s/cache"
`, tempDir, tempDir)
	if string(f) != expectedOutput {
		t.Fatalf("expected:\n%s\ngot:\n%s", expectedOutput, string(f))
	}
}

func TestRead(t *testing.T) {
	var tempConfig config.Config
	tempDir := t.TempDir()
	tempConfig.PersistentDir = tempDir + "/data"
	tempConfig.Port = 8080
	tempConfig.CacheDir = tempDir + "/cache"

	if err := tempConfig.Write(); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	var cfg config.Config
	cfg.PersistentDir = tempDir + "/data"
	if err := cfg.Read(); err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if cfg.CacheDir != tempConfig.CacheDir || cfg.PersistentDir != tempConfig.PersistentDir || cfg.Port != tempConfig.Port {
		t.Fatal("config values do not match expected")
	}
}
