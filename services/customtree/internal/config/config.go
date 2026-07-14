package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"os"
	shared "utility/pkg/config"
)

type Config struct {
	Env  string `yaml:"env" env-required:"true"`
	GRPC struct {
		Port int `yaml:"port" env-required:"true"`
	} `yaml:"grpc"`
	Postgres PostgresConfig `yaml:"postgres"`
	S3       S3Config       `yaml:"s3"`
}
type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username" env:"POSTGRES_USER"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
	Database string `yaml:"database"`
}
type S3Config struct {
	Endpoint        string `yaml:"endpoint" env:"S3_ENDPOINT"`
	Region          string `yaml:"region" env:"S3_REGION"`
	Bucket          string `yaml:"bucket" env:"S3_BUCKET"`
	AccessKeyID     string `yaml:"access_key_id" env:"S3_ACCESS_KEY_ID"`
	SecretAccessKey string `yaml:"secret_access_key" env:"S3_SECRET_ACCESS_KEY"`
	ForcePathStyle  bool   `yaml:"force_path_style" env:"S3_FORCE_PATH_STYLE"`
}

func MustLoad() *Config {
	_ = godotenv.Load()
	p := shared.FetchConfigPath()
	if p == "" {
		panic("config path is empty")
	}
	if _, err := os.Stat(p); err != nil {
		panic(err)
	}
	var c Config
	if err := cleanenv.ReadConfig(p, &c); err != nil {
		panic(err)
	}
	return &c
}
