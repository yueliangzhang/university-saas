package university

import (
	"errors"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	systemModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	universityModel "github.com/flipped-aurora/gin-vue-admin/server/model/university"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const tenantAdminAuthorityID uint = 2001

var (
	ErrTenantNotActive    = errors.New("高校不存在或已停用")
	ErrUsernameTaken      = errors.New("用户名已存在")
	ErrAdvisorLink        = errors.New("请选择本校有效且未关联其他高校账号的教师工号")
	ErrInvalidAccountRole = errors.New("请选择高校管理员或请假监管员")
	ErrMemberNotFound     = errors.New("高校账号不存在")
)

type AdminService struct{}

type AdminListItem struct {
	MembershipID     uint64 `json:"membershipId" gorm:"column:membership_id"`
	TenantID         uint64 `json:"tenantId" gorm:"column:tenant_id"`
	UserID           uint64 `json:"userId" gorm:"column:user_id"`
	Username         string `json:"username" gorm:"column:username"`
	NickName         string `json:"nickName" gorm:"column:nick_name"`
	UserEnabled      int    `json:"userEnabled" gorm:"column:user_enabled"`
	Status           int8   `json:"status" gorm:"column:status"`
	AuthorityID      uint   `json:"authorityId"`
	AdvisorTeacherNo string `json:"advisorTeacherNo"`
	TenantName       string `json:"tenantName"`
	TenantStatus     int8   `json:"tenantStatus"`
	Phone            string `json:"phone"`
	Email            string `json:"email"`
}

// CreateAdmin is atomic: sys_users + sys_user_authority + uni_memberships.
func (AdminService) CreateAdmin(tenantID uint64, username, password, nickName string, contact ...string) (uint64, error) {
	return (AdminService{}).CreateAccount(tenantID, username, password, nickName, 2001, contact...)
}
func (s AdminService) CreateAccount(tenantID uint64, username, password, nickName string, authorityID uint, contact ...string) (uint64, error) {
	return s.CreateAccountWithAdvisor(tenantID, username, password, nickName, authorityID, "", contact...)
}

func (AdminService) CreateAccountWithAdvisor(tenantID uint64, username, password, nickName string, authorityID uint, advisorTeacherNo string, contact ...string) (uint64, error) {
	if authorityID == 0 {
		authorityID = 2001
	}
	role := universityAccountRole(authorityID)
	if role == "" {
		return 0, ErrInvalidAccountRole
	}
	username = strings.TrimSpace(username)
	nickName = strings.TrimSpace(nickName)
	if tenantID == 0 || username == "" || len(username) > 191 || len(nickName) > 191 {
		return 0, errors.New("账号参数不合法")
	}
	if err := utils.ValidateUniversityAccountPassword(password); err != nil {
		return 0, err
	}
	phone, email := "", ""
	if len(contact) == 2 {
		phone, email = contact[0], contact[1]
	}
	if len(phone) > 64 || len(email) > 191 || nickName == "" {
		return 0, errors.New("姓名或联系方式不合法")
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
		// Reclaim usernames from deleted university accounts without reviving their identity.
		var previous systemModel.SysUser
		err := tx.Unscoped().Select("id, authority_id, deleted_at").Where("username = ?", username).Take(&previous).Error
		switch {
		case err == nil:
			if !previous.DeletedAt.Valid || (previous.AuthorityId != 2001 && previous.AuthorityId != 2002) || previous.ID == 1 {
				return ErrUsernameTaken
			}
			result := tx.Unscoped().Model(&systemModel.SysUser{}).
				Where("id = ? AND username = ? AND deleted_at IS NOT NULL AND authority_id = ?", previous.ID, username, previous.AuthorityId).
				UpdateColumn("username", nil)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ErrUsernameTaken
			}
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return err
		}
		if err := checkUniversityAuthority(tx, authorityID); err != nil {
			return err
		}

		// Map inserts deliberately avoid GVA model hooks/associations; role is explicitly fixed server-side.
		user := map[string]interface{}{
			"uuid": uuid.New().String(), "username": username,
			"password": utils.BcryptHash(password), "nick_name": nickName,
			"authority_id": authorityID, "enable": 1, "phone": phone, "email": email, "created_at": time.Now(), "updated_at": time.Now(),
		}
		if err := tx.Table("sys_users").Create(user).Error; err != nil {
			// DB unique index is authoritative, including concurrent requests.
			if isDuplicateKey(err) || errors.Is(err, gorm.ErrDuplicatedKey) {
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
			"sys_user_id": userID, "sys_authority_authority_id": authorityID,
		}).Error; err != nil {
			return err
		}
		membership := universityModel.Membership{TenantID: tenantID, UserID: userID, Role: role, Status: 1}
		if err := tx.Create(&membership).Error; err != nil {
			return err
		}
		if strings.TrimSpace(advisorTeacherNo) != "" {
			return setAccountAdvisor(tx, tenantID, membership.ID, advisorTeacherNo)
		}
		return nil
	})
	return userID, err
}

func universityAccountRole(authorityID uint) string {
	switch authorityID {
	case 2001:
		return "tenant_admin"
	case 2002:
		return "leave_monitor"
	default:
		return ""
	}
}

func checkUniversityAuthority(tx *gorm.DB, authorityID uint) error {
	var count int64
	if err := tx.Table("sys_authorities").Where("authority_id = ? AND deleted_at IS NULL", authorityID).Count(&count).Error; err != nil {
		return err
	}
	if count != 1 {
		return ErrInvalidAccountRole
	}
	return nil
}

func isDuplicateKey(err error) bool {
	// MySQL duplicate entry code 1062; avoid coupling to a particular mysql driver package.
	return err != nil && strings.Contains(err.Error(), "Error 1062")
}

func (AdminService) ListAdmins(tenantID uint64, page, pageSize int, keywords ...string) ([]AdminListItem, int64, error) {
	if page < 1 || pageSize < 1 || pageSize > 100 || page > 1000000 {
		return nil, 0, errors.New("分页参数不合法")
	}
	q := global.GVA_DB.Table("uni_memberships AS m").
		Joins("JOIN sys_users AS u ON u.id = m.user_id AND u.deleted_at IS NULL").
		Joins("JOIN uni_tenants AS t ON t.id=m.tenant_id").
		Joins("LEFT JOIN uni_teachers AS advisor ON advisor.tenant_id=m.tenant_id AND advisor.id=m.advisor_teacher_id").Where("((m.role='tenant_admin' AND u.authority_id=2001) OR (m.role='leave_monitor' AND u.authority_id=2002))")
	if tenantID != 0 {
		q = q.Where("m.tenant_id=?", tenantID)
	}
	if len(keywords) > 0 && strings.TrimSpace(keywords[0]) != "" {
		kw := "%" + strings.TrimSpace(keywords[0]) + "%"
		q = q.Where("(u.username LIKE ? OR u.nick_name LIKE ?)", kw, kw)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []AdminListItem
	err := q.Select("m.id AS membership_id, m.tenant_id, m.user_id, u.username, u.nick_name, u.authority_id, u.enable AS user_enabled, m.status, t.name AS tenant_name, t.status AS tenant_status, COALESCE(advisor.teacher_no,'') AS advisor_teacher_no, COALESCE(u.phone,'') AS phone, COALESCE(u.email,'') AS email").
		Order("m.id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Scan(&rows).Error
	return rows, total, err
}

func (AdminService) SetAdminStatus(tenantID, membershipID uint64, status int8) error {
	if status != 1 && status != 2 {
		return errors.New("状态不合法")
	}
	return adminMutation(tenantID, membershipID, func(tx *gorm.DB, uid uint64) error {
		if err := tx.Table("sys_users").Where("id=?", uid).Update("enable", status).Error; err != nil {
			return err
		}
		return tx.Table("uni_memberships").Where("id=?", membershipID).Updates(map[string]interface{}{"status": status, "session_cutoff": time.Now().Unix(), "updated_at": time.Now()}).Error
	})
}
func adminMutation(tid, mid uint64, fn func(*gorm.DB, uint64) error) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var row struct{ UserID uint64 }
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Table("uni_memberships m").Select("m.user_id").Joins("JOIN sys_users u ON u.id=m.user_id AND u.deleted_at IS NULL").Where("m.id=? AND m.tenant_id=? AND ((m.role='tenant_admin' AND u.authority_id=2001) OR (m.role='leave_monitor' AND u.authority_id=2002)) AND u.id<>1", mid, tid).Take(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrMemberNotFound
			}
			return err
		}
		var count int64
		if err := tx.Table("uni_memberships").Where("user_id=?", row.UserID).Count(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return ErrMemberNotFound
		}
		return fn(tx, row.UserID)
	})
}
func (AdminService) Edit(tid, mid uint64, name, phone, email string, advisorTeacherNo *string, authorityID *uint) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 191 || len(phone) > 64 || len(email) > 191 {
		return errors.New("参数不合法")
	}
	if authorityID != nil && universityAccountRole(*authorityID) == "" {
		return ErrInvalidAccountRole
	}
	return adminMutation(tid, mid, func(tx *gorm.DB, uid uint64) error {
		if advisorTeacherNo != nil {
			if err := setAccountAdvisor(tx, tid, mid, *advisorTeacherNo); err != nil {
				return err
			}
		}
		now := time.Now()
		updates := map[string]interface{}{"nick_name": name, "phone": phone, "email": email, "updated_at": now}
		if authorityID != nil {
			var current systemModel.SysUser
			if err := tx.Select("authority_id").First(&current, uid).Error; err != nil {
				return err
			}
			if current.AuthorityId != *authorityID {
				if err := checkUniversityAuthority(tx, *authorityID); err != nil {
					return err
				}
				if err := tx.Delete(&systemModel.SysUserAuthority{}, "sys_user_id = ?", uid).Error; err != nil {
					return err
				}
				if err := tx.Create(&systemModel.SysUserAuthority{
					SysUserId: uint(uid), SysAuthorityAuthorityId: *authorityID,
				}).Error; err != nil {
					return err
				}
				if err := tx.Table("uni_memberships").Where("id = ?", mid).
					Updates(map[string]interface{}{
						"role": universityAccountRole(*authorityID), "session_cutoff": now.Unix(), "updated_at": now,
					}).Error; err != nil {
					return err
				}
				updates["authority_id"] = *authorityID
			}
		}
		return tx.Table("sys_users").Where("id=?", uid).Updates(updates).Error
	})
}

func (AdminService) Reset(tid, mid uint64, password string) error {
	if err := utils.ValidateUniversityAccountPassword(password); err != nil {
		return err
	}
	return adminMutation(tid, mid, func(tx *gorm.DB, uid uint64) error {
		if err := tx.Table("sys_users").Where("id=?", uid).Updates(map[string]interface{}{"password": utils.BcryptHash(password), "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		return tx.Table("uni_memberships").Where("id=?", mid).Update("session_cutoff", time.Now().Unix()).Error
	})
}

func setAccountAdvisor(tx *gorm.DB, tenantID, membershipID uint64, teacherNo string) error {
	var teacherID *uint64
	if number := strings.TrimSpace(teacherNo); number != "" {
		var teacher struct{ ID uint64 }
		err := tx.Table("uni_teachers").Select("id").
			Where("tenant_id=? AND status=1 AND teacher_no=?", tenantID, number).Take(&teacher).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAdvisorLink
		}
		if err != nil {
			return err
		}
		teacherID = &teacher.ID
	}
	err := tx.Table("uni_memberships").Where("id=? AND tenant_id=?", membershipID, tenantID).
		Update("advisor_teacher_id", teacherID).Error
	if isDuplicateKey(err) || errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrAdvisorLink
	}
	return err
}

func (AdminService) DeleteAccount(tenantID, membershipID uint64) error {
	return adminMutation(tenantID, membershipID, func(tx *gorm.DB, userID uint64) error {
		if err := tx.Delete(&universityModel.Membership{}, "id = ? AND tenant_id = ?", membershipID, tenantID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&systemModel.SysUserAuthority{}, "sys_user_id = ?", userID).Error; err != nil {
			return err
		}
		return tx.Delete(&systemModel.SysUser{}, "id = ?", userID).Error
	})
}
