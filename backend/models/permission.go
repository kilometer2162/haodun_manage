package models

import (
	"time"

	"gorm.io/gorm"
)

type Permission struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string `json:"name" gorm:"type:varchar(255);uniqueIndex;not null"`
	Code        string `json:"code" gorm:"type:varchar(255);uniqueIndex;not null"`
	Description string `json:"description"`

	ResourceID uint     `json:"resource_id"`
	Resource   Resource `json:"resource" gorm:"foreignKey:ResourceID"`
}

