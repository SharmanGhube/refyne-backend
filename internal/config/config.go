package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type Config struct {
	version string

	Environment string
	Port        string
	SMTP        SMTPConfig
	Database    DatabaseConfig
	Instagram   InstagramConfig
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

type InstagramConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

func NewConfig() (*Config, error) {
	logger := logging.GetComponentLogger("config")

	env := os.Getenv("ENVIRONMENT")
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

	config := &Config{
		version:     os.Getenv("APP_VERSION"),
		Environment: env,
		Port:        port,
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
	} else {
		logger.Info("Running in development mode")
	}

	return config, nil

}
