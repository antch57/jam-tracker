package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func CreateBand(c *gin.Context) {
    c.JSON(http.StatusNotImplemented, gin.H{"message": "CreateBand not implemented yet"})
}

func GetBands(c *gin.Context) {
    c.JSON(http.StatusNotImplemented, gin.H{"message": "GetBands not implemented yet"})
}

func GetBand(c *gin.Context) {
    c.JSON(http.StatusNotImplemented, gin.H{"message": "GetBand not implemented yet"})
}

func UpdateBand(c *gin.Context) {
    c.JSON(http.StatusNotImplemented, gin.H{"message": "UpdateBand not implemented yet"})
}

func DeleteBand(c *gin.Context) {
    c.JSON(http.StatusNotImplemented, gin.H{"message": "DeleteBand not implemented yet"})
}
