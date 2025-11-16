package config

type CORSConfig struct {
	AllowOrigins     string
	AllowMethods     string
	AllowHeaders     string
	ExposeHeaders    string
	AllowCredentials bool
	MaxAge           int
}

func loadCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     getenv("CORS_ALLOW_ORIGINS", "*"),
		AllowMethods:     getenv("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"),
		AllowHeaders:     getenv("CORS_ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization"),
		ExposeHeaders:    getenv("CORS_EXPOSE_HEADERS", ""),
		AllowCredentials: mustBool("CORS_ALLOW_CREDENTIALS", false),
		MaxAge:           mustInt("CORS_MAX_AGE", 600),
	}
}
