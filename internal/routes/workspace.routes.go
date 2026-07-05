package routes

import (
	"task-system/internal/handlers"
	"task-system/internal/middleware"
	"task-system/internal/repository"
	"task-system/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterWorkspaceRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	workspaceRepo := repository.NewWorkspaceRepository(db)
	workspaceService := service.NewWorkspaceService(workspaceRepo)
	workspaceHandler := handlers.NewWorkspaceHandler(workspaceService)

	auth := router.Group("/workspace", middleware.RequireAuth(db))
	auth.POST("/create", workspaceHandler.Create)
	auth.POST("/invite", workspaceHandler.Invite)
	auth.POST("/join", workspaceHandler.Join)
	auth.GET("/:id", workspaceHandler.Get)
	auth.GET("", workspaceHandler.List)
	auth.PUT("/:id", workspaceHandler.Update)
	auth.DELETE("/:id", workspaceHandler.Delete)
	auth.GET("/:id/members", workspaceHandler.ListMembers)
	auth.PUT("/:id/members/:user_id", workspaceHandler.UpdateMemberRole)
	auth.DELETE("/:id/members/:user_id", workspaceHandler.KickMember)
	auth.POST("/:id/leave", workspaceHandler.Leave)
	auth.POST("/:id/transfer", workspaceHandler.TransferOwnership)
}
