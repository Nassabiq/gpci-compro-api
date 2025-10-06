package config

import "time"

type AuthConfig struct {
	JWTSecret      string
	JWTExpires     time.Duration
	RefreshExpires time.Duration
}

func loadAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:      getenv("JWT_SECRET", "changeme"),
		JWTExpires:     mustDuration("JWT_EXPIRES", "15m"),
		RefreshExpires: mustDuration("REFRESH_EXPIRES", "168h"),
	}
}
