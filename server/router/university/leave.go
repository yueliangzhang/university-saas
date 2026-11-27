package university

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type uniLeave struct {
	ID         uint64     `json:"id" gorm:"primaryKey"`
	TenantID   uint64     `json:"-"`
	StudentID  uint64     `json:"studentId"`
	LeaveType  string     `json:"leaveType"`
	Reason     string     `json:"reason"`
	StartAt    time.Time  `json:"startAt"`
	EndAt      time.Time  `json:"endAt"`
	Status     string     `json:"status"`
	ReviewNote string     `json:"reviewNote"`
	ReviewerID *uint64    `json:"reviewerId"`
	ReviewedAt *time.Time `json:"reviewedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

func (uniLeave) TableName() string { return "uni_leave_requests" }

type leaveInput struct {
	ID        uint64 `json:"id"`
	StudentID uint64 `json:"studentId"`
	LeaveType string `json:"leaveType"`
	Reason    string `json:"reason"`
	StartAt   string `json:"startAt"`
	EndAt     string `json:"endAt"`
}
type leaveAction struct {
	ID     uint64 `json:"id"`
	Action string `json:"action"`
	Note   string `json:"note"`
}

func leaveOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "成功", "data": data})
}
func leaveFail(c *gin.Context, status int, msg string) { c.JSON(status, gin.H{"code": 7, "msg": msg}) }
func leaveTenant(c *gin.Context) (uint64, bool)        { return middleware.GetTenantID(c) }
func leaveTime(s string) (time.Time, error)            { return time.Parse("2006-01-02 15:04:05", s) }
func leaveStudentValid(tx *gorm.DB, tenantID, studentID uint64) (bool, error) {
	var count int64
	err := tx.Table("uni_students s").Where("s.tenant_id=? AND s.id=? AND s.status=1", tenantID, studentID).Count(&count).Error
	return count == 1, err
}
func leaveList(c *gin.Context) {
	tid, ok := leaveTenant(c)
	if !ok {
		leaveFail(c, 403, "缺少高校身份")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	q := global.GVA_DB.Model(&uniLeave{}).Where("uni_leave_requests.tenant_id=?", tid)
	if s := c.Query("status"); s != "" {
		if !leaveStatusValid(s) {
			leaveFail(c, 400, "状态不合法")
			return
		}
		q = q.Where("uni_leave_requests.status=?", s)
	}
	if s := c.Query("studentId"); s != "" {
		id, e := strconv.ParseUint(s, 10, 64)
		if e != nil || id == 0 {
			leaveFail(c, 400, "学生ID不合法")
			return
		}
		q = q.Where("uni_leave_requests.student_id=?", id)
	}
	if s := strings.TrimSpace(c.Query("keyword")); s != "" {
		like := "%" + s + "%"
		q = q.Joins("JOIN uni_students st ON st.tenant_id=uni_leave_requests.tenant_id AND st.id=uni_leave_requests.student_id").Where("st.name LIKE ? OR st.student_no LIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		leaveFail(c, 500, "统计请假记录失败")
		return
	}
	var rows []uniLeave
	if err := q.Select("uni_leave_requests.*").Order("uni_leave_requests.id DESC").Limit(size).Offset((page - 1) * size).Find(&rows).Error; err != nil {
		leaveFail(c, 500, "查询请假记录失败")
		return
	}
	leaveOK(c, gin.H{"list": rows, "total": total, "page": page, "pageSize": size})
}
func leaveStatusValid(s string) bool {
	return s == "pending" || s == "approved" || s == "rejected" || s == "cancelled"
}
func leaveWrite(update bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tid, ok := leaveTenant(c)
		if !ok {
			leaveFail(c, 403, "缺少高校身份")
			return
		}
		var in leaveInput
		if c.ShouldBindJSON(&in) != nil {
			leaveFail(c, 400, "JSON参数不合法")
			return
		}
		in.Reason = strings.TrimSpace(in.Reason)
		if in.StudentID == 0 || len(in.Reason) == 0 || len(in.Reason) > 2000 || (in.LeaveType != "sick" && in.LeaveType != "personal" && in.LeaveType != "other") || (update && in.ID == 0) {
			leaveFail(c, 400, "请假参数不合法")
			return
		}
		start, e1 := leaveTime(in.StartAt)
		end, e2 := leaveTime(in.EndAt)
		if e1 != nil || e2 != nil || !end.After(start) {
			leaveFail(c, 400, "请假时间不合法，结束时间必须晚于开始时间")
			return
		}
		db := global.GVA_DB
		valid, err := leaveStudentValid(db, tid, in.StudentID)
		if err != nil {
			leaveFail(c, 500, "学生校验失败")
			return
		}
		if !valid {
			leaveFail(c, 400, "学生不存在、不在读或不属于当前高校")
			return
		}
		if update {
			tx := db.Model(&uniLeave{}).Where("tenant_id=? AND id=? AND status='pending'", tid, in.ID).Updates(map[string]interface{}{"student_id": in.StudentID, "leave_type": in.LeaveType, "reason": in.Reason, "start_at": start, "end_at": end})
			if tx.Error != nil {
				leaveFail(c, 409, "更新失败")
				return
			}
			if tx.RowsAffected == 0 {
				leaveFail(c, 409, "记录不存在或已审批，不能修改")
				return
			}
			leaveOK(c, gin.H{"id": in.ID})
			return
		}
		row := uniLeave{TenantID: tid, StudentID: in.StudentID, LeaveType: in.LeaveType, Reason: in.Reason, StartAt: start, EndAt: end, Status: "pending"}
		if err := db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			return recordLeaveNotifications(tx, tid, []uint64{row.ID}, "submitted")
		}); err != nil {
			leaveFail(c, 409, "创建失败")
			return
		}
		leaveOK(c, gin.H{"id": row.ID})
	}
}
func leaveReview(c *gin.Context) {
	tid, ok := leaveTenant(c)
	if !ok {
		leaveFail(c, 403, "缺少高校身份")
		return
	}
	var in leaveAction
	if c.ShouldBindJSON(&in) != nil || in.ID == 0 || (in.Action != "approved" && in.Action != "rejected") {
		leaveFail(c, 400, "审批参数不合法")
		return
	}
	in.Note = strings.TrimSpace(in.Note)
	if len(in.Note) > 1000 || (in.Action == "rejected" && in.Note == "") {
		leaveFail(c, 400, "驳回必须填写原因，备注不超过1000字")
		return
	}
	reviewer := uint64(utils.GetUserID(c))
	now := time.Now()
	err := changePendingLeave(c, tid, in.ID, 0, map[string]interface{}{
		"status": in.Action, "review_note": in.Note, "reviewer_id": reviewer, "reviewed_at": now,
	}, in.Action)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		leaveFail(c, 409, "请假单不存在或已处理")
		return
	}
	if err != nil {
		leaveFail(c, 500, "审批失败")
		return
	}
	leaveOK(c, gin.H{"id": in.ID, "status": in.Action})
}
func leaveCancel(c *gin.Context) {
	tid, ok := leaveTenant(c)
	if !ok {
		leaveFail(c, 403, "缺少高校身份")
		return
	}
	var in leaveAction
	if c.ShouldBindJSON(&in) != nil || in.ID == 0 {
		leaveFail(c, 400, "参数不合法")
		return
	}
	err := changePendingLeave(c, tid, in.ID, 0, map[string]interface{}{"status": "cancelled"}, "cancelled")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		leaveFail(c, 409, "记录不存在或已处理")
		return
	}
	if err != nil {
		leaveFail(c, 500, "撤销失败")
		return
	}
	leaveOK(c, gin.H{"id": in.ID, "status": "cancelled"})
}
func changePendingLeave(c *gin.Context, tenantID, leaveID, studentID uint64, values map[string]interface{}, event string) error {
	return global.GVA_DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		q := tx.Model(&uniLeave{}).Where("tenant_id=? AND id=? AND status='pending'", tenantID, leaveID)
		if studentID != 0 {
			q = q.Where("student_id=?", studentID)
		}
		result := q.Updates(values)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return recordLeaveNotifications(tx, tenantID, []uint64{leaveID}, event)
	})
}

func InitLeaveRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("uni").Use(middleware.TenantAuth())
	group.GET("/leaves", leaveList)
	group.POST("/leaves", leaveWrite(false))
	group.PUT("/leaves", leaveWrite(true))
	group.POST("/leaves/review", leaveReview)
	group.POST("/leaves/cancel", leaveCancel)
}
