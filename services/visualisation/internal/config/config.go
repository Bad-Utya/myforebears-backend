package config

import (
	"os"
	"time"
	"utility/pkg/config"

	"github.com/joho/godotenv"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string                 `yaml:"env" env-required:"true"`
	GRPC       GRPCConfig             `yaml:"grpc" env-required:"true"`
	Postgres   PostgresConfig         `yaml:"postgres" env-required:"true"`
	S3         S3Config               `yaml:"s3" env-required:"true"`
	FamilyTree FamilyTreeClientConfig `yaml:"familytree" env-required:"true"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-required:"true"`
}

type PostgresConfig struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	Username string `yaml:"username" env:"POSTGRES_USER" env-required:"true"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-required:"true"`
	Database string `yaml:"database" env-required:"true"`
}

type S3Config struct {
	Endpoint        string `yaml:"endpoint" env:"S3_ENDPOINT" env-required:"true"`
	Region          string `yaml:"region" env:"S3_REGION" env-required:"true"`
	Bucket          string `yaml:"bucket" env:"S3_VISUALISATIONS_BUCKET" env-required:"true"`
	AccessKeyID     string `yaml:"access_key_id" env:"S3_ACCESS_KEY_ID" env-required:"true"`
	SecretAccessKey string `yaml:"secret_access_key" env:"S3_SECRET_ACCESS_KEY" env-required:"true"`
	ForcePathStyle  bool   `yaml:"force_path_style" env:"S3_FORCE_PATH_STYLE" env-required:"true"`
}

type FamilyTreeClientConfig struct {
	Address      string        `yaml:"address" env-required:"true"`
	Timeout      time.Duration `yaml:"timeout" env-required:"true"`
	RetriesCount int           `yaml:"retries_count" env-required:"true"`
}

func MustLoad() *Config {
	_ = godotenv.Load()

	path := config.FetchConfigPath()
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
