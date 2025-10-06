package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App     AppConfig
	DB      DBConfig
	Redis   RedisConfig
	Asynq   AsynqConfig
	Auth    AuthConfig
	Storage StorageConfig
}

func mustDuration(key, def string) time.Duration {
	s := getenv(key, def)
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Fatalf("invalid duration for %s: %v", key, err)
	}
	return d
}

func mustInt(key string, def int) int {
	v := getenv(key, "")
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("invalid int for %s: %v", key, err)
	}
	return i
}

func mustBool(key string, def bool) bool {
	v := getenv(key, "")
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		log.Fatalf("invalid bool for %s: %v", key, err)
	}
	return b
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		App:     loadAppConfig(),
		DB:      loadDBConfig(),
		Redis:   loadRedisConfig(),
		Asynq:   loadAsynqConfig(),
		Auth:    loadAuthConfig(),
		Storage: loadStorageConfig(),
	}
}
