package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv               string
	APIPort              string
	JWTSecret            string
	WebhookSigningSecret string
	InternalSharedToken  string
	AdminUsername        string
	AdminPassword        string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisOptional bool

	WorkerConcurrency int
}

func Load() Config {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	workerConcurrency, _ := strconv.Atoi(getEnv("WORKER_CONCURRENCY", "10"))
	redisOptional, _ := strconv.ParseBool(getEnv("REDIS_OPTIONAL", "true"))

	return Config{
		AppEnv:               getEnv("APP_ENV", "development"),
		APIPort:              getEnv("API_PORT", "8080"),
		JWTSecret:            getEnv("JWT_SECRET", "change-me"),
		WebhookSigningSecret: getEnv("WEBHOOK_SIGNING_SECRET", "change-me"),
		InternalSharedToken:  getEnv("INTERNAL_SHARED_TOKEN", "change-me"),
		AdminUsername:        getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:        getEnv("ADMIN_PASSWORD", "admin123"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "3306"),
		DBUser:               getEnv("DB_USER", "waapp"),
		DBPassword:           getEnv("DB_PASSWORD", "waapp123"),
		DBName:               getEnv("DB_NAME", "wa_platform"),
		RedisAddr:            getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
		RedisDB:              redisDB,
		RedisOptional:        redisOptional,
		WorkerConcurrency:    workerConcurrency,
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
