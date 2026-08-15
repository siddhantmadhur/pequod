package config

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
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

func (c *Config) Write() error {
	err := os.MkdirAll(c.PersistentDir, 0755)
	if err != nil {
		return err
	}
	f, err := os.Create(c.PersistentDir + "/config.toml")
	if err != nil {
		return err
	}
	err = toml.NewEncoder(f).Encode(c)
	return err
}

func (c *Config) Read() error {
	file, err := os.ReadFile(c.PersistentDir + "/config.toml")
	if err != nil {
		key, err := rsa.GenerateKey(rand.Reader, 32*8)
		if err != nil {
			return err
		}

		var defaultConfig = Config{
			Port:           8080,
			SecretKey:      key.PublicKey.N.Text(62),
			FinishedWizard: false,
			Mutex:          &sync.Mutex{},
		}

		if os.Getenv("PERSISTENT_DATA") != "" {
			defaultConfig.PersistentDir = os.Getenv("PERSISTENT_DATA")
		} else {
			defaultConfig.PersistentDir = "/data"
		}
		err = defaultConfig.Write()
		*c = defaultConfig
		return err
	}
	err = toml.Unmarshal(file, c)

	c.Mutex = &sync.Mutex{}
	return err
}
