package routes

import (
	"task-system/internal/handlers"
	"task-system/internal/middleware"
	"task-system/internal/repository"
	"task-system/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterAuthRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	authService := service.NewAuthService(userRepo, sessionRepo)
	authHandler := handlers.NewAuthHandler(authService)

	auth := router.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/signin", authHandler.Signin)
	auth.POST("/refresh", authHandler.Refresh)
	auth.POST("/logout", middleware.RequireAuth(db), authHandler.Logout)
}
