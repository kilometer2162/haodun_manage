package middleware

import (
	"haodun_manage/backend/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// 记录访问日志
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		
		var uid uint
		if userID != nil {
			uid = userID.(uint)
		}
		
		var uname string
		if username != nil {
			uname = username.(string)
		}

		ip := utils.GetClientIP(c.Request)
		userAgent := c.Request.UserAgent()
		action := c.Request.Method
		module := c.FullPath()
		content := c.Request.URL.Path

		status := 1
		if c.Writer.Status() >= 400 {
			status = 0
		}

		utils.LogAction(
			uid,
			uname,
			action,
			module,
			content,
			ip,
			userAgent,
			status,
		)

		// 记录IP访问统计
		go utils.RecordIPAccess(ip, c.Request)

		_ = start // 可以用于记录响应时间
	}
}

