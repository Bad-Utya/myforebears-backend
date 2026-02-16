package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env                 string                    `yaml:"env" env-required:"true"`
	AccessTokenTTL      time.Duration             `yaml:"access_token_ttl" env-required:"true"`
	RefreshTokenTTL     time.Duration             `yaml:"refresh_token_ttl" env-required:"true"`
	JWTSecret           string                    `yaml:"jwt_secret" env:"JWT_SECRET" env-required:"true"`
	UserStorage         UserStorageConfig         `yaml:"user_storage" env-required:"true"`
	VerificationStorage VerificationStorageConfig `yaml:"verification_storage" env-required:"true"`
	GRPC                GRPCConfig                `yaml:"grpc" env-required:"true"`
}

type UserStorageConfig struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	Username string `yaml:"username" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	Database string `yaml:"database" env-required:"true"`
}

type VerificationStorageConfig struct {
	Address  string `yaml:"address" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	Database int    `yaml:"database" env-required:"true"`
}

type GRPCConfig struct {
	Port    int    `yaml:"port" env-required:"true"`
	Timeout string `yaml:"timeout" env-required:"true"`
}

func MustLoad() *Config {
	// Load .env if present to populate environment variables.
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
