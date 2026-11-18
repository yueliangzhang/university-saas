package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/router"
	universityRouter "github.com/flipped-aurora/gin-vue-admin/server/router/university"
	"github.com/gin-gonic/gin"
)

func holder(routers ...*gin.RouterGroup) {
	_ = routers
	_ = router.RouterGroupApp
}

func initBizRouter(routers ...*gin.RouterGroup) {
	privateGroup := routers[0]
	publicGroup := routers[1]

	holder(publicGroup, privateGroup)

	universityRouter.InitTenantRouter(privateGroup)
	universityRouter.InitAdminRouter(privateGroup)
}
