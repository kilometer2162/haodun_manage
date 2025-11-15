package models

import (
	"time"

	"gorm.io/gorm"
)

type Resource struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string     `json:"name" gorm:"type:varchar(255);not null"`
	Path        string     `json:"path" gorm:"type:varchar(255);uniqueIndex;not null"`
	Method      string     `json:"method" gorm:"not null"` // GET, POST, PUT, DELETE
	Description string     `json:"description"`
	Type        string     `json:"type"` // api, menu, button
	ParentID    *uint      `json:"parent_id" gorm:"index"` // 父资源ID，nil表示顶级资源
	Parent      *Resource  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children    []Resource `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Sort        int        `json:"sort" gorm:"default:0"` // 排序字段
	Icon        string     `json:"icon"`                  // 图标
}

