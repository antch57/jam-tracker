package main

import (
    "log"
    "jam-tracker/internal/config"
    "jam-tracker/internal/database"
    "jam-tracker/internal/handlers"

    "github.com/gin-gonic/gin"
)


func main() {
    // Load configuration
    cfg := config.Load()

    // Connect to database
    database.Connect()
    database.Migrate()

    // Set gin mode based on environment
    if cfg.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    // Initialize router
    router := gin.Default()

    // Add middleware
    router.Use(gin.Logger())
    router.Use(gin.Recovery())

    // Setup routes
    handlers.SetupRoutes(router)

    // Start server
    log.Printf("Starting server on port %s", cfg.Port)
    if err := router.Run(":" + cfg.Port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}