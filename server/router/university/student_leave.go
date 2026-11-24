package university

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
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
	q := global.GVA_DB.Table("uni_leave_requests l").Where("l.tenant_id=? AND l.student_id=?", tid, sid)
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
	type Leave = uniLeave
	var rows []struct {
		Leave
		HasProof   bool `json:"hasProof"`
		ProofCount int  `json:"proofCount"`
	}
	if err := q.Select("l.*, " + leaveProofColumns).Order("l.id DESC").Limit(size).Offset((page - 1) * size).Scan(&rows).Error; err != nil {
		studentAuthError(c, 500, "查询失败")
		return
	}
	studentAuthOK(c, gin.H{"list": rows, "total": total, "page": page, "pageSize": size})
}

func studentLeaveCreate(c *gin.Context) {
	tid := c.GetUint64(studentPortalTenantID)
	sid := c.GetUint64(studentPortalStudentID)
	var in struct {
		LeaveType string `json:"leaveType" form:"leaveType"`
		Reason    string `json:"reason" form:"reason"`
		StartAt   string `json:"startAt" form:"startAt"`
		EndAt     string `json:"endAt" form:"endAt"`
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxLeaveProofCount*maxLeaveProofSize+(64<<10))
	var proofs []*uniLeaveProof
	if c.ContentType() == "multipart/form-data" {
		form, err := c.MultipartForm()
		if err != nil {
			var limit *http.MaxBytesError
			if errors.As(err, &limit) {
				studentAuthError(c, http.StatusRequestEntityTooLarge, "证明文件总大小不能超过 25 MB")
			} else {
				studentAuthError(c, http.StatusBadRequest, "无法解析证明上传表单，请刷新页面后重新选择文件")
			}
			return
		}
		defer form.RemoveAll()
		if err := binding.MapFormWithTag(&in, form.Value, "form"); err != nil {
			studentAuthError(c, http.StatusBadRequest, "请假表单字段格式不合法")
			return
		}
		files := form.File["proof"]
		if len(files) > maxLeaveProofCount {
			studentAuthError(c, 400, "最多上传 5 份证明文件")
			return
		}
		for index, file := range files {
			proof, err := readLeaveProof(file)
			if err != nil {
				studentAuthError(c, 400, err.Error())
				return
			}
			proof.ProofIndex = uint8(index + 1)
			proofs = append(proofs, proof)
		}
	} else if err := c.ShouldBindJSON(&in); err != nil {
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
	if e1 != nil || e2 != nil || start.Format(time.DateTime) != in.StartAt || end.Format(time.DateTime) != in.EndAt {
		studentAuthError(c, 400, "请选择有效的开始时间和结束时间")
		return
	}
	if !end.After(start) {
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
	// Leave times are stored as wall times and compared with the database clock in reports.
	var databaseNow string
	if err := global.GVA_DB.Raw("SELECT CAST(CURRENT_TIMESTAMP AS CHAR)").Scan(&databaseNow).Error; err != nil {
		studentAuthError(c, 500, "暂时无法校验请假时间，请重试")
		return
	}
	now, err := leaveTime(databaseNow)
	if err != nil {
		studentAuthError(c, 500, "暂时无法校验请假时间，请重试")
		return
	}
	if !start.After(now) {
		studentAuthError(c, 400, "开始时间必须晚于当前时间")
		return
	}
	row := uniLeave{TenantID: tid, StudentID: sid, LeaveType: in.LeaveType, Reason: in.Reason, StartAt: start, EndAt: end, Status: "pending"}
	if err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		for _, proof := range proofs {
			proof.LeaveID = row.ID
			if err := tx.Create(proof).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		studentAuthError(c, 500, "提交失败")
		return
	}
	studentAuthOK(c, gin.H{"id": row.ID, "status": row.Status})
}

func studentLeaveProof(c *gin.Context) {
	downloadLeaveProof(c, func(q *gorm.DB) *gorm.DB {
		return q.Where("l.tenant_id=? AND l.student_id=?", c.GetUint64(studentPortalTenantID), c.GetUint64(studentPortalStudentID))
	})
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
	g.GET("/leaves/:id/proof", studentLeaveProof)
	g.POST("/leaves", studentLeaveCreate)
	g.POST("/leaves/cancel", studentLeaveCancel)
}
