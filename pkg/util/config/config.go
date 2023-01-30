package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	Mysql struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Passwd   string `yaml:"passwd"`
		Database string `yaml:"database"`
	}

	JWT struct {
		Secret     string        `yaml:"secret"`
		Timeout    time.Duration `yaml:"timeout"`
		MaxRefresh time.Duration `yaml:"maxRefresh"`
	}

	Redis struct {
		Endpoint string `yaml:"endpoint"`
		DB       int    `yaml:"db"`
		Passwd   string `yaml:"passwd"`
	}

	RabbitMQ struct {
		Host   string `yaml:"host"`
		Port   string `yaml:"port"`
		User   string `yaml:"user"`
		Passwd string `yaml:"passwd"`
	} `yaml:"rabbitmq"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{}

	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
