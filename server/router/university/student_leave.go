package university

import (
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
)

// All identity comes from the verified V13 student session, never from request JSON.
func studentLeaveList(c *gin.Context) {
	tid := c.GetUint64(studentPortalTenantID)
	sid := c.GetUint64(studentPortalStudentID)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	q := global.GVA_DB.Model(&uniLeave{}).Where("tenant_id=? AND student_id=?", tid, sid)
	if status := c.Query("status"); status != "" {
		if !leaveStatusValid(status) {
			studentAuthError(c, 400, "状态不合法")
			return
		}
		q = q.Where("status=?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		studentAuthError(c, 500, "查询失败")
		return
	}
	var rows []uniLeave
	if err := q.Order("id DESC").Limit(size).Offset((page - 1) * size).Find(&rows).Error; err != nil {
		studentAuthError(c, 500, "查询失败")
		return
	}
	studentAuthOK(c, gin.H{"list": rows, "total": total, "page": page, "pageSize": size})
}

func studentLeaveCreate(c *gin.Context) {
	tid := c.GetUint64(studentPortalTenantID)
	sid := c.GetUint64(studentPortalStudentID)
	var in struct {
		LeaveType string `json:"leaveType"`
		Reason    string `json:"reason"`
		StartAt   string `json:"startAt"`
		EndAt     string `json:"endAt"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		studentAuthError(c, 400, "请求格式不合法")
		return
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if in.LeaveType != "sick" && in.LeaveType != "personal" && in.LeaveType != "other" {
		studentAuthError(c, 400, "请假类型不合法")
		return
	}
	if in.Reason == "" || len(in.Reason) > 2000 {
		studentAuthError(c, 400, "请假原因不能为空且不超过2000字节")
		return
	}
	start, e1 := leaveTime(in.StartAt)
	end, e2 := leaveTime(in.EndAt)
	if e1 != nil || e2 != nil || !end.After(start) {
		studentAuthError(c, 400, "结束时间必须晚于开始时间")
		return
	}
	valid, err := leaveStudentValid(global.GVA_DB, tid, sid)
	if err != nil {
		studentAuthError(c, 500, "校验学生失败")
		return
	}
	if !valid {
		studentAuthError(c, 403, "学生已不在读或无权提交")
		return
	}
	row := uniLeave{TenantID: tid, StudentID: sid, LeaveType: in.LeaveType, Reason: in.Reason, StartAt: start, EndAt: end, Status: "pending"}
	if err := global.GVA_DB.Create(&row).Error; err != nil {
		studentAuthError(c, 500, "提交失败")
		return
	}
	studentAuthOK(c, gin.H{"id": row.ID, "status": row.Status})
}

func studentLeaveCancel(c *gin.Context) {
	tid := c.GetUint64(studentPortalTenantID)
	sid := c.GetUint64(studentPortalStudentID)
	var in struct {
		ID uint64 `json:"id"`
	}
	if c.ShouldBindJSON(&in) != nil || in.ID == 0 {
		studentAuthError(c, 400, "请假单ID不合法")
		return
	}
	r := global.GVA_DB.Model(&uniLeave{}).Where("tenant_id=? AND student_id=? AND id=? AND status='pending'", tid, sid, in.ID).Update("status", "cancelled")
	if r.Error != nil {
		studentAuthError(c, 500, "撤销失败")
		return
	}
	if r.RowsAffected != 1 {
		studentAuthError(c, 409, "记录不存在、不是本人申请或已处理")
		return
	}
	studentAuthOK(c, gin.H{"id": in.ID, "status": "cancelled"})
}

// Public GVA router group, protected by V13's independent student-session middleware.
func InitStudentLeaveRouter(publicGroup *gin.RouterGroup) {
	g := publicGroup.Group("uni/student-portal").Use(studentPortalAuth())
	g.GET("/leaves", studentLeaveList)
	g.POST("/leaves", studentLeaveCreate)
	g.POST("/leaves/cancel", studentLeaveCancel)
}
