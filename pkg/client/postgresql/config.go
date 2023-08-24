package postgresql

import "fmt"

type pgConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	Database string
	PoolSize int
	SslMode  string
}

func NewPgConfig(Username string, password string, host string, port int, database string, poolSize int, SSLMode string) pgConfig {
	return pgConfig{
		Username: Username,
		Password: password,
		Host:     host,
		Port:     port,
		Database: database,
		PoolSize: poolSize,
		SslMode:  SSLMode,
	}
}

func (pg pgConfig) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		pg.Username,
		pg.Password,
		pg.Host,
		pg.Port,
		pg.Database,
		pg.SslMode,
	)
}
