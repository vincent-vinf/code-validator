package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	JWT      JWT      `yaml:"jwt"`
	RabbitMQ RabbitMQ `yaml:"rabbitmq"`
	Minio    Minio    `yaml:"minio"`
	Mysql    Mysql    `yaml:"mysql"`
}

type Mysql struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Passwd   string `yaml:"passwd"`
	Database string `yaml:"database"`
}

type RabbitMQ struct {
	Host   string `yaml:"host"`
	Port   string `yaml:"port"`
	User   string `yaml:"user"`
	Passwd string `yaml:"passwd"`
}

type JWT struct {
	Secret     string        `yaml:"secret"`
	Timeout    time.Duration `yaml:"timeout"`
	MaxRefresh time.Duration `yaml:"maxRefresh"`
}

type Minio struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	Bucket          string `yaml:"bucket"`
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
