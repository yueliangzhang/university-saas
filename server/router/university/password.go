package university

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func provisionAccount(tx *gorm.DB, role string, tid, id uint64) error {
	hash, e := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if e != nil {
		return e
	}
	return tx.Table("uni_" + role + "_accounts").Create(map[string]interface{}{"tenant_id": tid, role + "_id": id, "password_hash": string(hash), "status": 1, "must_change_password": true}).Error
}
func recordLoginFailure(role string, tid, id uint64, hash string) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var row struct {
			FailedAttempts int
			LockedUntil    *time.Time
			PasswordHash   string
		}
		q := tx.Table("uni_"+role+"_accounts").Where("tenant_id=? AND id=?", tid, id)
		if e := q.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&row).Error; e != nil {
			return e
		}
		if row.PasswordHash != hash {
			return nil
		}
		now := time.Now()
		if row.LockedUntil != nil {
			if row.LockedUntil.After(now) {
				return nil
			}
			row.FailedAttempts = 0
		}
		row.FailedAttempts++
		var until *time.Time
		if row.FailedAttempts >= 5 {
			t := now.Add(15 * time.Minute)
			until = &t
		}
		return q.Updates(map[string]interface{}{"failed_attempts": row.FailedAttempts, "locked_until": until}).Error
	})
}
