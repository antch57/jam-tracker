package handlers

import (
	"jam-tracker/internal/database"
	"jam-tracker/internal/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateShowRequest struct {
	BandName  string `json:"band_name" binding:"required,min=1"`
	VenueName string `json:"venue_name" binding:"required,min=1"`
	Date      string `json:"date" binding:"required"` // Format: "2024-12-25"
	Notes     string `json:"notes" binding:"omitempty,max=500"`
}

type UpdateShowRequest struct {
	BandName  string `json:"band_name" binding:"required,min=1"`
	VenueName string `json:"venue_name" binding:"required,min=1"`
	Date      string `json:"date" binding:"required"`
	Notes     string `json:"notes" binding:"omitempty,max=500"`
}

// Parse request body
type AttendShowRequest struct {
	Attended     bool     `json:"attended"`
	Rating       *float64 `json:"rating" binding:"omitempty,min=1,max=5"`
	FavoriteSong string   `json:"favorite_song" binding:"omitempty,max=200"`
	Notes        string   `json:"notes" binding:"omitempty,max=500"`
}

func CreateShow(c *gin.Context) {
	var req CreateShowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the band
	var band models.Band
	normalizedBandName := strings.ToLower(strings.TrimSpace(req.BandName))
	if err := database.DB.Where("name = ?", normalizedBandName).First(&band).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Band not found"})
		return
	}

	// Find the venue
	var venue models.Venue
	normalizedVenueName := strings.ToLower(strings.TrimSpace(req.VenueName))
	if err := database.DB.Where("name = ?", normalizedVenueName).First(&venue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Venue not found"})
		return
	}

	// Parse the date/time - try multiple formats
	var showDate time.Time
	var err error

	// Try ISO format with time first (2024-12-25T20:00:00Z)
	showDate, err = time.Parse(time.RFC3339, req.Date)
	if err != nil {
		// Try date only format (2024-12-25)
		showDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			// Try date with time format (2024-12-25 20:00)
			showDate, err = time.Parse("2006-01-02 15:04", req.Date)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid date format. Use YYYY-MM-DD, YYYY-MM-DD HH:MM, or ISO format",
				})
				return
			}
		}
	}

	// Check for duplicate show (same band, venue, and date)
	var existingShow models.Show
	if err := database.DB.Where("band_id = ? AND venue_id = ? AND DATE(date) = ?",
		band.ID, venue.ID, showDate.Format("2006-01-02")).First(&existingShow).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Show already exists for this band, venue, and date"})
		return
	}

	// Create the show
	show := models.Show{
		BandID:  band.ID,
		VenueID: venue.ID,
		Date:    showDate,
		Notes:   strings.TrimSpace(req.Notes),
	}

	if err := database.DB.Create(&show).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create show"})
		return
	}

	// Preload band and venue for response
	if err := database.DB.Preload("Band").Preload("Venue").First(&show, show.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load show details"})
		return
	}

	c.JSON(http.StatusCreated, show)
}

func GetShows(c *gin.Context) {
	var shows []models.Show

	// Get query parameters for filtering
	bandName := c.Query("band")
	venueName := c.Query("venue")
	city := c.Query("city")
	state := c.Query("state")
	dateFrom := c.Query("date_from") // YYYY-MM-DD
	dateTo := c.Query("date_to")     // YYYY-MM-DD

	query := database.DB.Preload("Band").Preload("Venue")

	// Filter by band name
	if bandName != "" {
		normalizedBandName := strings.ToLower(strings.ReplaceAll(bandName, "-", " "))
		query = query.Joins("JOIN bands ON shows.band_id = bands.id").
			Where("LOWER(bands.name) LIKE ?", "%"+normalizedBandName+"%")
	}

	// Filter by venue name
	if venueName != "" {
		normalizedVenueName := strings.ToLower(strings.ReplaceAll(venueName, "-", " "))
		query = query.Joins("JOIN venues ON shows.venue_id = venues.id").
			Where("LOWER(venues.name) LIKE ?", "%"+normalizedVenueName+"%")
	}

	// Filter by city
	if city != "" {
		normalizedCity := strings.ToLower(strings.ReplaceAll(city, "-", " "))
		query = query.Joins("JOIN venues ON shows.venue_id = venues.id").
			Where("LOWER(venues.city) LIKE ?", "%"+normalizedCity+"%")
	}

	// Filter by state
	if state != "" {
		normalizedState := strings.ToLower(strings.ReplaceAll(state, "-", " "))
		query = query.Joins("JOIN venues ON shows.venue_id = venues.id").
			Where("LOWER(venues.state) LIKE ?", "%"+normalizedState+"%")
	}

	// Filter by date range
	if dateFrom != "" {
		if fromDate, err := time.Parse("2006-01-02", dateFrom); err == nil {
			query = query.Where("DATE(shows.date) >= ?", fromDate.Format("2006-01-02"))
		}
	}

	if dateTo != "" {
		if toDate, err := time.Parse("2006-01-02", dateTo); err == nil {
			query = query.Where("DATE(shows.date) <= ?", toDate.Format("2006-01-02"))
		}
	}

	// Order by date (most recent first)
	query = query.Order("shows.date DESC")

	if err := query.Find(&shows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shows"})
		return
	}

	c.JSON(http.StatusOK, shows)
}

func GetShow(c *gin.Context) {
	showIDParam := c.Param("id")

	showID, err := uuid.Parse(showIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid show ID"})
		return
	}

	var show models.Show
	if err := database.DB.Preload("Band").Preload("Venue").Where("id = ?", showID).First(&show).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Show not found"})
		return
	}

	c.JSON(http.StatusOK, show)
}

func UpdateShow(c *gin.Context) {
	showIDParam := c.Param("id")

	showID, err := uuid.Parse(showIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid show ID"})
		return
	}

	var req UpdateShowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the show
	var show models.Show
	if err := database.DB.Where("id = ?", showID).First(&show).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Show not found"})
		return
	}

	// Find the band
	var band models.Band
	normalizedBandName := strings.ToLower(strings.TrimSpace(req.BandName))
	if err := database.DB.Where("name = ?", normalizedBandName).First(&band).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Band not found"})
		return
	}

	// Find the venue
	var venue models.Venue
	normalizedVenueName := strings.ToLower(strings.TrimSpace(req.VenueName))
	if err := database.DB.Where("name = ?", normalizedVenueName).First(&venue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Venue not found"})
		return
	}

	// Parse the date/time
	var showDate time.Time
	showDate, err = time.Parse(time.RFC3339, req.Date)
	if err != nil {
		showDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			showDate, err = time.Parse("2006-01-02 15:04", req.Date)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid date format. Use YYYY-MM-DD, YYYY-MM-DD HH:MM, or ISO format",
				})
				return
			}
		}
	}

	// Check for duplicate show (excluding current show)
	var existingShow models.Show
	if err := database.DB.Where("band_id = ? AND venue_id = ? AND DATE(date) = ? AND id != ?",
		band.ID, venue.ID, showDate.Format("2006-01-02"), show.ID).First(&existingShow).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Show already exists for this band, venue, and date"})
		return
	}

	// Update the show
	show.BandID = band.ID
	show.VenueID = venue.ID
	show.Date = showDate
	show.Notes = strings.TrimSpace(req.Notes)

	if err := database.DB.Save(&show).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update show"})
		return
	}

	// Reload with preloads
	if err := database.DB.Preload("Band").Preload("Venue").First(&show, show.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load show details"})
		return
	}

	c.JSON(http.StatusOK, show)
}

func DeleteShow(c *gin.Context) {
	showIDParam := c.Param("id")

	showID, err := uuid.Parse(showIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid show ID"})
		return
	}

	// Find the show
	var show models.Show
	if err := database.DB.Where("id = ?", showID).First(&show).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Show not found"})
		return
	}

	// Delete associated attendance records first
	if err := database.DB.Where("show_id = ?", show.ID).Delete(&models.ShowAttendance{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete attendance records"})
		return
	}

	// Delete the show
	if err := database.DB.Delete(&show).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete show"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Show deleted successfully"})
}

func AttendShow(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	showIDParam := c.Param("id")
	showID, err := uuid.Parse(showIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid show ID"})
		return
	}

	var req AttendShowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify show exists
	var show models.Show
	if err := database.DB.Where("id = ?", showID).First(&show).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Show not found"})
		return
	}

	// Check if user already has attendance for this show
	var existingAttendance models.ShowAttendance
	if err := database.DB.Where("user_id = ? AND show_id = ?", userID, showID).First(&existingAttendance).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Attendance already exists for this show. Use PUT to update."})
		return
	}

	// Create attendance record
	attendance := models.ShowAttendance{
		UserID:       userID.(uuid.UUID),
		ShowID:       showID,
		Attended:     req.Attended,
		Rating:       req.Rating,
		FavoriteSong: strings.TrimSpace(req.FavoriteSong),
		Notes:        strings.TrimSpace(req.Notes),
	}

	if err := database.DB.Create(&attendance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create attendance record"})
		return
	}

	// Preload relationships for response
	if err := database.DB.Preload("Show.Band").Preload("Show.Venue").First(&attendance, attendance.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load attendance details"})
		return
	}

	c.JSON(http.StatusCreated, attendance)
}

// UpdateAttendance - Update an existing attendance record
func UpdateAttendance(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get attendance ID from URL
	attendanceIDParam := c.Param("id")
	attendanceID, err := uuid.Parse(attendanceIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attendance ID"})
		return
	}

	// Parse request body
	type UpdateAttendanceRequest struct {
		Attended     bool     `json:"attended"`
		Rating       *float64 `json:"rating" binding:"omitempty,min=1,max=5"`
		FavoriteSong string   `json:"favorite_song" binding:"omitempty,max=200"`
		Notes        string   `json:"notes" binding:"omitempty,max=500"`
	}

	var req UpdateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the attendance record
	var attendance models.ShowAttendance
	if err := database.DB.Where("id = ?", attendanceID).First(&attendance).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attendance record not found"})
		return
	}

	// Verify the attendance belongs to the authenticated user
	if attendance.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own attendance records"})
		return
	}

	// Update the attendance
	attendance.Attended = req.Attended
	attendance.Rating = req.Rating
	attendance.FavoriteSong = strings.TrimSpace(req.FavoriteSong)
	attendance.Notes = strings.TrimSpace(req.Notes)

	if err := database.DB.Save(&attendance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update attendance record"})
		return
	}

	// Reload with relationships
	if err := database.DB.Preload("Show.Band").Preload("Show.Venue").First(&attendance, attendance.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load attendance details"})
		return
	}

	c.JSON(http.StatusOK, attendance)
}

// DeleteAttendance - Remove an attendance record
func DeleteAttendance(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get attendance ID from URL
	attendanceIDParam := c.Param("id")
	attendanceID, err := uuid.Parse(attendanceIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attendance ID"})
		return
	}

	// Find the attendance record
	var attendance models.ShowAttendance
	if err := database.DB.Where("id = ?", attendanceID).First(&attendance).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attendance record not found"})
		return
	}

	// Verify the attendance belongs to the authenticated user
	if attendance.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own attendance records"})
		return
	}

	// Delete the attendance record
	if err := database.DB.Delete(&attendance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete attendance record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Attendance record deleted successfully"})
}

func GetRecommendations(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "GetRecommendations not implemented yet"})
}
