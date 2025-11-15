package models

import (
	"time"

	"gorm.io/gorm"
)

// DictType 字典类型
type DictType struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Code        string     `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"` // 字典类型代码，如：user_status
	Name        string     `json:"name" gorm:"type:varchar(255);not null"`             // 字典类型名称，如：用户状态
	Description string     `json:"description"`                                        // 描述
	Status      int        `json:"status" gorm:"default:1"`                            // 状态：1-启用，0-禁用
	Items       []DictItem `json:"items" gorm:"foreignKey:TypeID"`                     // 字典项列表
}

// DictItem 字典项
type DictItem struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	TypeID      uint   `json:"type_id" gorm:"index;not null"`                          // 字典类型ID
	Label       string `json:"label" gorm:"type:varchar(255);not null"`               // 显示标签，如：正常
	Value       string `json:"value" gorm:"type:varchar(255);not null"`               // 值，如：1
	Sort        int    `json:"sort" gorm:"default:0"`                                  // 排序
	Status      int    `json:"status" gorm:"default:1"`                                // 状态：1-启用，0-禁用
	Description string `json:"description"`                                           // 描述
	Type        DictType `json:"type" gorm:"foreignKey:TypeID"`                       // 关联的字典类型
}

