package handlers

import (
	"net/http"
	"strings"

	"jam-tracker/internal/database"
	"jam-tracker/internal/models"

	"github.com/gin-gonic/gin"
)

type CreateVenueRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=100"`
	City     string `json:"city" binding:"required,min=1,max=200"`
	State    string `json:"state" binding:"required,min=1,max=100"`
	Country  string `json:"country" binding:"omitempty,min=1,max=100"`
	Address  string `json:"address" binding:"omitempty,min=1,max=200"`
	Capacity string `json:"capacity" binding:"omitempty,min=1"`
}

type UpdateVenueRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=100"`
	City     string `json:"city" binding:"required,min=1,max=200"`
	State    string `json:"state" binding:"required,min=1,max=100"`
	Country  string `json:"country" binding:"omitempty,min=1,max=100"`
	Address  string `json:"address" binding:"omitempty,min=1,max=200"`
	Capacity string `json:"capacity" binding:"omitempty,min=1"`
}

func CreateVenue(c *gin.Context) {
	var req CreateVenueRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if venue already exists.
	var existingVenue models.Venue
	normalizedVenueName := strings.ToLower(strings.TrimSpace(req.Name))
	normalizedCity := strings.ToLower(strings.TrimSpace(req.City))
	normalizedState := strings.ToLower(strings.TrimSpace(req.State))
	normalizedCountry := strings.ToLower(strings.TrimSpace(req.Country))
	normalizedAddress := strings.TrimSpace(req.Address)

	if err := database.DB.Where("name = ? AND city = ? AND state = ?",
		normalizedVenueName,
		normalizedCity,
		normalizedState).First(&existingVenue).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Venue already exists"})
		return
	}

	venue := models.Venue{
		Name:     normalizedVenueName,
		City:     normalizedCity,
		State:    normalizedState,
		Country:  normalizedCountry,
		Address:  normalizedAddress,
		Capacity: req.Capacity,
	}

	if err := database.DB.Create(&venue).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create venue"})
		return
	}

	c.JSON(http.StatusCreated, venue)
}

func GetVenues(c *gin.Context) {
	var venues []models.Venue

	// Get query parameters for filtering
	name := c.Query("name")
	city := c.Query("city")
	state := c.Query("state")
	country := c.Query("country")

	query := database.DB

	// Add filters if provided
	if name != "" {
		normalizedName := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(name, "-", " ")))
		query = query.Where("name LIKE ?", "%"+normalizedName+"%")
	}

	if city != "" {
		normalizedCity := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(city, "-", " ")))
		query = query.Where("city LIKE ?", "%"+normalizedCity+"%")
	}

	if state != "" {
		normalizedState := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(state, "-", " ")))
		query = query.Where("state LIKE ?", "%"+normalizedState+"%")
	}

	if country != "" {
		normalizedCountry := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(country, "-", " ")))
		query = query.Where("country LIKE ?", "%"+normalizedCountry+"%")
	}

	if err := query.Find(&venues).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch venues"})
		return
	}

	c.JSON(http.StatusOK, venues)
}

func GetVenue(c *gin.Context) {
	venueName := c.Param("id")
	normalizedVenueName := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(venueName, "-", " ")))

	var venue models.Venue
	if err := database.DB.Where("name = ?", normalizedVenueName).First(&venue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Venue not found"})
		return
	}

	c.JSON(http.StatusOK, venue)
}

func UpdateVenue(c *gin.Context) {
	var req UpdateVenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	venueName := c.Param("id")
	normalizedVenueName := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(venueName, "-", " ")))

	var venue models.Venue
	if err := database.DB.Where("name = ?", normalizedVenueName).First(&venue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Venue not found"})
		return
	}

	normalizedCity := strings.ToLower(strings.TrimSpace(req.City))
	normalizedState := strings.ToLower(strings.TrimSpace(req.State))
	normalizedCountry := strings.ToLower(strings.TrimSpace(req.Country))
	normalizedAddress := strings.TrimSpace(req.Address)

	// Update the venue
	if err := database.DB.Model(&venue).Updates(models.Venue{
		Name:     normalizedVenueName,
		City:     normalizedCity,
		State:    normalizedState,
		Country:  normalizedCountry,
		Address:  normalizedAddress,
		Capacity: req.Capacity,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update band"})
		return
	}

	// Return updated venue
	c.JSON(http.StatusOK, venue)
}

func DeleteVenue(c *gin.Context) {
	venueName := c.Param("id")
	normalizedVenueName := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(venueName, "-", " ")))

	if normalizedVenueName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Venue ID is required"})
		return
	}

	// Find venue
	var venue models.Venue
	if err := database.DB.Where("name = ?", normalizedVenueName).First(&venue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Venue not found"})
		return
	}

	// Check if venue has any shows
	var showCount int64
	if err := database.DB.Model(&models.Show{}).Where("venue_id = ?", venue.ID).Count(&showCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check venue usage"})
		return
	}

	if showCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":       "Cannot delete venue that has shows associated with it",
			"shows_count": showCount,
		})
		return
	}

	// Delete venue
	if err := database.DB.Delete(&venue).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete venue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Venue deleted successfully"})
}
