package models

import (
	"time"

	"gorm.io/gorm"
)

// Notice 消息公告
type Notice struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Title     string `json:"title" gorm:"type:varchar(255);not null"`      // 标题
	Content   string `json:"content" gorm:"type:text"`                     // 内容
	Type      string `json:"type" gorm:"type:varchar(50);default:'notice'"`  // 类型：notice-公告, message-消息
	Status    int    `json:"status" gorm:"default:1"`                        // 状态：1-发布，0-草稿
	Priority  int    `json:"priority" gorm:"default:0"`                      // 优先级：0-普通，1-重要，2-紧急
	CreatedBy uint   `json:"created_by"`                                     // 创建人ID
}

// NoticeRead 消息阅读记录
type NoticeRead struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	NoticeID uint `json:"notice_id" gorm:"index;not null"` // 消息ID
	UserID   uint `json:"user_id" gorm:"index;not null"`   // 用户ID
	IsRead   int  `json:"is_read" gorm:"default:0"`         // 是否已读：1-已读，0-未读
	ReadAt   *time.Time `json:"read_at"`                    // 阅读时间
}
