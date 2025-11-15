package controllers

import (
	"haodun_manage/backend/database"
	"haodun_manage/backend/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetNotices 获取消息公告列表
func GetNotices(c *gin.Context) {
	var notices []models.Notice
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	title := c.Query("title")
	isRead := c.Query("is_read")
	userID, _ := c.Get("user_id")

	query := database.DB.Model(&models.Notice{}).Where("status = 1")
	
	// 标题搜索
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}

	// 如果指定了阅读状态，需要关联查询
	if isRead != "" {
		readStatus, _ := strconv.Atoi(isRead)
		var noticeIDs []uint
		readQuery := database.DB.Model(&models.NoticeRead{}).Where("user_id = ? AND is_read = ?", userID, readStatus)
		readQuery.Pluck("notice_id", &noticeIDs)
		
		if len(noticeIDs) > 0 {
			query = query.Where("id IN ?", noticeIDs)
		} else if readStatus == 1 {
			// 如果查询已读但没有记录，返回空
			query = query.Where("1 = 0")
		}
	}

	var total int64
	query.Count(&total)
	query.Order("priority DESC, created_at DESC").Offset(offset).Limit(pageSize).Find(&notices)

	// 获取阅读状态
	if userID != nil {
		var readRecords []models.NoticeRead
		var noticeIDList []uint
		for _, notice := range notices {
			noticeIDList = append(noticeIDList, notice.ID)
		}
		if len(noticeIDList) > 0 {
			database.DB.Where("user_id = ? AND notice_id IN ?", userID, noticeIDList).Find(&readRecords)
			readMap := make(map[uint]bool)
			for _, record := range readRecords {
				if record.IsRead == 1 {
					readMap[record.NoticeID] = true
				}
			}
	for i := range notices {
				if readMap[notices[i].ID] {
					notices[i].ID = notices[i].ID // 这里可以通过扩展字段来标记，暂时用ID占位
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      notices,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetNotice 获取单个消息公告
func GetNotice(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var notice models.Notice
	if err := database.DB.First(&notice, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "消息不存在"})
		return
	}

	// 标记为已读
	if userID != nil {
		var noticeRead models.NoticeRead
		if err := database.DB.Where("notice_id = ? AND user_id = ?", id, userID).First(&noticeRead).Error; err != nil {
		// 创建阅读记录
			now := time.Now()
			noticeRead = models.NoticeRead{
			NoticeID: notice.ID,
				UserID:   userID.(uint),
				IsRead:   1,
				ReadAt:   &now,
		}
			database.DB.Create(&noticeRead)
		} else if noticeRead.IsRead == 0 {
			// 更新为已读
			now := time.Now()
			noticeRead.IsRead = 1
			noticeRead.ReadAt = &now
			database.DB.Save(&noticeRead)
		}
	}

	c.JSON(http.StatusOK, notice)
}

// CreateNotice 创建消息公告
func CreateNotice(c *gin.Context) {
	var notice models.Notice
	if err := c.ShouldBindJSON(&notice); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	if userID != nil {
		notice.CreatedBy = userID.(uint)
	}

	if err := database.DB.Create(&notice).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建消息失败"})
		return
	}

	c.JSON(http.StatusOK, notice)
}

// UpdateNotice 更新消息公告
func UpdateNotice(c *gin.Context) {
	id := c.Param("id")
	var notice models.Notice
	if err := database.DB.First(&notice, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "消息不存在"})
		return
	}

	var req models.Notice
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notice.Title = req.Title
	notice.Content = req.Content
	notice.Type = req.Type
	notice.Status = req.Status
	notice.Priority = req.Priority

	database.DB.Save(&notice)
	c.JSON(http.StatusOK, notice)
}

// DeleteNotice 删除消息公告
func DeleteNotice(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&models.Notice{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// MarkNoticeRead 标记消息为已读
func MarkNoticeRead(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var noticeRead models.NoticeRead
	noticeID, _ := strconv.ParseUint(id, 10, 32)
	if err := database.DB.Where("notice_id = ? AND user_id = ?", noticeID, userID).First(&noticeRead).Error; err != nil {
		// 创建阅读记录
		now := time.Now()
		noticeRead = models.NoticeRead{
			NoticeID: uint(noticeID),
			UserID:   userID.(uint),
			IsRead:   1,
			ReadAt:   &now,
		}
		database.DB.Create(&noticeRead)
	} else {
		// 更新为已读
		now := time.Now()
		noticeRead.IsRead = 1
		noticeRead.ReadAt = &now
		database.DB.Save(&noticeRead)
	}

	c.JSON(http.StatusOK, gin.H{"message": "标记成功"})
}

// GetUserNotices 获取当前用户的消息列表（包含阅读状态）
func GetUserNotices(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	title := c.Query("title")
	isRead := c.Query("is_read")

	query := database.DB.Table("notices").
		Select("notices.*, COALESCE(notice_reads.is_read, 0) as is_read").
		Joins("LEFT JOIN notice_reads ON notices.id = notice_reads.notice_id AND notice_reads.user_id = ?", userID).
		Where("notices.status = 1 AND notices.deleted_at IS NULL")

	if title != "" {
		query = query.Where("notices.title LIKE ?", "%"+title+"%")
	}

	if isRead != "" {
		readStatus, _ := strconv.Atoi(isRead)
		if readStatus == 1 {
			query = query.Where("notice_reads.is_read = 1")
		} else {
			query = query.Where("(notice_reads.is_read = 0 OR notice_reads.is_read IS NULL)")
		}
	}

	var results []map[string]interface{}
	var total int64
	query.Count(&total)
	query.Order("notices.priority DESC, notices.created_at DESC").Offset(offset).Limit(pageSize).Find(&results)

	c.JSON(http.StatusOK, gin.H{
		"data":      results,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
