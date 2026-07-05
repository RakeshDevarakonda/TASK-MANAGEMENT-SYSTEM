package main

import (
	"log"
	"task-system/internal/config"
	"task-system/internal/database"
	"task-system/internal/middleware"
	"task-system/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/joho/godotenv"
)

func main() {
	// Enable strict JSON validation (Joi-style: reject unknown fields)
	binding.EnableDecoderDisallowUnknownFields = true

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, falling back to system environment variables")
	}

	// Load config struct properties
	config.LoadConfig()

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.DB.Close()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.RateLimit())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	routes.Register(router)

	log.Printf("Server listening on %s", ":8080")
router.Run(":8080")
}

