package config

import "time"

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func loadDBConfig() DBConfig {
	return DBConfig{
		Host:            getenv("DB_HOST", "localhost"),
		Port:            getenv("DB_PORT", "5432"),
		User:            getenv("DB_USER", "postgres"),
		Password:        getenv("DB_PASSWORD", ""),
		Name:            getenv("DB_NAME", "postgres"),
		SSLMode:         getenv("DB_SSLMODE", "disable"),
		MaxOpenConns:    mustInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    mustInt("DB_MAX_IDLE_CONNS", 10),
		ConnMaxLifetime: mustDuration("DB_CONN_MAX_LIFETIME", "30m"),
	}
}
