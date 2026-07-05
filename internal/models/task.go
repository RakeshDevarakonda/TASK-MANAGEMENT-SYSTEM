package models

type CreateTaskRequest struct {
	WorkspaceID string `binding:"required,uuid"`
	Title       string `binding:"required,min=3"`
	Description string
	AssignedTo  string `binding:"omitempty,uuid"`
}

type Task struct {
	ID          string  `db:"id"`
	WorkspaceID string  `db:"workspace_id"`
	Title       string  `db:"title"`
	Description *string `db:"description"`
	AssignedTo  *string `db:"assigned_to"`
	CreatedBy   string  `db:"created_by"`
	Status      string  `db:"status"`
	CreatedAt   int64   `db:"created_at"`
	UpdatedAt   int64   `db:"updated_at"`
}

type UpdateTaskRequest struct {
	Title       *string `binding:"omitempty,min=3"`
	Description *string
	AssignedTo  *string `binding:"omitempty,uuid"`
	Status      *string `binding:"omitempty,oneof=todo in_progress submitted completed"`
}

type AssignTaskRequest struct {
	AssignedTo string `binding:"required,uuid"`
}
