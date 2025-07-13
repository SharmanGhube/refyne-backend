package config

import (
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type Config struct {
	Environment     string
	Port            string
	AutoMigrate     bool
	SMTPConfig      SMTPConfig
	InstagramConfig InstagramConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
	UseSSL   bool
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

	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	autoMigrate := os.Getenv("AUTO_MIGRATE") == "true"
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))

	config := &Config{
		Environment: env,
		Port:        port,
		AutoMigrate: autoMigrate,
		SMTPConfig: SMTPConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     smtpPort,
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			UseTLS:   os.Getenv("SMTP_USE_TLS") == "true",
			UseSSL:   os.Getenv("SMTP_USE_SSL") == "true",
		},
		InstagramConfig: InstagramConfig{
			ClientID:     os.Getenv("INSTAGRAM_CLIENT_ID"),
			ClientSecret: os.Getenv("INSTAGRAM_CLIENT_SECRET"),
			RedirectURI:  os.Getenv("INSTAGRAM_REDIRECT_URI"),
		},
	}

	logger.Info("Configuration loaded")

	return config, nil

}
