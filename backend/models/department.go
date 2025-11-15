package models

import (
	"time"

	"gorm.io/gorm"
)

// Department 部门信息
type Department struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string      `json:"name" gorm:"size:128;uniqueIndex;not null"`
	ParentID    *uint       `json:"parent_id" gorm:"column:parent_id"`
	Description string      `json:"description" gorm:"size:255"`
	Status      int         `json:"status" gorm:"default:1"` // 1:启用 0:禁用
	Sort        int         `json:"sort" gorm:"default:0"`
	Parent      *Department `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
}
