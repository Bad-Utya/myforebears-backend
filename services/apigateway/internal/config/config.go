package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env       string        `yaml:"env" env-required:"true"`
	HTTP      HTTPConfig    `yaml:"http" env-required:"true"`
	Clients   ClientsConfig `yaml:"clients" env-required:"true"`
	JWTSecret string        `yaml:"jwt_secret" env:"JWT_SECRET" env-required:"true"`
}

type HTTPConfig struct {
	Port        int           `yaml:"port" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

type ClientsConfig struct {
	Auth         AuthClientConfig   `yaml:"auth" env-required:"true"`
	TokenStorage TokenStorageConfig `yaml:"token_storage" env-required:"true"`
}

type AuthClientConfig struct {
	Address      string        `yaml:"address" env-required:"true"`
	Timeout      time.Duration `yaml:"timeout" env-required:"true"`
	RetriesCount int           `yaml:"retries_count" env-required:"true"`
}

type TokenStorageConfig struct {
	Address  string `yaml:"address" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	Database int    `yaml:"database" env-required:"true"`
}

func MustLoad() *Config {
	_ = godotenv.Load()

	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
