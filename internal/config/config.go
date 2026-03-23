package config

import "os"

type Config struct {
	Port      string
	RedisURL  string
	JWTSecret string
}


func Load() (*Config, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "mysecretkey"
	}

	return &Config{
		Port:      getEnv("PORT", "8080"),
		RedisURL:  redisURL,
		JWTSecret: secret,
	}, nil
}


func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}