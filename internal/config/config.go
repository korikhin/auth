package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
)

type EnvType string

type Config struct {
	Env        EnvType `yaml:"env" koanf:"env"`
	CORS       `yaml:"cors" koanf:"cors"`
	HTTPServer `yaml:"http-server" koanf:"http-server"`
	JWT        `yaml:"jwt" koanf:"jwt"`
	Storage    `yaml:"storage" koanf:"storage"`
}

type HTTPServer struct {
	Address         string        `yaml:"address" koanf:"address"`
	ReadTimeout     time.Duration `yaml:"read-timeout" koanf:"read-timeout"`
	WriteTimeout    time.Duration `yaml:"write-timeout" koanf:"write-timeout"`
	IdleTimeout     time.Duration `yaml:"idle-timeout" koanf:"idle-timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown-timeout" koanf:"shutdown-timeout"`
	HealthTimeout   time.Duration `yaml:"health-timeout" koanf:"health-timeout"`
}

type CORS struct {
	AllowedOrigins []string `yaml:"allowed-origins" koanf:"allowed-origins"`
	MaxAge         int      `yaml:"max-age" koanf:"max-age"`
}

type JWT struct {
	Issuer     string        `yaml:"issuer" koanf:"issuer"`
	AccessTTL  time.Duration `yaml:"access-ttl" koanf:"access-ttl"`
	RefreshTTL time.Duration `yaml:"refresh-ttl" koanf:"refresh-ttl"`
	Leeway     time.Duration `yaml:"leeway" koanf:"leeway"`
}

type Storage struct {
	URL          string        `yaml:"url" koanf:"url"`
	MinConns     int32         `yaml:"min-conns" koanf:"min-conns"`
	MaxConns     int32         `yaml:"max-conns" koanf:"max-conns"`
	ReadTimeout  time.Duration `yaml:"read-timeout" koanf:"read-timeout"`
	WriteTimeout time.Duration `yaml:"write-timeout" koanf:"write-timeout"`
	IdleTimeout  time.Duration `yaml:"idle-timeout" koanf:"idle-timeout"`
}

// Developing environment
const (
	Local EnvType = "local"
	Dev   EnvType = "dev"
	Prod  EnvType = "prod"
)

const (
	DefaultConfigPath = "config/local.yaml"
	DefultEnvPrefix   = "STD_AUTH__"
	EnvConfigPath     = "CONFIG_PATH"
	TagKoanf          = "koanf"
)

func MustLoad(path string) *Config {
	k := koanf.New(".")

	if path == "" {
		path = os.Getenv(fmt.Sprintf("%s%s", DefultEnvPrefix, EnvConfigPath))
	}

	var isDefault bool
	if path == "" {
		path = DefaultConfigPath
		isDefault = true
	}

	cfg := Default()
	if err := k.Load(structs.Provider(cfg, TagKoanf), nil); err != nil {
		log.Fatalf("error setting default config values: %v", err)
	}

	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		if isDefault {
			log.Fatalf("error loading default config: %v", err)
		}
		log.Fatalf("error loading config: %v", err)
	}

	if err := k.Load(env.Provider(DefultEnvPrefix, ".", ParseEnvVariable), nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	if err := k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{Tag: TagKoanf}); err != nil {
		log.Fatalf("error unmarshaling config: %v", err)
	}

	return cfg
}

func ParseEnvVariable(s string) string {
	s = strings.TrimPrefix(s, DefultEnvPrefix)
	s = strings.Replace(s, "__", ".", -1)
	s = strings.Replace(s, "_", "-", -1)
	return strings.ToLower(s)
}

func Default() *Config {
	return &Config{
		Env: Local,
		HTTPServer: HTTPServer{
			Address:         "localhost:8080",
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			HealthTimeout:   1 * time.Second,
		},
		JWT: JWT{
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 24 * time.Hour,
			Leeway:     0 * time.Second,
		},
		Storage: Storage{
			MinConns:     1,
			MaxConns:     1,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  30 * time.Minute,
		},
	}
}
