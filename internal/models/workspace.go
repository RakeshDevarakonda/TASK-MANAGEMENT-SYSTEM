package models



type CreateWorkspace struct {
	WorkspaceName string `binding:"required"`
	Description   string `binding:"required"`
}

type Workspace struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	UserID      string `db:"user_id"`
	CreatedAt   int64  `db:"created_at"`
	UpdatedAt   int64  `db:"updated_at"`
}

type InviteToWorkspaceRequest struct {
	WorkspaceID string `binding:"required,uuid"`
	Email       string `binding:"required,email"`
	Role        string `binding:"required,oneof=super_admin admin member"`
}

type WorkspaceInvitation struct {
	ID          string `db:"id"`
	WorkspaceID string `db:"workspace_id"`
	Email       string `db:"email"`
	InvitedBy   string `db:"invited_by"`
	Role        string `db:"role"`
	Status      string `db:"status"`
	CreatedAt   int64  `db:"created_at"`
	UpdatedAt   int64  `db:"updated_at"`
}

type JoinWorkspaceRequest struct {
	WorkspaceID string `binding:"required,uuid"`
}

type WorkspaceMember struct {
	ID          string `db:"id"`
	WorkspaceID string `db:"workspace_id"`
	UserID      string `db:"user_id"`
	Role        string `db:"role"`
	CreatedAt   int64  `db:"created_at"`
}

type UpdateWorkspaceRequest struct {
	Name        *string `binding:"omitempty,min=3"`
	Description *string
}

type WorkspaceMemberDetail struct {
	ID          string `db:"id"`
	WorkspaceID string `db:"workspace_id"`
	UserID      string `db:"user_id"`
	Role        string `db:"role"`
	CreatedAt   int64  `db:"created_at"`
	Name        string `db:"name"`
	Email       string `db:"email"`
}

type UpdateMemberRoleRequest struct {
	Role string `binding:"required,oneof=super_admin admin member"`
}

type TransferOwnershipRequest struct {
	NewOwnerID string `binding:"required,uuid"`
}
