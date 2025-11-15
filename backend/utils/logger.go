package utils

import (
	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"net/http"
	"time"
)

func LogAction(userID uint, username, action, module, content, ip, userAgent string, status int) {
	log := models.Log{
		UserID:    userID,
		Username:  username,
		Action:    action,
		Module:    module,
		Content:   content,
		IP:        ip,
		UserAgent: userAgent,
		Status:    status,
	}
	database.DB.Create(&log)
}

func RecordIPAccess(ip string, r *http.Request) {
	today := time.Now().Format("2006-01-02")
	date, _ := time.Parse("2006-01-02", today)

	var ipAccess models.IPAccess
	result := database.DB.Where("ip = ? AND date = ?", ip, date).First(&ipAccess)

	if result.Error != nil {
		// 创建新记录，查询IP地理位置信息
		country, province, city, isp := GetIPGeoInfoSafe(ip)
		
		ipAccess = models.IPAccess{
			IP:          ip,
			Date:        date,
			Country:     country,
			Province:    province,
			City:        city,
			ISP:         isp,
			AccessCount: 1,
		}
		database.DB.Create(&ipAccess)
	} else {
		// 如果地理位置信息为空，尝试更新
		if ipAccess.Country == "" || ipAccess.Country == "未知" {
			country, province, city, isp := GetIPGeoInfoSafe(ip)
			ipAccess.Country = country
			ipAccess.Province = province
			ipAccess.City = city
			ipAccess.ISP = isp
		}
		// 更新访问次数
		ipAccess.AccessCount++
		database.DB.Save(&ipAccess)
	}
}

