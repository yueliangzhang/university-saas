package university

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	universityModel "github.com/flipped-aurora/gin-vue-admin/server/model/university"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const tenantAdminAuthorityID uint = 2001

var (
	ErrTenantNotActive = errors.New("高校不存在或已停用")
	ErrUsernameTaken   = errors.New("用户名已存在")
	ErrMemberNotFound  = errors.New("高校管理员成员不存在")
)

type AdminService struct{}

type AdminListItem struct {
	MembershipID uint64 `json:"membershipId" gorm:"column:membership_id"`
	TenantID     uint64 `json:"tenantId" gorm:"column:tenant_id"`
	UserID       uint64 `json:"userId" gorm:"column:user_id"`
	Username     string `json:"username" gorm:"column:username"`
	NickName     string `json:"nickName" gorm:"column:nick_name"`
	UserEnabled  int    `json:"userEnabled" gorm:"column:user_enabled"`
	Status       int8   `json:"status" gorm:"column:status"`
}

// CreateAdmin is atomic: sys_users + sys_user_authority + uni_memberships.
func (AdminService) CreateAdmin(tenantID uint64, username, password, nickName string) (uint64, error) {
	username = strings.TrimSpace(username)
	nickName = strings.TrimSpace(nickName)
	if tenantID == 0 || username == "" || len(username) > 191 || len(password) < 12 || len(password) > 72 || len(nickName) > 191 {
		return 0, errors.New("参数不合法：密码须为12至72字节")
	}
	var userID uint64
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var tenant struct {
			ID     uint64
			Status int8
		}
		if err := tx.Table("uni_tenants").Select("id, status").Where("id = ? AND status = 1", tenantID).Take(&tenant).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTenantNotActive
			}
			return err
		}
		var count int64
		if err := tx.Table("sys_users").Where("username = ?", username).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrUsernameTaken
		}
		var roleCount int64
		if err := tx.Table("sys_authorities").Where("authority_id = ? AND deleted_at IS NULL", tenantAdminAuthorityID).Count(&roleCount).Error; err != nil {
			return err
		}
		if roleCount != 1 {
			return errors.New("高校管理员系统角色2001不存在")
		}

		// Map inserts deliberately avoid GVA model hooks/associations; role is explicitly fixed server-side.
		user := map[string]interface{}{
			"uuid": uuid.New().String(), "username": username,
			"password": utils.BcryptHash(password), "nick_name": nickName,
			"authority_id": tenantAdminAuthorityID, "enable": 1,
		}
		if err := tx.Table("sys_users").Create(user).Error; err != nil {
			// DB unique index is authoritative, including concurrent requests.
			if isDuplicateKey(err) {
				return ErrUsernameTaken
			}
			return err
		}
		// Map insert does not populate auto-increment ID; look it up inside the transaction.
		var created struct {
			ID uint64 `gorm:"column:id"`
		}
		if err := tx.Table("sys_users").Select("id").Where("username = ?", username).Take(&created).Error; err != nil {
			return err
		}
		userID = created.ID
		if err := tx.Table("sys_user_authority").Create(map[string]interface{}{
			"sys_user_id": userID, "sys_authority_authority_id": tenantAdminAuthorityID,
		}).Error; err != nil {
			return err
		}
		return tx.Create(&universityModel.Membership{
			TenantID: tenantID, UserID: userID, Role: "tenant_admin", Status: 1,
		}).Error
	})
	return userID, err
}

func isDuplicateKey(err error) bool {
	// MySQL duplicate entry code 1062; avoid coupling to a particular mysql driver package.
	return err != nil && strings.Contains(err.Error(), "Error 1062")
}

func (AdminService) ListAdmins(tenantID uint64, page, pageSize int) ([]AdminListItem, int64, error) {
	if tenantID == 0 || page < 1 || pageSize < 1 || pageSize > 100 || page > 1000000 {
		return nil, 0, errors.New("分页参数不合法")
	}
	q := global.GVA_DB.Table("uni_memberships AS m").
		Joins("JOIN sys_users AS u ON u.id = m.user_id AND u.deleted_at IS NULL").
		Where("m.tenant_id = ? AND m.role = ?", tenantID, "tenant_admin")
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []AdminListItem
	err := q.Select("m.id AS membership_id, m.tenant_id, m.user_id, u.username, u.nick_name, u.enable AS user_enabled, m.status").
		Order("m.id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Scan(&rows).Error
	return rows, total, err
}

func (AdminService) SetAdminStatus(tenantID, membershipID uint64, status int8) error {
	if tenantID == 0 || membershipID == 0 || (status != 1 && status != 2) {
		return errors.New("参数不合法")
	}
	res := global.GVA_DB.Model(&universityModel.Membership{}).
		Where("id = ? AND tenant_id = ? AND role = ?", membershipID, tenantID, "tenant_admin").
		Update("status", status)
	if res.Error != nil {
		return fmt.Errorf("更新管理员状态失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		var count int64
		if err := global.GVA_DB.Model(&universityModel.Membership{}).Where("id = ? AND tenant_id = ? AND role = ?", membershipID, tenantID, "tenant_admin").Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrMemberNotFound
		}
	}
	return nil
}
