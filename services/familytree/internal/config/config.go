package config

import (
	"os"
	"time"
	"utility/pkg/config"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string             `yaml:"env" env-required:"true"`
	GRPC     GRPCConfig         `yaml:"grpc" env-required:"true"`
	Postgres PostgresConfig     `yaml:"postgres" env-required:"true"`
	Neo4j    Neo4jConfig        `yaml:"neo4j" env-required:"true"`
	Events   EventsClientConfig `yaml:"events" env-required:"true"`
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

type Neo4jConfig struct {
	URI      string `yaml:"uri" env-required:"true"`
	Username string `yaml:"username" env:"NEO4J_USER" env-required:"true"`
	Password string `yaml:"password" env:"NEO4J_PASSWORD" env-required:"true"`
}

type EventsClientConfig struct {
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
