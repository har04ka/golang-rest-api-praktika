package config

import "os"

type Config struct {
	Port          string
	DBUrl         string
	AdminLogin    string
	AdminPassword string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", ":8080"),
		DBUrl:         getEnv("DB_URL", ""),
		AdminLogin:    getEnv("ADMIN_LOGIN", ""),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
