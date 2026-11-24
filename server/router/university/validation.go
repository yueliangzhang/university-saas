package university

import (
	"net/mail"

	"gorm.io/gorm"
)

func universityValueExists(db *gorm.DB, table, field string, tenantID, excludeID uint64, value string) (bool, error) {
	query := db.Table(table).Where(map[string]interface{}{"tenant_id": tenantID, field: value})
	if excludeID != 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

func validUniversityEmail(value string) bool {
	if value == "" {
		return true
	}
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}
