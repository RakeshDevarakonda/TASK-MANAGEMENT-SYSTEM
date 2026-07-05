package repository

import (
	"fmt"
	"task-system/internal/models"
	"task-system/internal/utils"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var ErrDuplicateInvitation = utils.ErrConflict

type WorkspaceRepository struct {
	db *sqlx.DB
}

func NewWorkspaceRepository(db *sqlx.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

// GetByID retrieves a workspace by its unique ID
func (r *WorkspaceRepository) GetByID(id string) (*models.Workspace, error) {
	var workspace models.Workspace
	query := "SELECT id, name, description, user_id, created_at, updated_at FROM workspaces WHERE id = $1"
	err := r.db.Get(&workspace, query, id)
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

// CreateWorkspace inserts a new organization/workspace record into the database
func (r *WorkspaceRepository) CreateWorkspace(workspaceName string, description string, userId string) (*models.Workspace, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	workspace := &models.Workspace{
		Name:        workspaceName,
		Description: description,
		UserID:      userId,
	}

	query := `
		INSERT INTO workspaces (name, description, user_id)
		VALUES (:name, :description, :user_id)
		RETURNING id, created_at, updated_at
	`
	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.Get(workspace, workspace)
	if err != nil {
		return nil, err
	}

	// Also insert owner into workspace_members
	memberQuery := `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'super_admin')
	`
	_, err = tx.Exec(memberQuery, workspace.ID, userId)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return workspace, nil
}

// CreateInvitation inserts a new workspace invitation into the database
func (r *WorkspaceRepository) CreateInvitation(invitation *models.WorkspaceInvitation) error {
	query := `
		INSERT INTO workspace_invitations (workspace_id, email, invited_by, role)
		VALUES (:workspace_id, :email, :invited_by, :role)
		RETURNING id, status, role, created_at, updated_at
	`
	rows, err := r.db.NamedQuery(query, invitation)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrDuplicateInvitation
		}
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(invitation)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetPendingInvitation retrieves a pending invitation matching a workspace ID and user email
func (r *WorkspaceRepository) GetPendingInvitation(workspaceID string, email string) (*models.WorkspaceInvitation, error) {
	var invitation models.WorkspaceInvitation
	query := `
		SELECT id, workspace_id, email, invited_by, role, status, created_at, updated_at 
		FROM workspace_invitations 
		WHERE workspace_id = $1 AND email = $2 AND status = 'pending'
	`
	err := r.db.Get(&invitation, query, workspaceID, email)
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

// AcceptInvitation updates an invitation status to 'accepted' and inserts a membership record atomically
func (r *WorkspaceRepository) AcceptInvitation(invitationID string, userID string, workspaceID string, role string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update invitation status to 'accepted'
	updateQuery := "UPDATE workspace_invitations SET status = 'accepted', updated_at = extract(epoch from now())::bigint WHERE id = $1"
	_, err = tx.Exec(updateQuery, invitationID)
	if err != nil {
		return err
	}

	// 2. Insert member record linking the user to the workspace with the specific role
	insertQuery := "INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING"
	_, err = tx.Exec(insertQuery, workspaceID, userID, role)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// IsOwnerOrAdmin checks if a user is either the workspace owner/creator or has the admin/super_admin role
func (r *WorkspaceRepository) IsOwnerOrAdmin(workspaceID string, userID string) (bool, error) {
	// 1. Check if the user is the workspace owner
	var isOwner bool
	ownerQuery := "SELECT EXISTS(SELECT 1 FROM workspaces WHERE id = $1 AND user_id = $2)"
	err := r.db.Get(&isOwner, ownerQuery, workspaceID, userID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}

	// 2. Check if the user is an admin or super_admin of the workspace
	var isAdmin bool
	adminQuery := "SELECT EXISTS(SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2 AND role IN ('admin', 'super_admin'))"
	err = r.db.Get(&isAdmin, adminQuery, workspaceID, userID)
	if err != nil {
		return false, err
	}
	return isAdmin, nil
}

// IsMemberOrOwner checks if a user is either the workspace owner or is a registered member of the workspace
func (r *WorkspaceRepository) IsMemberOrOwner(workspaceID string, userID string) (bool, error) {
	// 1. Check if the user is the workspace owner
	var isOwner bool
	ownerQuery := "SELECT EXISTS(SELECT 1 FROM workspaces WHERE id = $1 AND user_id = $2)"
	err := r.db.Get(&isOwner, ownerQuery, workspaceID, userID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}

	// 2. Check if the user is registered as a member in the workspace
	var isMember bool
	memberQuery := "SELECT EXISTS(SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2)"
	err = r.db.Get(&isMember, memberQuery, workspaceID, userID)
	if err != nil {
		return false, err
	}
	return isMember, nil
}

// ListByUserID retrieves all workspaces the user owns or belongs to as a member
func (r *WorkspaceRepository) ListByUserID(userID string) ([]models.Workspace, error) {
	var workspaces []models.Workspace
	query := `
		SELECT DISTINCT w.id, w.name, w.description, w.user_id, w.created_at, w.updated_at 
		FROM workspaces w
		LEFT JOIN workspace_members wm ON w.id = wm.workspace_id
		WHERE w.user_id = $1 OR wm.user_id = $1
		ORDER BY w.created_at DESC
	`
	err := r.db.Select(&workspaces, query, userID)
	if err != nil {
		return nil, err
	}
	return workspaces, nil
}

// UpdateWorkspace updates a workspace's properties in the database
func (r *WorkspaceRepository) UpdateWorkspace(workspace *models.Workspace) error {
	query := `
		UPDATE workspaces 
		SET name = :name, description = :description, updated_at = extract(epoch from now())::bigint
		WHERE id = :id
	`
	_, err := r.db.NamedExec(query, workspace)
	return err
}

// DeleteWorkspace deletes a workspace by its unique ID
func (r *WorkspaceRepository) DeleteWorkspace(id string) error {
	query := "DELETE FROM workspaces WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}

// ListInvitationsByEmail retrieves all pending invitations matching a user's email address
func (r *WorkspaceRepository) ListInvitationsByEmail(email string) ([]models.WorkspaceInvitation, error) {
	var invitations []models.WorkspaceInvitation
	query := `
		SELECT id, workspace_id, email, invited_by, role, status, created_at, updated_at 
		FROM workspace_invitations 
		WHERE email = $1 AND status = 'pending' 
		ORDER BY created_at DESC
	`
	err := r.db.Select(&invitations, query, email)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// ListMembers lists all registered users in a workspace with their profiles and roles
func (r *WorkspaceRepository) ListMembers(workspaceID string) ([]models.WorkspaceMemberDetail, error) {
	var members []models.WorkspaceMemberDetail
	query := `
		SELECT wm.id, wm.workspace_id, wm.user_id, wm.role, wm.created_at, u.name, u.email 
		FROM workspace_members wm 
		JOIN users u ON wm.user_id = u.id 
		WHERE wm.workspace_id = $1 
		ORDER BY wm.created_at ASC
	`
	err := r.db.Select(&members, query, workspaceID)
	if err != nil {
		return nil, err
	}
	return members, nil
}

// UpdateMemberRole alters a workspace user's role
func (r *WorkspaceRepository) UpdateMemberRole(workspaceID string, userID string, role string) error {
	query := "UPDATE workspace_members SET role = $1 WHERE workspace_id = $2 AND user_id = $3"
	_, err := r.db.Exec(query, role, workspaceID, userID)
	return err
}

// DeleteMember deletes a workspace member linking
func (r *WorkspaceRepository) DeleteMember(workspaceID string, userID string) error {
	query := "DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2"
	_, err := r.db.Exec(query, workspaceID, userID)
	return err
}

// GetUserRole returns a user's role inside a workspace ('super_admin', 'admin', 'member')
func (r *WorkspaceRepository) GetUserRole(workspaceID string, userID string) (string, error) {
	// 1. Check if the user is the workspace owner
	var isOwner bool
	ownerQuery := "SELECT EXISTS(SELECT 1 FROM workspaces WHERE id = $1 AND user_id = $2)"
	err := r.db.Get(&isOwner, ownerQuery, workspaceID, userID)
	if err != nil {
		return "", err
	}
	if isOwner {
		return "super_admin", nil
	}

	// 2. Otherwise, retrieve role from workspace_members
	var role string
	memberQuery := "SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2"
	err = r.db.Get(&role, memberQuery, workspaceID, userID)
	if err != nil {
		return "", err
	}
	return role, nil
}

// TransferWorkspaceOwnership transfers workspace owner ID and alters memberships atomically
func (r *WorkspaceRepository) TransferWorkspaceOwnership(workspaceID string, oldOwnerID string, newOwnerID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update owner ID in workspaces table
	ownerQuery := "UPDATE workspaces SET user_id = $1, updated_at = extract(epoch from now())::bigint WHERE id = $2"
	_, err = tx.Exec(ownerQuery, newOwnerID, workspaceID)
	if err != nil {
		return err
	}

	// 2. Set old owner's role to 'admin' in workspace_members
	oldOwnerQuery := `
		INSERT INTO workspace_members (workspace_id, user_id, role) 
		VALUES ($1, $2, 'admin') 
		ON CONFLICT (workspace_id, user_id) 
		DO UPDATE SET role = 'admin'
	`
	_, err = tx.Exec(oldOwnerQuery, workspaceID, oldOwnerID)
	if err != nil {
		return err
	}

	// 3. Set new owner's role to 'super_admin' in workspace_members
	newOwnerQuery := `
		INSERT INTO workspace_members (workspace_id, user_id, role) 
		VALUES ($1, $2, 'super_admin') 
		ON CONFLICT (workspace_id, user_id) 
		DO UPDATE SET role = 'super_admin'
	`
	_, err = tx.Exec(newOwnerQuery, workspaceID, newOwnerID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ListWorkspaceMembersCursor retrieves a paginated slice of workspace members using keyset pagination
func (r *WorkspaceRepository) ListWorkspaceMembersCursor(workspaceID string, limit int, cursorTime int64, cursorID string) ([]models.WorkspaceMemberDetail, error) {
	queryLimit := limit + 1

	baseQuery := "WHERE wm.workspace_id = $1"
	var args []interface{}
	args = append(args, workspaceID)

	if cursorTime > 0 && cursorID != "" {
		baseQuery += " AND (wm.created_at > $2 OR (wm.created_at = $2 AND wm.id > $3))"
		args = append(args, cursorTime, cursorID)
	}

	query := fmt.Sprintf(`
		SELECT wm.id, wm.workspace_id, wm.user_id, wm.role, wm.created_at, u.name, u.email 
		FROM workspace_members wm 
		JOIN users u ON wm.user_id = u.id 
		%s 
		ORDER BY wm.created_at ASC, wm.id ASC 
		LIMIT $%d
	`, baseQuery, len(args)+1)

	args = append(args, queryLimit)

	var members []models.WorkspaceMemberDetail
	err := r.db.Select(&members, query, args...)
	if err != nil {
		return nil, err
	}

	return members, nil
}