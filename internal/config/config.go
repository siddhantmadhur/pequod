package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Port           int         `toml:"port"`
	SecretKey      string      `toml:"secret_key"`
	FinishedWizard bool        `toml:"finished_wizard"`
	PersistentDir  string      `toml:"persistent_dir"`
	CacheDir       string      `toml:"cache_dir"`
	Mutex          *sync.Mutex `toml:"-"`
}

func generateRandomSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (c *Config) Write() error {
	if c.PersistentDir == "" {
		if os.Getenv("PERSISTENT_DATA") != "" {
			c.PersistentDir = os.Getenv("PERSISTENT_DATA")
		} else {
			c.PersistentDir = "/data"
		}
	}

	if err := os.MkdirAll(c.PersistentDir, 0755); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(c.PersistentDir, "config.toml"))
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}

func (c *Config) Read() error {
	if c.PersistentDir == "" {
		if os.Getenv("PERSISTENT_DATA") != "" {
			c.PersistentDir = os.Getenv("PERSISTENT_DATA")
		} else {
			c.PersistentDir = "/data"
		}
	}

	configFilePath := filepath.Join(c.PersistentDir, "config.toml")
	file, err := os.ReadFile(configFilePath)
	if err != nil {
		secret, err := generateRandomSecret()
		if err != nil {
			return err
		}

		cacheDir := os.Getenv("CACHE_DIR")
		if cacheDir == "" {
			cacheDir = filepath.Join(c.PersistentDir, "cache")
		}

		defaultConfig := Config{
			Port:           8080,
			SecretKey:      secret,
			FinishedWizard: false,
			PersistentDir:  c.PersistentDir,
			CacheDir:       cacheDir,
			Mutex:          &sync.Mutex{},
		}

		if err := defaultConfig.Write(); err != nil {
			return err
		}
		*c = defaultConfig
		return nil
	}

	if err := toml.Unmarshal(file, c); err != nil {
		return err
	}

	if c.CacheDir == "" {
		c.CacheDir = filepath.Join(c.PersistentDir, "cache")
	}

	c.Mutex = &sync.Mutex{}
	return nil
}
