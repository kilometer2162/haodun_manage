package models

import (
	"time"

	"gorm.io/gorm"
)

type Log struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	UserID    uint   `json:"user_id" gorm:"index"`
	Username  string `json:"username" gorm:"type:varchar(255);index"`
	Action    string `json:"action" gorm:"type:varchar(255);index"`    // 操作类型
	Module    string `json:"module" gorm:"type:varchar(255);index"`     // 模块
	Content   string `json:"content"`    // 操作内容
	IP        string `json:"ip"`         // IP地址
	UserAgent string `json:"user_agent"` // 用户代理
	Status    int    `json:"status"`     // 状态 1:成功 0:失败
}

