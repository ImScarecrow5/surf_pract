package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port     string
	NodeEnv  string
	DB       DatabaseConfig
	JWT      JWTConfig
	Booking  BookingConfig
	SMS      SMSConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

type JWTConfig struct {
	Secret          string
	RefreshSecret   string
	ExpiresIn       string
	RefreshExpiresIn string
}

type BookingConfig struct {
	ConfirmationDeadlineMinutes int
	FreeCancellationHours       int
	LateCancellationPenaltyPercent int
}

type SMSConfig struct {
	CodeExpires int
}

func Load() *Config {
	return &Config{
		Port:    getEnv("PORT", "3000"),
		NodeEnv: getEnv("NODE_ENV", "development"),
		DB: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			Database: getEnv("DB_NAME", "climbing"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "secret"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			RefreshSecret:   getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key-change-in-production"),
			ExpiresIn:       getEnv("JWT_EXPIRES_IN", "10m"),
			RefreshExpiresIn: getEnv("JWT_REFRESH_EXPIRES_IN", "7d"),
		},
		Booking: BookingConfig{
			ConfirmationDeadlineMinutes: 15,
			FreeCancellationHours:       2,
			LateCancellationPenaltyPercent: 10,
		},
		SMS: SMSConfig{
			CodeExpires: getEnvAsInt("SMS_CODE_EXPIRES", 300),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}