package routes

import (
	"task-system/internal/handlers"
	"task-system/internal/middleware"
	"task-system/internal/repository"
	"task-system/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterTaskRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	taskRepo := repository.NewTaskRepository(db)
	workspaceRepo := repository.NewWorkspaceRepository(db)
	taskService := service.NewTaskService(taskRepo, workspaceRepo)
	taskHandler := handlers.NewTaskHandler(taskService)

	auth := router.Group("/task", middleware.RequireAuth(db))
	auth.POST("/create", taskHandler.Create)
	auth.GET("/:id", taskHandler.Get)
	auth.GET("/workspace/:workspace_id", taskHandler.List)
	auth.PUT("/:id", taskHandler.Update)
	auth.DELETE("/:id", taskHandler.Delete)
	auth.POST("/:id/assign", taskHandler.Assign)
	auth.POST("/:id/submit", taskHandler.Submit)
}
