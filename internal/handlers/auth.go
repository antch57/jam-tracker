package handlers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"jam-tracker/internal/database"
	"jam-tracker/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Location  string `json:"location"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name" binding:"required,min=1,max=50"`
	LastName  string `json:"last_name" binding:"required,min=1,max=50"`
	Location  string `json:"location" binding:"max=100"`
}

type DeleteRequest struct {
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// RegisterUser creates a new user account
func RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email or username already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := models.User{
		Email:     req.Email,
		Username:  req.Username,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Location:  req.Location,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := generateJWT(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Remove password from response
	user.Password = ""

	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// LoginUser authenticates a user and returns a JWT token
func LoginUser(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := generateJWT(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Remove password from response
	user.Password = ""

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

// GetUserProfile returns the current user's profile
func GetUserProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	// Find user by ID
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Remove password from response
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// UpdateUserProfile updates the current user's profile
func UpdateUserProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by ID
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	// Update user fields
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Location = req.Location

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	// Remove password from response
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// DeleteUserAccount deletes the current user's account
func DeleteUserAccount(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required to delete account"})
		return
	}

	// Find user by ID
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found."})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password."})
		return
	}

	// Delete user's show attendance
	if err := database.DB.Where("user_id = ?", userID).Delete(&models.ShowAttendance{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user data"})
		return
	}

	// Hard delete user
	if err := database.DB.Unscoped().Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user account"})
		return
	}
	// Step 7: Return success
	c.JSON(http.StatusOK, gin.H{"message": "User account deleted successfully"})

}

// generateJWT creates a JWT token for the given user ID
func generateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// Get the JWT secret from environment variable
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET not set.")
	}
	return []byte(secret)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Validate the format of the header
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Check the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			// Return the secret key
			return getJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract user ID from token claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userIDStr, exists := claims["user_id"].(string)
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				c.Abort()
				return
			}

			// Convert string to UUID
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
				c.Abort()
				return
			}

			// Store user ID in context for handlers to use
			c.Set("user_id", userID)

			// Continue to the next handler
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
	}
}
