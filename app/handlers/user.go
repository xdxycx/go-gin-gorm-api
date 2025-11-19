package handlers

import (
	"net/http"

	"go-gin-gorm-api/config"
	"go-gin-gorm-api/models"

	"github.com/gin-gonic/gin"
)

// GetUsersHandler 获取所有用户列表
func GetUsersHandler(c *gin.Context) {
	var users []models.User
	
	result := config.DB.Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取用户数据"})
		return
	}

	c.JSON(http.StatusOK, users)
}
