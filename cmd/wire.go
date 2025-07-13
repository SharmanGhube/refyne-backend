//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/refynehq/refyne-backend/internal/api"
	"github.com/refynehq/refyne-backend/internal/bootstrap"
	"github.com/refynehq/refyne-backend/internal/config"
	"github.com/refynehq/refyne-backend/internal/database"
	riverqueue "github.com/refynehq/refyne-backend/internal/shared/river"
	"github.com/refynehq/refyne-backend/pkg/logging"
)

var AppSet = wire.NewSet(
	// Core infrastructure
	config.ProviderSet,
	database.ProviderSet,
	logging.ProviderSet,

	// Shared Services
	riverqueue.ProviderSet,

	// Domain layer

	// API Layer
	api.ProviderSet,

	// Application layer
	bootstrap.ProviderSet,
)

func InitializeApp() (*bootstrap.App, error) {
	wire.Build(AppSet)
	return nil, nil
}
