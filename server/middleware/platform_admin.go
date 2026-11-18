package middleware

import (
	"net/http"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

// PlatformAdmin 仅允许当前开发环境中的平台管理员访问。
// 后续应替换为独立的平台角色授权模型。
func PlatformAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := utils.GetUserID(c)
		authorityID := utils.GetUserAuthorityId(c)

		if userID != 1 || authorityID != 888 || global.GVA_DB == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": 7,
				"msg":  "仅平台管理员可以操作",
			})
			return
		}

		var count int64
		err := global.GVA_DB.Table("sys_users").
			Where("id = ? AND authority_id = ? AND enable = ?", userID, 888, 1).
			Count(&count).Error

		if err != nil || count != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": 7,
				"msg":  "平台管理员身份验证失败",
			})
			return
		}

		c.Next()
	}
}
