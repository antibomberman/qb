package internal

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type Config struct {
	Driver      string `env:"DRIVER"`
	DSN         string `env:"DSN"`
	MaxAttempts int    `env:"maxAttempts"`
	Timeout     int    `env:"timeout"`
}

func Load() *Config {
	cfg := &Config{}
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return fromEnv(cfg)
	}
	return fromFile(envFile, cfg)
}

// dblayer.env file
func fromFile(path string, cfg *Config) *Config {
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}

// system environment
func fromEnv(cfg *Config) *Config {
	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}
