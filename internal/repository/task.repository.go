package repository

import (
	"task-system/internal/models"

	"github.com/jmoiron/sqlx"
)

type TaskRepository struct {
	db *sqlx.DB
}

func NewTaskRepository(db *sqlx.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// CreateTask inserts a new task record into the PostgreSQL database
func (r *TaskRepository) CreateTask(task *models.Task) error {
	query := `
		INSERT INTO tasks (workspace_id, title, description, assigned_to, created_by, status)
		VALUES (:workspace_id, :title, :description, :assigned_to, :created_by, :status)
		RETURNING id, created_at, updated_at
	`
	rows, err := r.db.NamedQuery(query, task)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(task)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetByID fetches a specific task by its unique ID
func (r *TaskRepository) GetByID(id string) (*models.Task, error) {
	var task models.Task
	query := "SELECT id, workspace_id, title, description, assigned_to, created_by, status, created_at, updated_at FROM tasks WHERE id = $1"
	err := r.db.Get(&task, query, id)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// ListByWorkspaceID retrieves all tasks in a workspace
func (r *TaskRepository) ListByWorkspaceID(workspaceID string) ([]models.Task, error) {
	var tasks []models.Task
	query := "SELECT id, workspace_id, title, description, assigned_to, created_by, status, created_at, updated_at FROM tasks WHERE workspace_id = $1 ORDER BY created_at DESC"
	err := r.db.Select(&tasks, query, workspaceID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// UpdateTask updates the task properties in PostgreSQL
func (r *TaskRepository) UpdateTask(task *models.Task) error {
	query := `
		UPDATE tasks 
		SET title = :title, description = :description, assigned_to = :assigned_to, status = :status, updated_at = extract(epoch from now())::bigint
		WHERE id = :id
	`
	_, err := r.db.NamedExec(query, task)
	return err
}

// DeleteTask removes a task record by its ID
func (r *TaskRepository) DeleteTask(id string) error {
	query := "DELETE FROM tasks WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}

// ListWorkspaceTasksCursor retrieves tasks from workspace using keyset pagination
func (r *TaskRepository) ListWorkspaceTasksCursor(workspaceID string, limit int, cursorTime int64, cursorID string, status []string, assigneeID string) ([]models.Task, error) {
	queryLimit := limit + 1

	baseQuery := "WHERE workspace_id = ?"
	var args []interface{}
	args = append(args, workspaceID)

	if cursorTime > 0 && cursorID != "" {
		baseQuery += " AND (created_at < ? OR (created_at = ? AND id < ?))"
		args = append(args, cursorTime, cursorTime, cursorID)
	}

	if len(status) > 0 {
		baseQuery += " AND status IN (?)"
		args = append(args, status)
	}

	if assigneeID != "" {
		if assigneeID == "unassigned" {
			baseQuery += " AND assigned_to IS NULL"
		} else {
			baseQuery += " AND assigned_to = ?"
			args = append(args, assigneeID)
		}
	}

	selectQuery, selectArgs, err := sqlx.In(
		"SELECT id, workspace_id, title, description, assigned_to, created_by, status, created_at, updated_at FROM tasks "+
			baseQuery+" ORDER BY created_at DESC, id DESC LIMIT ?",
		append(args, queryLimit)...,
	)
	if err != nil {
		return nil, err
	}
	selectQuery = r.db.Rebind(selectQuery)

	var tasks []models.Task
	err = r.db.Select(&tasks, selectQuery, selectArgs...)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}
