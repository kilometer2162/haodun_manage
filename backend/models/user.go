package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Username string `json:"username" gorm:"type:varchar(255);uniqueIndex;not null"`
	Password string `json:"-" gorm:"not null"`
	Email    string `json:"email"`
	Status   int    `json:"status" gorm:"default:1"` // 1:正常 0:禁用

	RoleID       uint        `json:"role_id"`
	Role         Role        `json:"role" gorm:"foreignKey:RoleID"`
	DepartmentID uint        `json:"department_id"`
	Department   *Department `json:"department,omitempty" gorm:"foreignKey:DepartmentID"`
	EmployeeType string      `json:"employee_type" gorm:"size:32;not null;default:internal"`
}
