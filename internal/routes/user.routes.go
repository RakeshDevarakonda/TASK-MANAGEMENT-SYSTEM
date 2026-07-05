package routes

import (
	"task-system/internal/handlers"
	"task-system/internal/middleware"
	"task-system/internal/repository"
	"task-system/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterUserRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	workspaceRepo := repository.NewWorkspaceRepository(db)

	userService := service.NewUserService(userRepo, sessionRepo)
	workspaceService := service.NewWorkspaceService(workspaceRepo)

	userHandler := handlers.NewUserHandler(userService, workspaceService)

	auth := router.Group("/user", middleware.RequireAuth(db))
	auth.GET("/profile", userHandler.GetProfile)
	auth.PUT("/profile", userHandler.UpdateProfile)
	auth.DELETE("/profile", userHandler.DeleteProfile)
	auth.GET("/invitations", userHandler.GetInvitations)
}
