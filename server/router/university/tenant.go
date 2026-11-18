package university

import (
	tenantApi "github.com/flipped-aurora/gin-vue-admin/server/api/v1/university"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

func InitTenantRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("uni/tenant")
	group.Use(middleware.PlatformAdmin())

	api := new(tenantApi.TenantApi)

	group.POST("create", api.Create)
	group.POST("list", api.List)
	group.GET("detail", api.Detail)
	group.PUT("update", api.Update)
	group.PUT("status", api.SetStatus)
}
