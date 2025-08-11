//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/refynehq/refyne-backend/internal/api"
	"github.com/refynehq/refyne-backend/internal/bootstrap"
	"github.com/refynehq/refyne-backend/internal/config"
	database "github.com/refynehq/refyne-backend/internal/database"
	handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
	riverqueue "github.com/refynehq/refyne-backend/internal/shared/river"
	"github.com/refynehq/refyne-backend/pkg/logging"
)

var AppSet = wire.NewSet(
	// Core Infrastructure
	config.ProviderSet,
	database.ProviderSet,
	logging.ProviderSet,

	// Shared Services
	riverqueue.ProviderSet,
	handlerregistry.ProviderSet,

	// Domain Layer

	// API Layer
	api.ProviderSet,

	// Application Layer
	bootstrap.ProviderSet,
)

func InitializeApp() (*bootstrap.App, error) {
	wire.Build(AppSet)
	return nil, nil
}
