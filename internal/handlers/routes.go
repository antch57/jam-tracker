package handlers

import (
    "github.com/gin-gonic/gin"
)

// SetupRoutes configures all the routes for our application
func SetupRoutes(router *gin.Engine) {
    // Health check endpoint
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "ok",
            "message": "JamTracker API is running!",
        })
    })

    // API v1 routes
    v1 := router.Group("/api/v1")
    {
        // Authentication routes
        auth := v1.Group("/auth")
        {
            auth.POST("/register", RegisterUser)
            auth.POST("/login", LoginUser)
        }

        // Protected routes (we'll add JWT middleware later)
        protected := v1.Group("/")
        // protected.Use(AuthMiddleware()) // We'll implement this next
        {
            // User routes
            protected.GET("/profile", GetUserProfile)
            protected.PUT("/profile", UpdateUserProfile)

            // Show routes
            protected.POST("/shows", CreateShow)
            protected.GET("/shows", GetShows)
            protected.GET("/shows/:id", GetShow)
            protected.PUT("/shows/:id", UpdateShow)
            protected.DELETE("/shows/:id", DeleteShow)

            // Show attendance routes
            protected.POST("/shows/:id/attend", AttendShow)
            protected.PUT("/attendances/:id", UpdateAttendance)
            protected.DELETE("/attendances/:id", DeleteAttendance)

            // Band routes
            protected.POST("/bands", CreateBand)
            protected.GET("/bands", GetBands)
            protected.GET("/bands/:id", GetBand)

            // Venue routes
            protected.POST("/venues", CreateVenue)
            protected.GET("/venues", GetVenues)
            protected.GET("/venues/:id", GetVenue)

            // Recommendation routes
            protected.GET("/recommendations", GetRecommendations)
        }
    }
}