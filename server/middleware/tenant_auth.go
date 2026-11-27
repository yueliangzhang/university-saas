package middleware

import (
	"net/http"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

const (
	TenantIDContextKey   = "universityTenantID"
	TenantRoleContextKey = "universityTenantRole"
)

type tenantMembershipRow struct {
	TenantID     uint64 `gorm:"column:tenant_id"`
	Role         string `gorm:"column:role"`
	MemberStatus int8   `gorm:"column:member_status"`
	TenantStatus int8   `gorm:"column:tenant_status"`
	UserEnable   int    `gorm:"column:user_enable"`
	AuthorityID  uint64 `gorm:"column:authority_id"`
}

// TenantAuth validates current membership on every request.
// This version intentionally requires one active tenant membership.
func TenantAuth() gin.HandlerFunc {
	return tenantAccountAuth(false)
}

func AdvisorTenantAuth() gin.HandlerFunc {
	return tenantAccountAuth(true)
}

func tenantAccountAuth(allowMonitor bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := utils.GetUserID(c)

		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 7,
				"msg":  "请先登录",
			})
			return
		}

		var memberships []tenantMembershipRow

		err := global.GVA_DB.
			Table("uni_memberships AS m").
			Select(`
				m.tenant_id,
				m.role,
				m.status AS member_status,
				t.status AS tenant_status,
				u.enable AS user_enable,
				u.authority_id
			`).
			Joins("JOIN uni_tenants AS t ON t.id = m.tenant_id").
			Joins("JOIN sys_users AS u ON u.id = m.user_id").
			Where("m.user_id = ? AND u.deleted_at IS NULL", userID).
			Find(&memberships).Error

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code": 7,
				"msg":  "租户权限验证失败",
			})
			return
		}

		if len(memberships) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": 7,
				"msg":  "无有效的单一高校成员身份",
			})
			return
		}

		member := memberships[0]
		validRole := member.Role == "tenant_admin" && member.AuthorityID == 2001
		if allowMonitor && member.Role == "leave_monitor" && member.AuthorityID == 2002 {
			validRole = true
		}

		if member.MemberStatus != 1 ||
			member.TenantStatus != 1 ||
			member.UserEnable != 1 ||
			!validRole ||
			uint64(utils.GetUserAuthorityId(c)) != member.AuthorityID {

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": 7,
				"msg":  "高校账号权限不足或已停用",
			})
			return
		}

		c.Set(TenantIDContextKey, member.TenantID)
		c.Set(TenantRoleContextKey, member.Role)

		c.Next()
	}
}

func GetTenantID(c *gin.Context) (uint64, bool) {
	value, exists := c.Get(TenantIDContextKey)
	if !exists {
		return 0, false
	}

	tenantID, ok := value.(uint64)
	return tenantID, ok && tenantID != 0
}
