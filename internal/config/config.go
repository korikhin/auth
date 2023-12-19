package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type EnvType string

type Config struct {
	Env        EnvType `yaml:"env" env-default:"local" env-required:"true"`
	HTTPServer `yaml:"http_server"`
	JWT        `yaml:"jwt"`
	Storage    `yaml:"storage"`
}

type HTTPServer struct {
	Address         string        `yaml:"address" env-default:"localhost:8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env-default:"5s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env-default:"5s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env-default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"10s"`
	HealthTimeout   time.Duration `yaml:"health_timeout" env-default:"1s"`
}

type JWT struct {
	Issuer     string        `yaml:"issuer"`
	AccessTTL  time.Duration `yaml:"access_ttl" env-default:"15m"`
	RefreshTTL time.Duration `yaml:"refresh_ttl" env-default:"24h"`
}

type Storage struct {
	Alpha Replica `yaml:"alpha"` // master
	Beta  Replica `yaml:"beta"`
	Gamma Replica `yaml:"gamma"`
}

type Replica struct {
	URL             string        `yaml:"url"`
	MinConns        int32         `yaml:"min_conns" env-default:"1"`
	MaxConns        int32         `yaml:"max_conns" env-default:"1"`
	MaxConnIdleTime time.Duration `yaml:"idle_timeout" env-default:"30m"`
}

const (
	EnvLocal EnvType = "local"
	EnvDev   EnvType = "dev"
	EnvProd  EnvType = "prod"
)

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic(fmt.Sprintf("config file does not exist: %s", path))
	}

	var config Config
	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic(fmt.Sprintf("failed to read config: %s", err))
	}

	return &config
}

func fetchConfigPath() string {
	var path string

	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}

	return path
}
