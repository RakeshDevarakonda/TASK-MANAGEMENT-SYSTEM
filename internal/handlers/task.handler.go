package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"task-system/internal/models"
	"task-system/internal/service"
	"task-system/internal/utils"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(svc *service.TaskService) *TaskHandler {
	return &TaskHandler{service: svc}
}

// Create checks body payloads and maps service creation tasks
func (h *TaskHandler) Create(c *gin.Context) {
	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	currentUser := c.MustGet("currentUser").(*models.User)

	task, err := h.service.CreateTask(req.WorkspaceID, req.Title, req.Description, req.AssignedTo, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "task created successfully", task)
}

// Get retrieves a specific task by ID
func (h *TaskHandler) Get(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("currentUser").(*models.User)

	task, err := h.service.GetTask(id, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "task retrieved successfully", task)
}

// List lists all tasks within a workspace with cursor pagination and filters
func (h *TaskHandler) List(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	currentUser := c.MustGet("currentUser").(*models.User)

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	cursorStr := c.Query("cursor")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var cursorTime int64
	var cursorID string
	if cursorStr != "" {
		cursorTime, cursorID, err = utils.DecodeCursor(cursorStr)
		if err != nil {
			utils.SendError(c, http.StatusBadRequest, "invalid cursor parameter", nil)
			return
		}
	}

	// Parse status filter (comma-separated values)
	var statusList []string
	statusQuery := c.Query("status")
	if statusQuery != "" {
		for _, s := range strings.Split(statusQuery, ",") {
			cleaned := strings.TrimSpace(s)
			if cleaned != "" {
				statusList = append(statusList, cleaned)
			}
		}
	}

	assigneeID := c.Query("assignee_id")

	tasks, err := h.service.ListWorkspaceTasks(workspaceID, currentUser.ID, limit, cursorTime, cursorID, statusList, assigneeID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	var nextCursor string
	if len(tasks) > limit {
		lastTask := tasks[limit-1]
		nextCursor = utils.EncodeCursor(lastTask.CreatedAt, lastTask.ID)
		tasks = tasks[:limit]
	}

	utils.SendCursorSuccess(c, http.StatusOK, "tasks retrieved successfully", tasks, limit, nextCursor)
}

// Update modifies selected properties of a task
func (h *TaskHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	currentUser := c.MustGet("currentUser").(*models.User)

	task, err := h.service.UpdateTask(id, req, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "task updated successfully", task)
}

// Delete removes a task by ID
func (h *TaskHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("currentUser").(*models.User)

	err := h.service.DeleteTask(id, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "task deleted successfully", nil)
}

// Assign delegates a task to a workspace member (owner or admin only)
func (h *TaskHandler) Assign(c *gin.Context) {
	id := c.Param("id")
	var req models.AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	currentUser := c.MustGet("currentUser").(*models.User)

	task, err := h.service.AssignTask(id, req.AssignedTo, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "task assigned successfully", task)
}

// Submit completes the task (assigned member only)
func (h *TaskHandler) Submit(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("currentUser").(*models.User)

	task, err := h.service.SubmitTask(id, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "task submitted successfully", task)
}
