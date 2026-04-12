package handlerregistry

import (
	ai "github.com/refynehq/refyne-backend/internal/domains/ai"
	auth "github.com/refynehq/refyne-backend/internal/domains/auth"
	domaincontext "github.com/refynehq/refyne-backend/internal/domains/context"
	email "github.com/refynehq/refyne-backend/internal/domains/email"
	instagram "github.com/refynehq/refyne-backend/internal/domains/instagram"
	notification "github.com/refynehq/refyne-backend/internal/domains/notification"
	otto "github.com/refynehq/refyne-backend/internal/domains/otto"
	subscription "github.com/refynehq/refyne-backend/internal/domains/subscription"
	user "github.com/refynehq/refyne-backend/internal/domains/user"
	workspace "github.com/refynehq/refyne-backend/internal/domains/workspace"
)

type HandlerRegistry struct {
	Auth         *auth.AuthRegistry
	User         *user.UserRegistry
	AI           *ai.AIRegistry
	Context      *domaincontext.ContextRegistry
	Email        *email.EmailRegistry
	Instagram    *instagram.InstagramRegistry
	Notification *notification.NotificationRegistry
	Otto         *otto.OttoRegistry
	Workspace    *workspace.WorkspaceRegistry
	Subscription *subscription.SubscriptionRegistry
}

func NewHandlerRegistry(
	ar *auth.AuthRegistry,
	ur *user.UserRegistry,
	air *ai.AIRegistry,
	cr *domaincontext.ContextRegistry,
	er *email.EmailRegistry,
	ir *instagram.InstagramRegistry,
	nr *notification.NotificationRegistry,
	or *otto.OttoRegistry,
	wr *workspace.WorkspaceRegistry,
	sr *subscription.SubscriptionRegistry,
) *HandlerRegistry {
	return &HandlerRegistry{
		Auth:         ar,
		User:         ur,
		AI:           air,
		Context:      cr,
		Email:        er,
		Instagram:    ir,
		Notification: nr,
		Otto:         or,
		Workspace:    wr,
		Subscription: sr,
	}
}
