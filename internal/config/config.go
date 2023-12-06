package config

import (
	"log"
	"os"
	"time"

	"github.com/studopolis/auth-server/internal/lib/secrets"

	"github.com/ilyakaznacheev/cleanenv"
)

type EnvType string

type Config struct {
	Env        EnvType `yaml:"env" env-default:"local" env-required:"true"`
	HTTPServer `yaml:"http_server"`
	JWT        `yaml:"jwt"`
}

type HTTPServer struct {
	Address         string        `yaml:"address" env-default:"localhost:8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env-default:"5s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env-default:"5s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env-default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"10s"`
}

type JWT struct {
	Key        interface{}   `yaml:"-"`
	Issuer     string        `yaml:"issuer"`
	AccessTTL  time.Duration `yaml:"access_ttl" env-default:"15m"`
	RefreshTTL time.Duration `yaml:"refresh_ttl" env-default:"24h"`
}

const (
	EnvLocal EnvType = "local"
	EnvDev   EnvType = "dev"
	EnvProd  EnvType = "prod"
)

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var config Config
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("failed to read config: %s", err)
	}

	key, err := secrets.GetKey()
	if err != nil {
		log.Fatalf("failed to get encryption key: %s", err)
	}

	config.JWT.Key = key
	return &config
}
