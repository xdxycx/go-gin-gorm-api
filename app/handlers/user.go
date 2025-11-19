package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-gin-gorm-api/app/config"
	"go-gin-gorm-api/app/models"
	"go-gin-gorm-api/app/utils" // 【修改】引入 utils 包
)

// CreateUser 处理创建用户请求
func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: "请求参数错误", Data: gin.H{"detail": err.Error()}})
		return
	}

	result := config.DB.Create(&user)
	if result.Error != nil {
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusInternalServerError, utils.APIResponse{Code: 500, Message: "创建用户失败", Data: gin.H{"detail": result.Error.Error()}})
		return
	}

	// 【修改】使用统一的 APIResponse 结构
	c.JSON(http.StatusCreated, utils.APIResponse{Code: 0, Message: "创建成功", Data: user})
}

// GetUsers 处理获取所有用户请求
func GetUsers(c *gin.Context) {
	var users []models.User
	config.DB.Find(&users)
	// 【修改】使用统一的 APIResponse 结构
	c.JSON(http.StatusOK, utils.APIResponse{Code: 0, Message: "查询成功", Data: users})
}

// GetUserByID 处理根据 ID 获取用户请求
func GetUserByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: "用户ID格式错误"})
		return
	}
	
	var user models.User

	if err := config.DB.First(&user, id).Error; err != nil {
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusNotFound, utils.APIResponse{Code: 404, Message: "用户未找到"})
		return
	}

	// 【修改】使用统一的 APIResponse 结构
	c.JSON(http.StatusOK, utils.APIResponse{Code: 0, Message: "查询成功", Data: user})
}

// UpdateUser 处理更新用户请求
func UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: "用户ID格式错误"})
		return
	}
	
	var user models.User
	
	// 1. 查找用户
	if err := config.DB.First(&user, id).Error; err != nil {
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusNotFound, utils.APIResponse{Code: 404, Message: "用户未找到"})
		return
	}
	
	// 2. 绑定请求体
	if err := c.ShouldBindJSON(&user); err != nil {
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: "请求参数错误", Data: gin.H{"detail": err.Error()}})
		return
	}
	
	// 3. 更新用户
	config.DB.Save(&user)
	// 【修改】使用统一的 APIResponse 结构
	c.JSON(http.StatusOK, utils.APIResponse{Code: 0, Message: "更新成功", Data: user})
}

// DeleteUser 处理删除用户请求
func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusBadRequest, utils.APIResponse{Code: 400, Message: "用户ID格式错误"})
		return
	}
	
	// Gorm 软删除
	result := config.DB.Delete(&models.User{}, id)
	
	if result.Error != nil {
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusInternalServerError, utils.APIResponse{Code: 500, Message: "删除失败", Data: gin.H{"detail": result.Error.Error()}})
		return
	}

	if result.RowsAffected == 0 {
		// 【修改】使用统一的 APIResponse 结构
		c.JSON(http.StatusNotFound, utils.APIResponse{Code: 404, Message: "用户未找到"})
		return
	}
	
	// 【修改】使用统一的 APIResponse 结构
	c.JSON(http.StatusOK, utils.APIResponse{Code: 0, Message: "删除成功", Data: nil})
}
