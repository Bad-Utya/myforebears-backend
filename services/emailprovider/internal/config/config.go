package config

import (
	"os"
	"utility/pkg/config"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string         `yaml:"env" env-required:"true"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq" env-required:"true"`
	SMTP     SMTPConfig     `yaml:"smtp" env-required:"true"`
}

type SMTPConfig struct {
	Host     string `yaml:"host" env:"SMTP_HOST" env-required:"true"`
	Port     int    `yaml:"port" env:"SMTP_PORT" env-required:"true"`
	Username string `yaml:"username" env:"SMTP_USERNAME" env-required:"true"`
	Password string `yaml:"password" env:"SMTP_PASSWORD" env-required:"true"`
	From     string `yaml:"from" env:"SMTP_FROM" env-required:"true"`
}

type RabbitMQConfig struct {
	URL         string `yaml:"url" env:"RABBITMQ_URL" env-required:"true"`
	Exchange    string `yaml:"exchange" env-required:"true"`
	Queue       string `yaml:"queue" env-required:"true"`
	RoutingKey  string `yaml:"routing_key" env-required:"true"`
	ConsumerTag string `yaml:"consumer_tag" env-required:"true"`
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
