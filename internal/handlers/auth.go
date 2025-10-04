package handlers

import (
    "net/http"
    "time"

    "jam-tracker/internal/database"
    "jam-tracker/internal/models"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
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
    // We'll implement this after JWT middleware
    c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// UpdateUserProfile updates the current user's profile
func UpdateUserProfile(c *gin.Context) {
    // We'll implement this after JWT middleware
    c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// generateJWT creates a JWT token for the given user ID
func generateJWT(userID string) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
        "iat":     time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte("your-secret-key-change-this")) // TODO: Use config
}