package handlers

import (
	"net/http"
	"strings"

	"jam-tracker/internal/database"
	"jam-tracker/internal/models"

	"github.com/gin-gonic/gin"
)

type CreateBandRequest struct {
	Name        string `json:"band_name"`
	Genre       string `json:"genre"`
	Description string `json:"description"`
}

type UpdateBandRequest struct {
	Name        string `json:"band_name" binding:"required"`
	Genre       string `json:"genre" binding:"required"`
	Description string `json:"description" binding:"required"`
}

func CreateBand(c *gin.Context) {
	var req CreateBandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if band already exists
	var existingBand models.Band
	if err := database.DB.Where("LOWER(name) = LOWER(?)", req.Name).First(&existingBand).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Band with this name already exists"})
		return
	}

	// Create Band
	band := models.Band{
		Name:        strings.ToLower(req.Name),
		Genre:       req.Genre,
		Description: req.Description,
	}

	if err := database.DB.Create(&band).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create band"})
		return
	}

	c.JSON(http.StatusCreated, band)
}

func GetBands(c *gin.Context) {
	var bands []models.Band
	if err := database.DB.Find(&bands).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve bands"})
		return
	}

	c.JSON(http.StatusOK, bands)
}

func GetBand(c *gin.Context) {
	bandName := c.Param("id")
	bandName = strings.ToLower(bandName)

	// Find band by name
	var band models.Band
	if err := database.DB.Where("name = ?", bandName).First(&band).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Band not found"})
		return
	}

	c.JSON(http.StatusOK, band)
}

func UpdateBand(c *gin.Context) {
	// Get band name from URL
	bandName := c.Param("id")
	bandName = strings.ToLower(bandName)

	// Parse request body
	var req UpdateBandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find existing band
	var band models.Band
	if err := database.DB.Where("name = ?", bandName).First(&band).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Band not found"})
		return
	}

	// Update the band
	if err := database.DB.Model(&band).Updates(models.Band{
		Name:        strings.ToLower(req.Name),
		Genre:       req.Genre,
		Description: req.Description,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update band"})
		return
	}

	// Return updated band
	c.JSON(http.StatusOK, band)
}

func DeleteBand(c *gin.Context) {
	bandName := c.Param("id")
	bandName = strings.ToLower(bandName)

	// Find existing band
	var band models.Band
	if err := database.DB.Where("name = ?", bandName).First(&band).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Band not found"})
		return
	}

	// Delete the band
	if err := database.DB.Delete(&band).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete band"})
		return
	}

	// Return success message
	c.JSON(http.StatusOK, gin.H{"message": "Band deleted successfully"})
}
