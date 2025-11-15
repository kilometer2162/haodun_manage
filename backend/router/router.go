package router

import (
	"haodun_manage/backend/controllers"
	"haodun_manage/backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// 公开路由
	api := r.Group("/api")
	{
		// 认证相关
		api.POST("/auth/login", controllers.Login)
		api.GET("/auth/captcha", controllers.GetCaptcha)

		// 根据代码获取字典（公开接口，用于前端下拉选择等）
		api.GET("/dict/:code", controllers.GetDictByCode)

		// 根据键获取系统参数（公开接口）
		api.GET("/config/:key", controllers.GetConfigByKey)

		// Shein API 相关接口（公开，用于测试）
		shein := api.Group("/shein")
		{
			sheinController := controllers.NewSheinController()
			shein.GET("/sync-orders", sheinController.SyncOrders)
			shein.GET("/order-list", sheinController.GetOrderList)
			shein.GET("/order-detail", sheinController.GetOrderDetail)
		}

		// 需要认证的路由
		auth := api.Group("")
		auth.Use(middleware.AuthMiddleware())
		auth.Use(middleware.LogMiddleware())
		{
			// 用户信息
			auth.GET("/auth/current-user", controllers.GetCurrentUser)
			auth.PUT("/auth/change-password", controllers.ChangePassword)

			// 用户管理
			auth.GET("/users", controllers.GetUsers)
			auth.GET("/users/:id", controllers.GetUser)
			auth.POST("/users", controllers.CreateUser)
			auth.PUT("/users/:id", controllers.UpdateUser)
			auth.DELETE("/users/:id", controllers.DeleteUser)

			// 角色管理
			auth.GET("/roles", controllers.GetRoles)
			auth.GET("/roles/:id", controllers.GetRole)
			auth.POST("/roles", controllers.CreateRole)
			auth.PUT("/roles/:id", controllers.UpdateRole)
			auth.DELETE("/roles/:id", controllers.DeleteRole)
			auth.POST("/roles/:id/permissions", controllers.AssignPermissions)

			// 权限管理
			auth.GET("/permissions", controllers.GetPermissions)
			auth.GET("/permissions/:id", controllers.GetPermission)
			auth.POST("/permissions", controllers.CreatePermission)
			auth.PUT("/permissions/:id", controllers.UpdatePermission)
			auth.DELETE("/permissions/:id", controllers.DeletePermission)

			// 资源管理
			auth.GET("/resources", controllers.GetResources)
			auth.GET("/resources/:id", controllers.GetResource)
			auth.POST("/resources", controllers.CreateResource)
			auth.PUT("/resources/:id", controllers.UpdateResource)
			auth.DELETE("/resources/:id", controllers.DeleteResource)

			// 消息公告
			auth.GET("/notices", controllers.GetUserNotices)
			auth.GET("/notices/:id", controllers.GetNotice)
			auth.POST("/notices", controllers.CreateNotice)
			auth.PUT("/notices/:id", controllers.UpdateNotice)
			auth.DELETE("/notices/:id", controllers.DeleteNotice)
			auth.PUT("/notices/:id/read", controllers.MarkNoticeRead)

			// 日志管理
			auth.GET("/logs", controllers.GetLogs)
			auth.GET("/logs/options", controllers.GetLogOptions)
			auth.GET("/logs/:id", controllers.GetLog)

			// IP访问统计
			auth.GET("/ip-accesses", controllers.GetIPAccesses)
			auth.GET("/ip-statistics", controllers.GetIPStatistics)

			// 字典管理
			auth.GET("/dict-types", controllers.GetDictTypes)
			auth.GET("/dict-types/:id", controllers.GetDictType)
			auth.POST("/dict-types", controllers.CreateDictType)
			auth.PUT("/dict-types/:id", controllers.UpdateDictType)
			auth.DELETE("/dict-types/:id", controllers.DeleteDictType)

			auth.GET("/dict-items", controllers.GetDictItems)
			auth.GET("/dict-items/:id", controllers.GetDictItem)
			auth.POST("/dict-items", controllers.CreateDictItem)
			auth.PUT("/dict-items/:id", controllers.UpdateDictItem)
			auth.DELETE("/dict-items/:id", controllers.DeleteDictItem)

			// 系统参数管理
			auth.GET("/configs", controllers.GetConfigs)
			auth.GET("/configs/groups", controllers.GetConfigGroups)
			auth.GET("/configs/:id", controllers.GetConfig)
			auth.POST("/configs", controllers.CreateConfig)
			auth.PUT("/configs/:id", controllers.UpdateConfig)
			auth.DELETE("/configs/:id", controllers.DeleteConfig)

			// 部门管理
			auth.GET("/departments", controllers.ListDepartments)
			auth.GET("/departments/:id", controllers.GetDepartment)
			auth.POST("/departments", controllers.CreateDepartment)
			auth.PUT("/departments/:id", controllers.UpdateDepartment)
			auth.DELETE("/departments/:id", controllers.DeleteDepartment)

			// 系统监控
			auth.GET("/system/metrics", controllers.GetSystemMetrics)

			// 订单附件
			auth.GET("/orders/:id/attachments", controllers.ListOrderAttachments)
			auth.POST("/orders/:id/attachments", controllers.UploadOrderAttachment)
			auth.POST("/orders/:id/attachments/link", controllers.AttachOrderMaterial)
			auth.POST("/orders/batch-attachments", controllers.BatchUploadOrderAttachments)
			auth.GET("/orders/:id/attachments/:attachmentId/download", controllers.DownloadOrderAttachment)
			auth.DELETE("/orders/:id/attachments/:attachmentId", controllers.DeleteOrderAttachment)

			// 订单管理
			auth.GET("/orders", controllers.ListOrders)
			auth.POST("/orders", controllers.CreateOrder)
			auth.PUT("/orders/:id", controllers.UpdateOrder)
			auth.DELETE("/orders/:id", controllers.DeleteOrder)
			auth.POST("/orders/import", controllers.ImportOrders)
			auth.GET("/orders/export", controllers.ExportOrders)

			// 存储设置
			auth.GET("/storage/settings", controllers.GetStorageSettings)
			auth.PUT("/storage/settings", controllers.UpdateStorageSettings)

			// 素材图库
			auth.GET("/material-folders", controllers.ListMaterialFolders)
			auth.POST("/material-folders", controllers.CreateMaterialFolder)
			auth.PUT("/material-folders/:id", controllers.UpdateMaterialFolder)
			auth.DELETE("/material-folders/:id", controllers.DeleteMaterialFolder)

			auth.GET("/materials", controllers.ListMaterials)
			auth.GET("/materials/:id", controllers.GetMaterial)
			auth.POST("/materials", controllers.CreateMaterial)
			auth.PUT("/materials/:id", controllers.UpdateMaterial)
			auth.DELETE("/materials/:id", controllers.DeleteMaterial)
			auth.POST("/materials/upload", controllers.UploadMaterial)
			auth.GET("/materials/:id/download", controllers.DownloadMaterial)
		}
	}
}
