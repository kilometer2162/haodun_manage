package models

import (
	"time"

	"gorm.io/gorm"
)

// Config 系统参数
type Config struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Key         string `json:"key" gorm:"type:varchar(255);uniqueIndex;not null"`   // 参数键，如：site_name
	Value       string `json:"value" gorm:"type:text"`                                // 参数值
	Label       string `json:"label" gorm:"type:varchar(255);not null"`              // 参数标签，如：网站名称
	Type        string `json:"type" gorm:"type:varchar(50);default:'text'"`          // 参数类型：text, number, boolean, textarea, select
	Group       string `json:"group" gorm:"type:varchar(100);default:'system'"`     // 参数分组：system, email, upload等
	Description string `json:"description"`                                           // 描述
	Sort        int    `json:"sort" gorm:"default:0"`                                // 排序
	Status      int    `json:"status" gorm:"default:1"`                               // 状态：1-启用，0-禁用
}

