package university

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	systemAPI "github.com/flipped-aurora/gin-vue-admin/server/api/v1/system"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemService "github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type portalLoginAccount struct {
	ID                 uint64
	TenantID           uint64
	ProfileID          uint64
	MustChangePassword bool
	LockedUntil        *time.Time
	PasswordHash       string
}

func (account portalLoginAccount) locked() bool {
	return account.LockedUntil != nil && account.LockedUntil.After(time.Now())
}

func (account portalLoginAccount) matches(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)) == nil
}

// role is selected from the fixed student and teacher role names.
func findPortalLogin(role, code string, tenantID uint64, number string) (portalLoginAccount, error) {
	var account portalLoginAccount
	query := global.GVA_DB.Table("uni_"+role+"_accounts a").
		Select("a.id,a.tenant_id,a."+role+"_id AS profile_id,a.password_hash,a.must_change_password,a.locked_until").
		Joins("JOIN uni_"+role+"s p ON p.tenant_id=a.tenant_id AND p.id=a."+role+"_id").
		Joins("JOIN uni_tenants t ON t.id=a.tenant_id").
		Where("p."+role+"_no=? AND a.status=1 AND p.status=1 AND t.status=1", strings.TrimSpace(number))
	if code != "" {
		query = query.Where("t.code=?", code)
	} else {
		query = query.Where("a.tenant_id=?", tenantID)
	}
	err := query.Take(&account).Error
	return account, err
}

func issuePortalLogin(role string, row portalLoginAccount) (gin.H, error) {
	accounts := "uni_" + role + "_accounts"
	sessions := "uni_" + role + "_sessions"
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(random)
	expires := time.Now().Add(12 * time.Hour)
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var current struct {
			PasswordHash string
			LockedUntil  *time.Time
			Status       int
		}
		if e := tx.Table(accounts).Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id=? AND id=?", row.TenantID, row.ID).Take(&current).Error; e != nil {
			return e
		}
		if current.PasswordHash != row.PasswordHash || current.Status != 1 || (current.LockedUntil != nil && current.LockedUntil.After(time.Now())) {
			return gorm.ErrRecordNotFound
		}
		if e := tx.Table(sessions).Create(map[string]interface{}{"tenant_id": row.TenantID, "account_id": row.ID, "token_hash": studentHashToken(token), "expires_at": expires}).Error; e != nil {
			return e
		}
		return tx.Table(accounts).Where("id=? AND tenant_id=?", row.ID, row.TenantID).Updates(map[string]interface{}{"last_login_at": time.Now(), "failed_attempts": 0, "locked_until": nil}).Error
	})
	if err != nil {
		return nil, err
	}
	return gin.H{"mustChangePassword": row.MustChangePassword, "token": token, "expiresAt": expires}, nil
}

func unifiedPortalLogin(c *gin.Context) {
	var input struct {
		TenantCode string
		Account    string
		Password   string
		Role       string
	}
	if c.ShouldBindJSON(&input) != nil {
		studentAuthError(c, 400, "参数不合法")
		return
	}
	code, number := strings.TrimSpace(input.TenantCode), strings.TrimSpace(input.Account)
	if code == "" || len(code) > 64 || number == "" || len(number) > 64 || input.Password == "" || len(input.Password) > 256 ||
		(input.Role != "" && input.Role != "student" && input.Role != "teacher" && input.Role != "tenant_admin" && input.Role != "leave_monitor") {
		studentAuthError(c, 400, "参数不合法")
		return
	}
	base := systemAPI.BaseApi{}
	// Limit repeated failures independently of the platform captcha policy.
	failureKey := "university-login:" + c.ClientIP()
	if value, ok := global.BlackCache.Get(failureKey); ok {
		if count, ok := value.(int); ok && count >= 10 {
			studentAuthError(c, 429, "登录尝试过多，请稍后重试")
			return
		}
	}
	accounts := make(map[string]portalLoginAccount, 2)
	var matches []string
	var locked bool
	for _, role := range []string{"student", "teacher"} {
		if input.Role != "" && role != input.Role {
			continue
		}
		account, err := findPortalLogin(role, code, 0, number)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			studentAuthError(c, 500, "登录失败")
			return
		}
		if account.locked() {
			locked = true
			continue
		}
		accounts[role] = account
		if account.matches(input.Password) {
			matches = append(matches, role)
		}
	}

	var backendUser *system.SysUser
	if input.Role == "" || input.Role == "tenant_admin" || input.Role == "leave_monitor" {
		user, err := systemService.UserServiceApp.FindUniversityLoginUser(number, input.Password, code)
		if err != nil {
			studentAuthError(c, 500, "登录失败")
			return
		}
		if user != nil {
			role := systemService.UniversityLoginRole(user.AuthorityId)
			if input.Role == "" || input.Role == role {
				backendUser = user
				matches = append(matches, role)
			}
		}
	}

	switch len(matches) {
	case 0:
		_ = global.BlackCache.Add(failureKey, 0, 15*time.Minute)
		_ = global.BlackCache.Increment(failureKey, 1)
		for role, account := range accounts {
			if recordLoginFailure(role, account.TenantID, account.ID, account.PasswordHash) != nil {
				studentAuthError(c, 500, "登录失败")
				return
			}
		}
		if locked && len(accounts) == 0 {
			studentAuthError(c, 429, "登录暂时锁定，请15分钟后重试")
			return
		}
		studentAuthError(c, 401, "高校编号、账号或密码错误")
	case 1:
		if backendUser != nil {
			base.TokenNext(c, *backendUser)
			return
		}
		role := matches[0]
		data, err := issuePortalLogin(role, accounts[role])
		if err != nil {
			studentAuthError(c, 500, "创建会话失败")
			return
		}
		data["role"] = role
		studentAuthOK(c, data)
	default:
		studentAuthOK(c, gin.H{"roles": matches})
	}
}
