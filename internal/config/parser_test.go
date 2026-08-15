package config

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestWrite(t *testing.T) {
	// ensure no prexisting config

	var config Config
	tempDir := t.TempDir()
	config.PersistentDir = tempDir + "/data"
	config.Port = 8080
	config.CacheDir = tempDir + "/cache"

	err := config.Write()
	if err != nil {
		log.Printf("[ERROR]: %s\n", err.Error())
		t.FailNow()
	}

	f, err := os.ReadFile(tempDir + "/data/config.toml")
	if err != nil {
		log.Printf("[ERROR]: %s\n", err.Error())
		t.FailNow()
	}
	expectedOutput := fmt.Sprintf(`port = 8080
secret_key = ""
finished_wizard = false
persistent_dir = "%s/data"
cache_dir = "%s/cache"
`, tempDir, tempDir)
	if string(f) != expectedOutput {
		t.FailNow()
	}
}

func TestRead(t *testing.T) {

	// Above test confirms functionality of Write()
	var tempConfig Config
	tempDir := t.TempDir()
	tempConfig.PersistentDir = tempDir + "/data"
	tempConfig.Port = 8080
	tempConfig.CacheDir = tempDir + "/cache"

	err := tempConfig.Write()
	if err != nil {
		log.Printf("[ERROR]: %s\n", err.Error())
		t.FailNow()
	}

	var cfg Config
	cfg.PersistentDir = tempDir + "/data"
	err = cfg.Read()
	if err != nil {
		log.Printf("[ERROR]: %s\n", err.Error())
		t.FailNow()
	}
	if cfg.CacheDir != tempConfig.CacheDir || cfg.PersistentDir != tempConfig.PersistentDir || cfg.Port != tempConfig.Port {
		t.FailNow()
	}
}
