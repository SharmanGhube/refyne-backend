//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/refynehq/refyne-backend/internal/api"
	"github.com/refynehq/refyne-backend/internal/bootstrap"
	"github.com/refynehq/refyne-backend/internal/config"
	"github.com/refynehq/refyne-backend/internal/database"
	"github.com/refynehq/refyne-backend/internal/domain/auth"
	"github.com/refynehq/refyne-backend/internal/domain/user"
	registry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
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
	registry.ProviderSet,

	// Domain layer
	user.ProviderSet,
	auth.ProviderSet,

	// API Layer
	api.ProviderSet,

	// Application layer
	bootstrap.ProviderSet,
)

func InitializeApp() (*bootstrap.App, error) {
	wire.Build(AppSet)
	return nil, nil
}
