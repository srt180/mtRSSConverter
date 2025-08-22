package config

import (
	"os"
	"strings"

	"gorm.io/gorm"
)

type Config struct {
	DB *gorm.DB

	// db config for sqlite
	SQLitePath string
	// server base address for generating fetch URLs
	BaseAddr string
}

var C = &Config{
	SQLitePath: getEnv("SQLITE_PATH", "./mtRSSConverter.db"),
	BaseAddr:   getEnv("BASE_ADDR", "http://localhost:8080"),
}

func getEnv[T any](key string, defaultValue T) T {
	var zero T

	switch any(zero).(type) {
	case string:
		if value, exists := os.LookupEnv(key); exists {
			return any(value).(T)
		}
		return defaultValue
	case bool:
		if valueSrc, exists := os.LookupEnv(key); exists {
			value := strings.ToLower(valueSrc)

			if value == "false" || value == "0" || value == "no" || value == "off" {
				return any(false).(T)
			}
			return any(false).(T)
		}
		return defaultValue
	default:
		return zero
	}
}
