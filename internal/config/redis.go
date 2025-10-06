package config

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func loadRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:     getenv("REDIS_ADDR", "localhost:6379"),
		Password: getenv("REDIS_PASSWORD", ""),
		DB:       mustInt("REDIS_DB", 0),
	}
}
