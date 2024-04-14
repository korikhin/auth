package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf"
	kyaml "github.com/knadh/koanf/parsers/yaml"
	kenv "github.com/knadh/koanf/providers/env"
	kfile "github.com/knadh/koanf/providers/file"
	kstr "github.com/knadh/koanf/providers/structs"
)

type Stage string

type Config struct {
	Stage      Stage `yaml:"-" koanf:"stg"`
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
	MaxAge         int      `yaml:"max-age-seconds" koanf:"max-age-seconds"`
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
	StartTimeout time.Duration `yaml:"start-timeout" koanf:"start-timeout"`
	ReadTimeout  time.Duration `yaml:"read-timeout" koanf:"read-timeout"`
	WriteTimeout time.Duration `yaml:"write-timeout" koanf:"write-timeout"`
	IdleTimeout  time.Duration `yaml:"idle-timeout" koanf:"idle-timeout"`
}

const (
	envPrefix     = "AUTH_SERVER__ENV_PREFIX"
	envStage      = "STG"
	prefixDefault = "AUTH_SERVER__"
	Tag           = "koanf"
)

// Development stages
const (
	Local Stage = "local"
	Dev   Stage = "dev"
	Prod  Stage = "prod"
)

var stages = map[Stage]struct{}{
	Local: {},
	Dev:   {},
	Prod:  {},
}

func MustLoad(path string) *Config {
	prefix := os.Getenv(envPrefix)
	if prefix == "" {
		prefix = prefixDefault
	}
	log.Printf("using env prefix: %s", prefix)

	s := Stage(os.Getenv(fmt.Sprintf("%s%s", prefix, envStage)))
	if _, ok := stages[s]; !ok {
		log.Fatalf(
			"error loading config: please provide stage variable %s%s ('local', 'dev', 'prod')",
			prefix, envStage,
		)
	}

	cfg := defaultConfig()
	k := koanf.New(".")
	if err := k.Load(kstr.Provider(cfg, Tag), nil); err != nil {
		log.Fatalf("error setting default config values: %v", err)
	}

	if s == Local {
		if path == "" {
			log.Fatal("error loading config: please provide config path via --config flag")
		}
		if err := k.Load(kfile.Provider(path), kyaml.Parser()); err != nil {
			log.Fatalf("error loading config: %v", err)
		}
	}

	if err := k.Load(kenv.Provider(prefix, ".", env2Config(prefix)), nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	if err := k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{Tag: Tag}); err != nil {
		log.Fatalf("error unmarshaling config: %v", err)
	}

	return cfg
}

func env2Config(p string) func(string) string {
	return func(s string) string {
		s = strings.TrimPrefix(s, p)
		s = strings.Replace(s, "__", ".", -1)
		s = strings.Replace(s, "_", "-", -1)

		return strings.ToLower(s)
	}
}

func defaultConfig() *Config {
	return &Config{
		CORS: CORS{
			AllowedOrigins: []string{"*"},
			MaxAge:         0,
		},
		HTTPServer: HTTPServer{
			Address:         "0.0.0.0:8080",
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 20 * time.Second,
			HealthTimeout:   15 * time.Minute,
		},
		JWT: JWT{
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 24 * time.Hour,
			Leeway:     0 * time.Second,
		},
		Storage: Storage{
			MinConns:     1,
			MaxConns:     1,
			StartTimeout: 60 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  30 * time.Minute,
		},
	}
}
