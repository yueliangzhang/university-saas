package university

import "time"

type Membership struct {
	ID        uint64    `json:"id" gorm:"primaryKey;column:id"`
	TenantID  uint64    `json:"tenantId" gorm:"column:tenant_id"`
	UserID    uint64    `json:"userId" gorm:"column:user_id"`
	Role      string    `json:"role" gorm:"column:role"`
	Status    int8      `json:"status" gorm:"column:status"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (Membership) TableName() string { return "uni_memberships" }
