package config

import "time"

type AppConfig struct {
	Name            string
	Env             string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	CORS            CORSConfig
}

func loadAppConfig() AppConfig {
	return AppConfig{
		Name:            getenv("APP_NAME", "gpci-api"),
		Env:             getenv("APP_ENV", "development"),
		Port:            getenv("APP_PORT", "8080"),
		ReadTimeout:     mustDuration("APP_READ_TIMEOUT", "10s"),
		WriteTimeout:    mustDuration("APP_WRITE_TIMEOUT", "15s"),
		IdleTimeout:     mustDuration("APP_IDLE_TIMEOUT", "60s"),
		ShutdownTimeout: mustDuration("SHUTDOWN_TIMEOUT", "10s"),
		CORS:            loadCORSConfig(),
	}
}
