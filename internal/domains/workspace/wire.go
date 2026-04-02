package workspace

import (
	"github.com/google/wire"
	handlerPkg "github.com/refynehq/refyne-backend/internal/domains/workspace/handler"
	repoPkg "github.com/refynehq/refyne-backend/internal/domains/workspace/core/repository"
	servicesPkg "github.com/refynehq/refyne-backend/internal/domains/workspace/services"
)

var ProviderSet = wire.NewSet(
	// Registry
	NewWorkspaceRegistry,

	// Repository
	repoPkg.NewWorkspaceRepository,
	repoPkg.NewWorkspaceMemberRepository,

	// Services
	servicesPkg.NewWorkspaceService,
	servicesPkg.NewMemberService,

	// Handlers
	handlerPkg.NewWorkspaceHandler,
)
