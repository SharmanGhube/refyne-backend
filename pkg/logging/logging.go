package logging

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// safeFileWriter wraps a file and provides thread-safe writing with closed file handling
type safeFileWriter struct {
	file   *os.File
	mutex  sync.RWMutex
	closed bool
}

func newSafeFileWriter(file *os.File) *safeFileWriter {
	return &safeFileWriter{
		file:   file,
		closed: false,
	}
}

func (w *safeFileWriter) Write(p []byte) (n int, err error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if w.closed || w.file == nil {
		// Silently ignore writes to closed files during shutdown
		return len(p), nil
	}

	return w.file.Write(p)
}

func (w *safeFileWriter) Sync() error {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if w.closed || w.file == nil {
		return nil
	}

	return w.file.Sync()
}

func (w *safeFileWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.closed || w.file == nil {
		return nil
	}

	w.closed = true
	err := w.file.Close()
	w.file = nil
	return err
}

// Logging related

var Logger *zap.Logger
var logFile *os.File
var safeWriter *safeFileWriter

type LogConfig struct {
	Level       string `json:"level"`
	Environment string `json:"environment"`
	ServiceName string `json:"service_name"`
	Version     string `json:"version"`
	LogToFile   bool   `json:"log_to_file"`
	LogDir      string `json:"log_directory"`
}

func Initialize() error {
	config := LogConfig{
		Level:       getEnvOrDefault("LOG_LEVEL", "info"),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		ServiceName: getEnvOrDefault("SERVICE_NAME", "refyne-backend"),
		Version:     getEnvOrDefault("VERSION", "1.0.0"),
		LogToFile:   getEnvOrDefault("LOG_TO_FILE", "false") == "true",
		LogDir:      getEnvOrDefault("LOG_DIRECTORY", "logs"),
	}

	// Create logs directory if it doesn't exist
	if config.LogToFile {
		if err := os.MkdirAll(config.LogDir, 0755); err != nil {
			return err
		}
	}

	// Create log file
	if config.LogToFile {
		logFilePath := filepath.Join(config.LogDir, "app.log")
		var err error
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		safeWriter = newSafeFileWriter(logFile)
	}

	// Configure cores for different outputs
	var cores []zapcore.Core

	// Set the log level
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Console core (for development) - only to stdout
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(zapcore.AddSync(os.Stdout)), level)

	// File core (JSON format for production and Promtail/Loki)
	if config.LogToFile {
		fileEncoderConfig := zap.NewProductionEncoderConfig()
		fileEncoderConfig.TimeKey = "@timestamp"
		fileEncoderConfig.MessageKey = "message"
		fileEncoderConfig.LevelKey = "level"
		fileEncoderConfig.CallerKey = "caller"
		fileEncoderConfig.StacktraceKey = "stacktrace"
		fileEncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		fileEncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		fileEncoderConfig.EncodeDuration = zapcore.MillisDurationEncoder
		fileEncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		// Ensure no duplicate fields
		fileEncoderConfig.ConsoleSeparator = ""

		fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
		// Use the safe writer instead of direct file access
		syncedWriter := zapcore.Lock(zapcore.AddSync(safeWriter))
		fileCore := zapcore.NewCore(fileEncoder, syncedWriter, level)
		cores = append(cores, fileCore)
	}

	// Add console core only in development
	if config.Environment == "development" {
		cores = append(cores, consoleCore)
	}

	// Create tee core
	core := zapcore.NewTee(cores...)

	// Build logger
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// Set the global logger
	Logger = logger.With(
		zap.String("service_name", config.ServiceName),
		zap.String("version", config.Version),
		zap.String("environment", config.Environment),
	)

	return nil

}

func NewLogger() *zap.Logger {
	return GetLogger()
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if Logger == nil {
		// Fallback logger if not initialized
		Logger, _ = zap.NewProduction()
	}
	return Logger
}

// Close flushes any buffered log entries and closes log files
func Close() {
	if Logger != nil {
		// Sync the logger to flush any buffered logs
		Logger.Sync()
	}
	if safeWriter != nil {
		// Close the safe writer which will handle the file closure gracefully
		safeWriter.Close()
		safeWriter = nil
	}
	if logFile != nil {
		logFile = nil
	}
}

// GetHandlerLogger returns a logger with the request ID from the context
func GetHandlerLogger(c *gin.Context, handlerName string) *zap.Logger {
	if requestID := c.GetHeader("request_id"); requestID == "" {
		return GetLogger()
	}
	return GetLogger().With(
		zap.String("service_name", "refyne-"),
		zap.String("request_id", c.GetHeader("request_id")),
		zap.String("handler_name", handlerName),
		zap.String("layer", "handler"),
	)
}

// GetServiceLogger returns a logger with the service name
func GetServiceLogger(serviceName string) *zap.Logger {
	return GetLogger().With(
		zap.String("service_name", serviceName),
		zap.String("layer", "service"),
	)
}

// GetJobLogger returns a logger with the job name
func GetJobLogger(jobName string) *zap.Logger {
	return GetLogger().With(
		zap.String("job_name", jobName),
		zap.String("layer", "job"),
	)
}

// GetRepositoryLogger returns a logger with the repository name
func GetRepositoryLogger(repositoryName string) *zap.Logger {
	return GetLogger().With(
		zap.String("repository_name", repositoryName),
		zap.String("layer", "repository"),
	)
}

// GetComponentLogger returns a logger with the component name
func GetComponentLogger(componentName string) *zap.Logger {
	return GetLogger().With(
		zap.String("component_name", componentName),
		zap.String("layer", "component"),
	)
}

// Helpers

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
