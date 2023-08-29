package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

var instance *Config
var once sync.Once

type Logger struct {
	Development bool   `env:"LOGGER_DEVELOPMENT"`
	Level       string `env:"LOGGER_LEVEL"`
}

type Cachce struct {
	CSVExpiration     int `env:"CSV_CACHE_EXPIRATION"`
	SegmentExpiration int `env:"SEGMENT_CACHE_EXPIRATION"`
}

type HTTP struct {
	Host         string `env:"HTTP_HOST"`
	Port         int    `env:"HTTP_PORT"`
	ReadTimeout  int    `env:"HTTP_READ_TIMEOUT"`
	WriteTimeout int    `env:"HTTP_WRITE_TIMEOUT"`
}

type Postgres struct {
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	Database string `env:"POSTGRES_DB"`
	Host     string `env:"POSTGRES_HOST"`
	Port     int    `env:"POSTGRES_PORT"`
	PoolSize int    `env:"POSTGRES_POOL_SIZE"`
	SSLMode  string `env:"POSTGRES_SSL_MODE"`
}

type Config struct {
	Cachce   Cachce
	Logger   Logger
	Postgres Postgres
	HTTP     HTTP
}

func GetConfig() (*Config, error) {
	var cfgErr error
	once.Do(func() {
		instance = &Config{}
		if err := cleanenv.ReadEnv(instance); err != nil {
			cfgErr = err
		}
	})
	return instance, cfgErr
}
