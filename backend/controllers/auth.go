package controllers

import (
	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"haodun_manage/backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captcha_id"`   // 验证码暂时隐藏，改为可选
	CaptchaCode string `json:"captcha_code"` // 验证码暂时隐藏，改为可选
}

func getRolePermissionCodes(roleID uint) ([]string, error) {
	codes := make([]string, 0)
	if roleID == 0 {
		return codes, nil
	}

	err := database.DB.Table("role_permissions").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Pluck("permissions.code", &codes).Error

	return codes, err
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证验证码 - 验证码暂时隐藏，如果提供了验证码才验证
	if req.CaptchaID != "" && req.CaptchaCode != "" {
		if !utils.VerifyCaptcha(req.CaptchaID, req.CaptchaCode) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误"})
			return
		}
	}

	// 查找用户
	var user models.User
	if err := database.DB.Where("username = ?", req.Username).
		Preload("Role").
		Preload("Department").
		First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误1"})
		return
	}

	// 验证密码
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误2"})
		return
	}

	// 检查用户状态
	if user.Status != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "用户已被禁用"})
		return
	}

	// 生成token
	token, err := utils.GenerateToken(user.ID, user.Username, user.RoleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	// 记录登录日志
	ip := utils.GetClientIP(c.Request)
	utils.LogAction(user.ID, user.Username, "登录", "认证", "用户登录", ip, c.Request.UserAgent(), 1)

	// 获取角色权限与角色信息
	permissionCodes, _ := getRolePermissionCodes(user.RoleID)
	userResponse := gin.H{
		"id":            user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"role_id":       user.RoleID,
		"role":          user.Role,
		"department_id": user.DepartmentID,
		"department":    user.Department,
		"employee_type": user.EmployeeType,
		"permissions":   permissionCodes,
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  userResponse,
	})
}

func GetCaptcha(c *gin.Context) {
	captchaID, code := utils.GenerateCaptcha()
	c.JSON(http.StatusOK, gin.H{
		"captcha_id":   captchaID,
		"captcha_code": code, // 实际生产环境不应该返回验证码，前端应该生成图片
	})
}

func GetCurrentUser(c *gin.Context) {
	userIDValue, _ := c.Get("user_id")
	roleIDValue, _ := c.Get("role_id")

	userID, _ := userIDValue.(uint)
	roleID, _ := roleIDValue.(uint)

	var user models.User
	if err := database.DB.Preload("Role").Preload("Department").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	permissionCodes, _ := getRolePermissionCodes(roleID)
	response := gin.H{
		"id":            user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"role_id":       roleID,
		"role":          user.Role,
		"department_id": user.DepartmentID,
		"department":    user.Department,
		"employee_type": user.EmployeeType,
		"permissions":   permissionCodes,
	}

	c.JSON(http.StatusOK, response)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 验证旧密码（必须提供）
	if !utils.CheckPasswordHash(req.OldPassword, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "原密码错误"})
		return
	}

	// 更新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	user.Password = hashedPassword
	database.DB.Save(&user)

	// 记录操作日志
	ip := utils.GetClientIP(c.Request)
	utils.LogAction(user.ID, user.Username, "修改密码", "认证", "用户修改密码", ip, c.Request.UserAgent(), 1)

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}
