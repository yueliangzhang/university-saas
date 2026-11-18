package university

import "time"

// Tenant 高校租户
type Tenant struct {
	ID        uint64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Code      string     `json:"code" gorm:"size:64;not null;uniqueIndex:uk_uni_tenants_code"`
	Name      string     `json:"name" gorm:"size:200;not null"`
	Status    int8       `json:"status" gorm:"not null;default:1"`
	Plan      string     `json:"plan" gorm:"size:32;not null;default:basic"`
	ExpiresAt *time.Time `json:"expiresAt" gorm:"column:expires_at"`
	CreatedAt time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

func (Tenant) TableName() string {
	return "uni_tenants"
}
