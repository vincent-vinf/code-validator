package config

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/vincent-vinf/code-validator/pkg/util/log"
)

var (
	config = &Config{}
	once   sync.Once
	rwLock sync.RWMutex

	logger = log.GetLogger()
)

func Init(configPath string) {
	once.Do(func() {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			logger.Fatalln(err)
		}
		if err := viper.Unmarshal(config); err != nil {
			logger.Fatalln(err)
		}
		viper.WatchConfig()
		viper.OnConfigChange(func(in fsnotify.Event) {
			rwLock.Lock()
			if err := viper.Unmarshal(config); err != nil {
				logger.Errorf("unmarshal conf failed, err:%s", err)
			}
			rwLock.Unlock()
			logger.Info("config reloaded")
		})
	})
}

type Config struct {
	Mysql struct {
		Host     string
		Port     string
		User     string
		Passwd   string
		Database string
	}

	JWT struct {
		Secret     string
		Timeout    time.Duration
		MaxRefresh time.Duration
	}

	AdminJWT struct {
		Secret     string
		Timeout    time.Duration
		MaxRefresh time.Duration
	} `mapstructure:"admin_jwt"`

	Redis struct {
		Endpoint string
		DB       int
		Passwd   string
	}

	RabbitMQ struct {
		Host   string
		Port   string
		User   string
		Passwd string
	} `mapstructure:"rabbitmq"`

	Spike struct {
		RandUrlKey string `mapstructure:"rand_url_key"`
	}
}

func GetConfig() Config {
	rwLock.RLock()
	defer rwLock.RUnlock()

	return *config
}
