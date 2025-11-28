//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/refynehq/refyne-backend/internal/api"
	"github.com/refynehq/refyne-backend/internal/bootstrap"
	"github.com/refynehq/refyne-backend/internal/config"
	"github.com/refynehq/refyne-backend/internal/database"
	ai "github.com/refynehq/refyne-backend/internal/domains/ai"
	auth "github.com/refynehq/refyne-backend/internal/domains/auth"
	domaincontext "github.com/refynehq/refyne-backend/internal/domains/context"
	email "github.com/refynehq/refyne-backend/internal/domains/email"
	notification "github.com/refynehq/refyne-backend/internal/domains/notification"
	otto "github.com/refynehq/refyne-backend/internal/domains/otto"
	subscription "github.com/refynehq/refyne-backend/internal/domains/subscription"
	user "github.com/refynehq/refyne-backend/internal/domains/user"
	workspace "github.com/refynehq/refyne-backend/internal/domains/workspace"
	handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
	redisPackage "github.com/refynehq/refyne-backend/internal/shared/redis"
	riverqueue "github.com/refynehq/refyne-backend/internal/shared/river"
	"github.com/refynehq/refyne-backend/pkg/logging"
)

var AppSet = wire.NewSet(
	// Core Infrastructure
	config.ProviderSet,
	database.ProviderSet,
	logging.ProviderSet,

	// Shared Services
	redisPackage.ProviderSet,
	riverqueue.ProviderSet,
	handlerregistry.ProviderSet,

	// Domain Layer
	ai.ProviderSet,
	auth.ProviderSet,
	domaincontext.ProviderSet,
	email.ProviderSet,
	notification.ProviderSet,
	otto.ProviderSet,
	subscription.ProviderSet,
	user.ProviderSet,
	workspace.ProviderSet,

	// API Layer
	api.ProviderSet,

	// Application Layer
	bootstrap.ProviderSet,
)

func InitializeApp() (*bootstrap.App, error) {
	wire.Build(AppSet)
	return nil, nil
}
