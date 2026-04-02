package workspace

import "time"

// WorkspaceMember represents a member of a workspace
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

// IsMember checks if member is a regular member
func (m *WorkspaceMember) IsMember() bool {
	return m.Role == "member"
}

// InviteMemberRequest for inviting members to workspace
type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// RemoveMemberRequest for removing members from workspace
type RemoveMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// WorkspaceMembersResponse for returning members list
type WorkspaceMembersResponse struct {
	Members []MemberInfo `json:"members"`
	Total   int          `json:"total"`
}

// MemberInfo represents a workspace member in API responses
type MemberInfo struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
	Email     string `json:"email,omitempty"` // User's email if available
	Name      string `json:"name,omitempty"` // User's name if available
}
