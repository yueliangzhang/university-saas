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
	universityRouter.InitWorkspaceRouter(privateGroup)
	universityRouter.InitOrganizationRouter(privateGroup)
	universityRouter.InitStudentRouter(privateGroup)
    universityRouter.InitTeacherRouter(privateGroup)
    universityRouter.InitTeacherAccountRouter(privateGroup)
    universityRouter.InitTeacherPortalRouter(publicGroup)
	universityRouter.InitLeaveRouter(privateGroup)
	// V13：高校管理员管理学生登录账号，需要管理员 JWT
	universityRouter.InitStudentAccountRouter(privateGroup)

	// V13：学生登录与学生个人中心，使用独立学生认证
	universityRouter.InitStudentPortalRouter(publicGroup)
    universityRouter.InitStudentLeaveRouter(publicGroup)
}
