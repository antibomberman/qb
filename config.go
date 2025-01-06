package dblayer

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Driver      string `env:"DRIVER"`
	DSN         string `env:"DSN"`
	MaxAttempts int    `env:"MAX_ATTEMPTS"`
	Timeout     int    `env:"TIMEOUT"`
}

func Load() *Config {
	cfg := &Config{}
	if _, err := os.Stat("dblayer.env"); os.IsNotExist(err) {
		return fromEnv(cfg)
	}
	return fromFile("dblayer.env", cfg)
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
