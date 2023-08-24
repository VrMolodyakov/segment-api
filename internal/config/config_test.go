package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigEnvironmentSetCorrectly(t *testing.T) {
	var err error
	err = os.Setenv("LOGGER_DEVELOPMENT", "true")
	assert.NoError(t, err, "unexpected error")
	err = os.Setenv("LOGGER_LEVEL", "debug")
	assert.NoError(t, err, "unexpected error")
	err = os.Setenv("HTTP_HOST", "localhost")
	assert.NoError(t, err, "unexpected error")
	err = os.Setenv("HTTP_PORT", "8080")
	assert.NoError(t, err, "unexpected error")
	err = os.Setenv("HTTP_READ_TIMEOUT", "10")
	assert.NoError(t, err, "unexpected error")
	err = os.Setenv("HTTP_WRITE_TIMEOUT", "20")
	assert.NoError(t, err, "unexpected error")
	err = os.Setenv("POSTGRES_USER", "user")
	assert.NoError(t, err, "unexpected error")
	err = os.Setenv("POSTGRES_PASSWORD", "password")
	assert.NoError(t, err, "unexpected error")
	err = os.Setenv("POSTGRES_DB", "db")
	assert.NoError(t, err, "unexpected error")
	err = os.Setenv("POSTGRES_HOST", "localhost")
	assert.NoError(t, err, "unexpected error")
	err = os.Setenv("POSTGRES_PORT", "5432")
	assert.NoError(t, err, "unexpected error")
	err = os.Setenv("POSTGRES_POOL_SIZE", "10")
	assert.NoError(t, err, "unexpected error")
	err = os.Setenv("POSTGRES_SSL_MODE", "disable")
	assert.NoError(t, err, "unexpected error")

	config, err := GetConfig()
	assert.NoError(t, err, "unexpected error")
	expectedConfig := &Config{
		Logger: Logger{
			Development: true,
			Level:       "debug",
		},
		HTTP: HTTP{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  10,
			WriteTimeout: 20,
		},
		Postgres: Postgres{
			User:     "user",
			Password: "password",
			Database: "db",
			Host:     "localhost",
			Port:     5432,
			PoolSize: 10,
			SSLMode:  "disable",
		},
	}
	assert.Equal(t, expectedConfig, config, "unexpected config")
}
