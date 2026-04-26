package config

import (
	"log"
	"os"
)

type Config struct {
	MySQLDSN  string
	RedisAddr string
	RedisPass string
	JWTSecret string
}

var AppConfig *Config

func LoadConfig() {
	// In a real application, we would use Viper or similar to load from a .yaml file or ENV.
	// For this init, we'll read basic ENV vars with fallbacks.
	AppConfig = &Config{
		MySQLDSN:  getEnv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/gateway?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr: getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPass: getEnv("REDIS_PASS", ""),
		JWTSecret: getEnv("JWT_SECRET", "super-secret-key-123"),
	}
	log.Println("Configuration loaded successfully.")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
