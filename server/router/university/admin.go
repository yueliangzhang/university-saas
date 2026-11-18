package university

import (
	universityApi "github.com/flipped-aurora/gin-vue-admin/server/api/v1/university"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

func InitAdminRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("uni/admin").Use(middleware.PlatformAdmin())
	api := new(universityApi.AdminApi)
	group.POST("create", api.Create)
	group.POST("list", api.List)
	group.PUT("status", api.SetStatus)
}
