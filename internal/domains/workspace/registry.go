package workspace

import (
	handlerPkg "github.com/refynehq/refyne-backend/internal/domains/workspace/handler"
)

type WorkspaceRegistry struct {
	handlerPkg.WorkspaceHandler
}

func NewWorkspaceRegistry(handler handlerPkg.WorkspaceHandler) *WorkspaceRegistry {
	return &WorkspaceRegistry{
		WorkspaceHandler: handler,
	}
}
