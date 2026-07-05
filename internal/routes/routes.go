package routes

import (
	"task-system/internal/database"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	api := router.Group("/api/v1")

	RegisterAuthRoutes(api, database.DB)
	RegisterWorkspaceRoutes(api, database.DB)
	RegisterTaskRoutes(api, database.DB)
	RegisterUserRoutes(api, database.DB)

}
