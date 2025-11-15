package models

import (
	"time"

	"gorm.io/gorm"
)

// OrderInfo 订单信息
type OrderInfo struct {
	ID                     uint64            `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt              time.Time         `json:"created_at"`
	UpdatedAt              time.Time         `json:"updated_at"`
	DeletedAt              gorm.DeletedAt    `json:"-" gorm:"index"`
	GSPOrderNo             string            `json:"gsp_order_no" gorm:"size:32;not null"`
	OrderType              string            `json:"order_type" gorm:"size:32;not null;default:platform"`
	OrderCreatedAt         time.Time         `json:"order_created_at" gorm:"primaryKey"`
	Status                 int8              `json:"status"`
	PaymentTime            *time.Time        `json:"payment_time"`
	CompletedAt            *time.Time        `json:"completed_at"`
	ShippingWarehouseCode  string            `json:"shipping_warehouse_code" gorm:"size:16"`
	RequiredSignAt         *time.Time        `json:"required_sign_at"`
	ShopCode               string            `json:"shop_code" gorm:"size:32"`
	ProductID              string            `json:"product_id" gorm:"size:32"`
	OwnerName              string            `json:"owner_name" gorm:"size:64"`
	ProductName            string            `json:"product_name" gorm:"size:128"`
	Spec                   string            `json:"spec" gorm:"size:64"`
	ItemNo                 string            `json:"item_no" gorm:"size:64"`
	SellerSKU              string            `json:"seller_sku" gorm:"size:64"`
	PlatformSKU            string            `json:"platform_sku" gorm:"size:64"`
	PlatformSKC            string            `json:"platform_skc" gorm:"size:64"`
	PlatformSPU            string            `json:"platform_spu" gorm:"size:64"`
	ProductPrice           float64           `json:"product_price"`
	ExpectedRevenue        float64           `json:"expected_revenue"`
	SpecialProductNote     string            `json:"special_product_note" gorm:"size:200"`
	CurrencyCode           string            `json:"currency_code" gorm:"size:16"`
	ExpectedFulfillmentQty int               `json:"expected_fulfillment_qty"`
	ItemCount              int               `json:"item_count"`
	PostalCode             string            `json:"postal_code" gorm:"size:20"`
	Country                string            `json:"country" gorm:"size:64"`
	Province               string            `json:"province" gorm:"size:64"`
	City                   string            `json:"city" gorm:"size:64"`
	District               string            `json:"district" gorm:"size:64"`
	AddressLine1           string            `json:"address_line1" gorm:"size:128"`
	AddressLine2           string            `json:"address_line2" gorm:"size:128"`
	CustomerFullName       string            `json:"customer_full_name" gorm:"size:128"`
	CustomerLastName       string            `json:"customer_last_name" gorm:"size:64"`
	CustomerFirstName      string            `json:"customer_first_name" gorm:"size:64"`
	PhoneNumber            string            `json:"phone_number" gorm:"size:32"`
	Email                  string            `json:"email" gorm:"size:128"`
	TaxNumber              string            `json:"tax_number" gorm:"size:64"`
	CreatedBy              uint64            `json:"created_by" gorm:"column:created_by"`
	UpdatedBy              uint64            `json:"updated_by" gorm:"column:updated_by"`
	Attachments            []OrderAttachment `json:"attachments" gorm:"foreignKey:OrderID"`
}

// OrderAttachment 订单附件
type OrderAttachment struct {
	ID         uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	OrderID    uint64         `json:"order_id" gorm:"index;not null"`
	FileType   string         `json:"file_type" gorm:"size:32;not null"`
	FileName   string         `json:"file_name" gorm:"size:255;not null"`
	FilePath   string         `json:"file_path" gorm:"size:512;not null"`
	FileExt    string         `json:"file_ext" gorm:"size:16"`
	FileSize   int64          `json:"file_size"`
	Checksum   string         `json:"checksum" gorm:"size:128"`
	Storage    string         `json:"storage" gorm:"size:16;not null;default:local"`
	UploaderID uint           `json:"uploader_id" gorm:"index"`
	MaterialID *uint64        `json:"material_id" gorm:"index"`
	Material   *MaterialAsset `json:"material,omitempty" gorm:"foreignKey:MaterialID;references:ID"`
}

func (OrderInfo) TableName() string {
	return "order_info"
}
