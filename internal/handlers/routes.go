package handlers

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the routes for our application
func SetupRoutes(router *gin.Engine) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
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

		// Protected routes
		protected := v1.Group("/")
		protected.Use(AuthMiddleware())
		{
			// User routes
			protected.GET("/profile", GetUserProfile)
			protected.PUT("/profile", UpdateUserProfile)
			protected.DELETE("/profile", DeleteUserAccount)

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
			protected.PUT("/bands/:id", UpdateBand)
			protected.GET("/bands", GetBands)
			protected.GET("/bands/:id", GetBand)
			protected.DELETE("/bands/:id", DeleteBand)

			// Venue routes
			protected.POST("/venues", CreateVenue)
			protected.GET("/venues", GetVenues)
			protected.GET("/venues/:id", GetVenue)
			protected.PUT("/venues/:id", UpdateVenue)
			protected.DELETE("/venues/:id", DeleteVenue)

			// Recommendation routes
			protected.GET("/recommendations", GetRecommendations)
		}
	}
}
