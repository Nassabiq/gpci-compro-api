package config

type StorageConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
	BasePath  string
}

func loadStorageConfig() StorageConfig {
	return StorageConfig{
		Endpoint:  getenv("STORAGE_ENDPOINT", "localhost:9000"),
		AccessKey: getenv("STORAGE_ACCESS_KEY", "minioadmin"),
		SecretKey: getenv("STORAGE_SECRET_KEY", "minioadmin"),
		Bucket:    getenv("STORAGE_BUCKET", "uploads"),
		Region:    getenv("STORAGE_REGION", ""),
		UseSSL:    mustBool("STORAGE_USE_SSL", false),
		BasePath:  getenv("STORAGE_BASE_PATH", "uploads"),
	}
}
