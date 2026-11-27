package university

import (
	"errors"
	"net/http"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func advisorWorkspaceTeacher(c *gin.Context) (uint64, error) {
	tenantID, _ := middleware.GetTenantID(c)
	var teacher struct{ ID uint64 }
	err := global.GVA_DB.WithContext(c.Request.Context()).Table("uni_memberships m").
		Select("t.id").
		Joins("JOIN uni_teachers t ON t.tenant_id=m.tenant_id AND t.id=m.advisor_teacher_id AND t.status=1").
		Where("m.tenant_id=? AND m.user_id=? AND m.role=? AND m.status=1", tenantID, utils.GetUserID(c), c.GetString(middleware.TenantRoleContextKey)).
		Take(&teacher).Error
	return teacher.ID, err
}

func advisorWorkspaceStatus(c *gin.Context) {
	_, err := advisorWorkspaceTeacher(c)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		teacherPortalFail(c, http.StatusInternalServerError, "查询辅导员身份失败")
		return
	}
	teacherPortalOK(c, gin.H{"available": err == nil})
}

func advisorWorkspaceAuth(c *gin.Context) {
	teacherID, err := advisorWorkspaceTeacher(c)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		teacherPortalFail(c, http.StatusForbidden, "未关联有效的辅导员身份")
		c.Abort()
		return
	}
	if err != nil {
		teacherPortalFail(c, http.StatusInternalServerError, "查询辅导员身份失败")
		c.Abort()
		return
	}
	tenantID, _ := middleware.GetTenantID(c)
	c.Set(teacherTenantKey, tenantID)
	c.Set(teacherIDKey, teacherID)
	c.Next()
}

func InitAdvisorWorkspaceRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("uni/advisor-workspace", middleware.AdvisorTenantAuth())
	group.GET("/status", advisorWorkspaceStatus)
	workspace := group.Group("", advisorWorkspaceAuth)
	workspace.GET("/me", teacherPortalMe)
	workspace.GET("/classes", teacherAssignedClasses)
	workspace.GET("/students", teacherAssignedStudents)
	workspace.GET("/leaves", teacherLeaveList)
	workspace.GET("/leaves/:id/proof", teacherLeaveProof)
	workspace.POST("/leaves/review", teacherLeaveReview)
}
