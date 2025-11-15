package models

import (
	"time"

	"gorm.io/gorm"
)

// MaterialFolder represents a hierarchical folder for materials.
type MaterialFolder struct {
	ID        uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name     string  `json:"name" gorm:"size:128;not null"`
	ParentID *uint64 `json:"parent_id" gorm:"index"`
	Path     string  `json:"path" gorm:"size:512"`

	Children []MaterialFolder `json:"children" gorm:"foreignKey:ParentID"`
}

func (MaterialFolder) TableName() string {
	return "material_folders"
}

// MaterialAsset stores information about a material library entry.
type MaterialAsset struct {
	ID        uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Code        string `json:"code" gorm:"size:64;uniqueIndex;not null"`
	FileName    string `json:"file_name" gorm:"size:255;not null"`
	Title       string `json:"title" gorm:"size:255"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Dimensions  string `json:"dimensions" gorm:"size:64"`
	Format      string `json:"format" gorm:"size:32"`
	FileSize    int64  `json:"file_size"`
	Storage     string `json:"storage" gorm:"size:16;not null;default:local"`
	FilePath    string `json:"file_path" gorm:"size:512"`
	DownloadURL string `json:"download_url" gorm:"-"`
	PreviewURL  string `json:"preview_url" gorm:"-"`

	CreatedBy  uint64  `json:"created_by" gorm:"index"`
	UpdatedBy  uint64  `json:"updated_by" gorm:"index"`
	OrderCount int     `json:"order_count"`
	FolderID   *uint64 `json:"folder_id" gorm:"index"`
	Shape      string  `json:"shape" gorm:"size:32"`

	Folder        *MaterialFolder `json:"folder,omitempty" gorm:"foreignKey:FolderID;references:ID"`
	CreatedByName string          `json:"created_by_name,omitempty" gorm:"-"`
	UpdatedByName string          `json:"updated_by_name,omitempty" gorm:"-"`
}

func (MaterialAsset) TableName() string {
	return "material_assets"
}
