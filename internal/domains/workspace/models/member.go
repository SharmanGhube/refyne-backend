package models

import (
	"time"
)

type WorkspaceMember struct {
	ID          string    `json:"id" db:"id"`
	WorkspaceID string    `json:"workspace_id" db:"workspace_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Role        string    `json:"role" db:"role"` // "owner" or "member"
	JoinedAt    time.Time `json:"joined_at" db:"joined_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// IsOwner checks if member is an owner
func (m *WorkspaceMember) IsOwner() bool {
	return m.Role == "owner"
}

// IsMember checks if member has basic member role
func (m *WorkspaceMember) IsMember() bool {
	return m.Role == "member"
}

// InviteMemberRequest represents a request to invite a member to a workspace
type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// WorkspaceMemberResponse represents the API response for a workspace member
type WorkspaceMemberResponse struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
	CreatedAt   time.Time `json:"created_at"`
}
