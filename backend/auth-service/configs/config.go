package configs

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
}

type ServerConfig struct {
	Port string
	Host string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

type OAuthConfig struct {
	Google   OAuthProvider
	Facebook OAuthProvider
}

type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func LoadConfig() *Config {
	// JWT expires in parsing
	jwtExpiresIn, _ := time.ParseDuration(getEnv("JWT_EXPIRES_IN", "24h"))

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "3001"),
			Host: getEnv("HOST", "localhost"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "threads_db"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres123"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			ExpiresIn: jwtExpiresIn,
		},
		OAuth: OAuthConfig{
			Google: OAuthProvider{
				ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:3001/auth/google/callback"),
			},
			Facebook: OAuthProvider{
				ClientID:     getEnv("FACEBOOK_CLIENT_ID", ""),
				ClientSecret: getEnv("FACEBOOK_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("FACEBOOK_REDIRECT_URL", "http://localhost:3001/auth/facebook/callback"),
			},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}
