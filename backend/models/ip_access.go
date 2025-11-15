package models

import (
	"time"

	"gorm.io/gorm"
)

type IPAccess struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	IP        string    `json:"ip" gorm:"type:varchar(255);index;not null"`
	Date      time.Time `json:"date" gorm:"index;type:date"`
	Country   string    `json:"country"`   // 国家
	Province  string    `json:"province"`  // 省份
	City      string    `json:"city"`      // 城市
	ISP       string    `json:"isp"`       // 运营商
	AccessCount int    `json:"access_count" gorm:"default:1"` // 访问次数
}

