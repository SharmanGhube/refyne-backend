package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type Config struct {
	version string

	Environment  string
	Port         string
	FrontendURL  string
	ResendAPIKey string
	SMTP         SMTPConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	Instagram    InstagramConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
	UseSSL   bool
}

type DatabaseConfig struct {
	AutoMigrate bool
	host        string
	port        int
	user        string
	password    string
	database    string
	sslMode     string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type InstagramConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

func NewConfig() (*Config, error) {
	logger := logging.GetComponentLogger("config")

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		env = "development"
	}

	if err := godotenv.Load(".env"); err != nil {
		logger.Warn("Could not load environment file", zap.String("file", ".env"), zap.Error(err))
	}

	autoMigrate := os.Getenv("AUTO_MIGRATE") == "true"
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	config := &Config{
		version:      os.Getenv("APP_VERSION"),
		Environment: env,
		Port:        port,
		FrontendURL: frontendURL,
		ResendAPIKey: os.Getenv("RESEND_API_KEY"),
		SMTP: SMTPConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     587, // Default SMTP port
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			UseTLS:   os.Getenv("SMTP_USE_TLS") == "true",
			UseSSL:   os.Getenv("SMTP_USE_SSL") == "true",
		},
		Database: DatabaseConfig{
			AutoMigrate: autoMigrate,
			host:        os.Getenv("DB_HOST"),
			port:        5432, // Default PostgreSQL port
			user:        os.Getenv("DB_USER"),
			password:    os.Getenv("DB_PASSWORD"),
			database:    os.Getenv("DB_NAME"),
			sslMode:     os.Getenv("DB_SSL_MODE"),
		},
		Redis: RedisConfig{
			Host:     getRedisHost(),
			Port:     getRedisPort(),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0, // Default Redis DB
		},
		Instagram: InstagramConfig{
			ClientID:     os.Getenv("INSTAGRAM_CLIENT_ID"),
			ClientSecret: os.Getenv("INSTAGRAM_CLIENT_SECRET"),
			RedirectURI:  os.Getenv("INSTAGRAM_REDIRECT_URI"),
		},
	}

	logger.Info("Configuration loaded",
		zap.String("environment", config.Environment))

	if config.Environment == "production" {
		logger.Info("Running in production mode")
		validateProductionConfig(logger, config)
	} else {
		logger.Info("Running in development mode")
	}

	return config, nil
}

func getRedisHost() string {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		return "localhost"
	}
	return host
}

func getRedisPort() string {
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		return "6379"
	}
	return port
}

// validateProductionConfig checks for critical production environment variables
func validateProductionConfig(logger *zap.Logger, cfg *Config) {
	missingVars := []string{}

	// Check database configuration
	if cfg.Database.host == "" {
		missingVars = append(missingVars, "DB_HOST")
	}
	if cfg.Database.user == "" {
		missingVars = append(missingVars, "DB_USER")
	}
	if cfg.Database.password == "" {
		missingVars = append(missingVars, "DB_PASSWORD")
	}
	if cfg.Database.database == "" {
		missingVars = append(missingVars, "DB_NAME")
	}

	// Check Redis configuration
	if cfg.Redis.Host == "localhost" {
		logger.Warn("Redis host is set to localhost - this will fail in production containers. Use Railway service reference: ${{Redis.REDIS_HOST}}")
	}

	// Check Resend configuration
	if cfg.ResendAPIKey == "" {
		missingVars = append(missingVars, "RESEND_API_KEY")
	}

	// Check JWT configuration
	if os.Getenv("JWT_SECRET") == "" {
		missingVars = append(missingVars, "JWT_SECRET")
	}

	if len(missingVars) > 0 {
		logger.Warn("Production: Missing critical environment variables",
			zap.Strings("missing_vars", missingVars),
			zap.String("hint", "Use railway.env.template to configure or check Railway service linking"))
	}

	// Log connection info for debugging
	logger.Info("Database connection info",
		zap.String("host", maskSensitive(cfg.Database.host)),
		zap.String("user", cfg.Database.user),
		zap.String("database", cfg.Database.database))

	logger.Info("Redis connection info",
		zap.String("host", maskSensitive(cfg.Redis.Host)),
		zap.String("port", cfg.Redis.Port))
}

// maskSensitive masks password in connection details for logging
func maskSensitive(value string) string {
	if len(value) == 0 {
		return "(empty)"
	}
	return value
}
