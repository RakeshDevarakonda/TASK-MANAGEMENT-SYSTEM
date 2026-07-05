package service

import (
	"fmt"
	"task-system/internal/models"
	"task-system/internal/repository"
	"task-system/internal/utils"
)

type TaskService struct {
	taskRepo      *repository.TaskRepository
	workspaceRepo *repository.WorkspaceRepository
}

func NewTaskService(taskRepo *repository.TaskRepository, workspaceRepo *repository.WorkspaceRepository) *TaskService {
	return &TaskService{
		taskRepo:      taskRepo,
		workspaceRepo: workspaceRepo,
	}
}

// CreateTask verifies authorization constraints and inserts a task
func (s *TaskService) CreateTask(workspaceID string, title string, description string, assignedTo string, createdBy string) (*models.Task, error) {
	// 1. Authorize: user must belong to the workspace (owner, admin, or member)
	isMember, err := s.workspaceRepo.IsMemberOrOwner(workspaceID, createdBy)
	if err != nil || !isMember {
		return nil, fmt.Errorf("%w: you do not have permission to create tasks in this workspace", utils.ErrForbidden)
	}

	// 2. Validate Assignee: if assigned, the user must be a member or owner of the workspace
	if assignedTo != "" {
		isAssigneeMember, err := s.workspaceRepo.IsMemberOrOwner(workspaceID, assignedTo)
		if err != nil {
			return nil, err
		}
		if !isAssigneeMember {
			return nil, fmt.Errorf("%w: assigned user must be a member of the workspace", utils.ErrForbidden)
		}
	}

	task := &models.Task{
		WorkspaceID: workspaceID,
		Title:       title,
		CreatedBy:   createdBy,
		Status:      "todo",
	}

	if description != "" {
		task.Description = &description
	}
	if assignedTo != "" {
		task.AssignedTo = &assignedTo
	}

	// 3. Insert the task into the database
	err = s.taskRepo.CreateTask(task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetTask retrieves a specific task if the user belongs to the task's workspace
func (s *TaskService) GetTask(taskID string, userID string) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("%w: task not found", utils.ErrNotFound)
	}

	// Authorize: user must be a member or owner of the workspace
	isMember, err := s.workspaceRepo.IsMemberOrOwner(task.WorkspaceID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("%w: you do not have access to this workspace", utils.ErrForbidden)
	}

	return task, nil
}

// ListWorkspaceTasks retrieves tasks in a workspace for members/owners using keyset pagination
func (s *TaskService) ListWorkspaceTasks(workspaceID string, userID string, limit int, cursorTime int64, cursorID string, status []string, assigneeID string) ([]models.Task, error) {
	// Authorize: user must be a member or owner of the workspace
	isMember, err := s.workspaceRepo.IsMemberOrOwner(workspaceID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("%w: you do not have access to this workspace", utils.ErrForbidden)
	}

	return s.taskRepo.ListWorkspaceTasksCursor(workspaceID, limit, cursorTime, cursorID, status, assigneeID)
}

// UpdateTask updates the task details matching user role restrictions and state transitions
func (s *TaskService) UpdateTask(taskID string, req models.UpdateTaskRequest, userID string) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("%w: task not found", utils.ErrNotFound)
	}

	// Fetch requester role
	reqRole, err := s.workspaceRepo.GetUserRole(task.WorkspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: you do not have access to this workspace", utils.ErrForbidden)
	}

	isOwnerOrAdmin := (reqRole == "super_admin" || reqRole == "admin")
	isCreator := (task.CreatedBy == userID)
	isAssignee := (task.AssignedTo != nil && *task.AssignedTo == userID)

	// Check general access: must be owner, admin, creator, or assignee
	if !isOwnerOrAdmin && !isCreator && !isAssignee {
		return nil, fmt.Errorf("%w: you do not have permission to update this task", utils.ErrForbidden)
	}

	// 1. Validate Status Transitions (if changed)
	if req.Status != nil && *req.Status != task.Status {
		if !isValidTransition(task.Status, *req.Status, isOwnerOrAdmin) {
			return nil, fmt.Errorf("%w: invalid task status transition", utils.ErrForbidden)
		}
		task.Status = *req.Status
	}

	// 2. Validate Properties Modifications (Title, Description, Assignee)
	if req.Title != nil || req.Description != nil || req.AssignedTo != nil {
		// Only owners, admins, and creators can update properties. Assignees who aren't creators can only change status.
		if !isOwnerOrAdmin && !isCreator {
			return nil, fmt.Errorf("%w: members can only update task status", utils.ErrForbidden)
		}

		if req.Title != nil {
			task.Title = *req.Title
		}
		if req.Description != nil {
			if *req.Description == "" {
				task.Description = nil
			} else {
				task.Description = req.Description
			}
		}
		if req.AssignedTo != nil {
			if *req.AssignedTo == "" {
				task.AssignedTo = nil
			} else {
				// Validate that new assignee is a workspace member/owner
				isMember, err := s.workspaceRepo.IsMemberOrOwner(task.WorkspaceID, *req.AssignedTo)
				if err != nil {
					return nil, err
				}
				if !isMember {
					return nil, fmt.Errorf("%w: assigned user must be a member of the workspace", utils.ErrForbidden)
				}
				task.AssignedTo = req.AssignedTo
			}
		}
	}

	err = s.taskRepo.UpdateTask(task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// DeleteTask removes a task if the user is the creator or an admin/owner
func (s *TaskService) DeleteTask(taskID string, userID string) error {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return fmt.Errorf("%w: task not found", utils.ErrNotFound)
	}

	// Fetch requester role
	reqRole, err := s.workspaceRepo.GetUserRole(task.WorkspaceID, userID)
	if err != nil {
		return fmt.Errorf("%w: you do not have access to this workspace", utils.ErrForbidden)
	}

	isOwnerOrAdmin := (reqRole == "super_admin" || reqRole == "admin")
	isCreator := (task.CreatedBy == userID)

	// Creator, Owner, and Admin can delete
	if !isOwnerOrAdmin && !isCreator {
		return fmt.Errorf("%w: only creators, admins, and owners can delete tasks", utils.ErrForbidden)
	}

	return s.taskRepo.DeleteTask(taskID)
}

// AssignTask assigns a task to a user (owner or admin only)
func (s *TaskService) AssignTask(taskID string, assignedTo string, userID string) (*models.Task, error) {
	req := models.UpdateTaskRequest{
		AssignedTo: &assignedTo,
	}
	return s.UpdateTask(taskID, req, userID)
}

// SubmitTask updates the status of the task to completed (assigned member only)
func (s *TaskService) SubmitTask(taskID string, userID string) (*models.Task, error) {
	status := "submitted"
	req := models.UpdateTaskRequest{
		Status: &status,
	}
	return s.UpdateTask(taskID, req, userID)
}

// Helper: verifies if a task transition matches the state machine permissions
func isValidTransition(from, to string, isAdmin bool) bool {
	if from == to {
		return true
	}
	switch from {
	case "todo":
		return to == "in_progress"
	case "in_progress":
		return to == "submitted"
	case "submitted":
		// Moving to completed or reopening back to in_progress requires admin
		if to == "completed" || to == "in_progress" {
			return isAdmin
		}
		return false
	case "completed":
		// Reopening requires admin
		if to == "in_progress" {
			return isAdmin
		}
		return false
	}
	return false
}
