package university

import (
	"net/http"

	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

func InitWorkspaceRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("uni/workspace").
		Use(middleware.TenantAuth())

	group.GET("context", func(c *gin.Context) {
		tenantID, ok := middleware.GetTenantID(c)
		if !ok {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"tenantId": tenantID,
				"role":     "tenant_admin",
			},
			"msg": "成功",
		})
	})
}
