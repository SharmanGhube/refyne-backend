package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

func main() {
	// Initialize logging first
	if err := logging.Initialize(); err != nil {
		log.Fatalf("Failed to initialize logging: %v", err)
	}
	defer logging.Close()

	logger := logging.GetLogger()

	// Wire does all the heavy lifting - no more AppContext!
	app, err := InitializeApp()
	if err != nil {
		logger.Fatal("Failed to initialize app", zap.Error(err))
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the app
	if err := app.Start(ctx); err != nil {
		logger.Fatal("Failed to start app", zap.Error(err))
	}

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Received shutdown signal")
	cancel()

	// Graceful shutdown
	if err := app.Stop(context.Background()); err != nil {
		logger.Error("Failed to stop app gracefully", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Application shut down successfully")
}
