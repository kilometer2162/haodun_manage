package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"haodun_manage/backend/utils"
)

func GetUsers(c *gin.Context) {
	var users []models.User
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	var total int64
	database.DB.Model(&models.User{}).Count(&total)
	database.DB.Preload("Role").Preload("Department").Offset(offset).Limit(pageSize).Find(&users)

	c.JSON(http.StatusOK, gin.H{
		"data":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.Preload("Role").Preload("Department").First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, user)
}

type createUserRequest struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password"`
	Email        string `json:"email"`
	Status       *int   `json:"status"`
	RoleID       uint   `json:"role_id" binding:"required"`
	DepartmentID uint   `json:"department_id" binding:"required"`
	EmployeeType string `json:"employee_type" binding:"required"`
}

func CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if password is empty
	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码不能为空"})
		return
	}

	// Check if password is already hashed
	if utils.IsBcryptHash(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码不能是哈希格式，请提供明文密码"})
		return
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	if req.DepartmentID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择所属部门"})
		return
	}

	employeeType := strings.TrimSpace(req.EmployeeType)
	if !isValidEmployeeType(employeeType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "员工类型不合法"})
		return
	}
	employeeType = strings.ToLower(employeeType)

	var department models.Department
	if err := database.DB.First(&department, req.DepartmentID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "所属部门不存在"})
		return
	}
	if department.Status != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "所属部门已禁用"})
		return
	}

	// Set default status if not provided
	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	user := models.User{
		Username:     req.Username,
		Password:     hashedPassword,
		Email:        req.Email,
		Status:       status,
		RoleID:       req.RoleID,
		DepartmentID: req.DepartmentID,
		EmployeeType: employeeType,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建用户失败"})
		return
	}

	database.DB.Preload("Role").Preload("Department").First(&user, user.ID)
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

type updateUserRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Email        string `json:"email"`
	Status       *int   `json:"status"`
	RoleID       uint   `json:"role_id"`
	DepartmentID uint   `json:"department_id"`
	EmployeeType string `json:"employee_type"`
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.Username = req.Username
	user.Email = req.Email
	if req.Status != nil {
		user.Status = *req.Status
	}
	user.RoleID = req.RoleID

	if req.DepartmentID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择所属部门"})
		return
	}

	employeeType := strings.TrimSpace(req.EmployeeType)
	if !isValidEmployeeType(employeeType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "员工类型不合法"})
		return
	}
	employeeType = strings.ToLower(employeeType)

	var department models.Department
	if err := database.DB.First(&department, req.DepartmentID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "所属部门不存在"})
		return
	}
	if department.Status != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "所属部门已禁用"})
		return
	}

	user.DepartmentID = req.DepartmentID
	user.EmployeeType = employeeType

	if req.Password != "" {
		if utils.IsBcryptHash(req.Password) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "密码不能是哈希格式，请提供明文密码"})
			return
		}
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		user.Password = hashedPassword
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "更新用户失败"})
		return
	}

	database.DB.Preload("Role").Preload("Department").First(&user, user.ID)
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&models.User{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func isValidEmployeeType(value string) bool {
	switch strings.ToLower(value) {
	case "internal", "external":
		return true
	default:
		return false
	}
}
