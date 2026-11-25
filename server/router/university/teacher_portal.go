package university

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Teacher sessions are separate from administrator JWTs and student sessions.
const teacherTenantKey = "uniTeacherTenantID"
const teacherIDKey = "uniTeacherID"
const teacherAccountKey = "uniTeacherAccountID"

func teacherPortalFail(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"code": 7, "msg": msg})
}
func teacherPortalOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "成功", "data": data})
}
func teacherPortalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			teacherPortalFail(c, 401, "请先登录教师端")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		if len(token) != 43 {
			teacherPortalFail(c, 401, "凭证无效")
			c.Abort()
			return
		}
		var row struct {
			MustChangePassword bool
			AccountID          uint64 `gorm:"column:account_id"`
			TeacherID          uint64 `gorm:"column:teacher_id"`
			TenantID           uint64 `gorm:"column:tenant_id"`
		}
		err := global.GVA_DB.Table("uni_teacher_sessions s").Select("s.account_id,a.teacher_id,a.tenant_id,a.must_change_password").
			Joins("JOIN uni_teacher_accounts a ON a.tenant_id=s.tenant_id AND a.id=s.account_id").
			Joins("JOIN uni_teachers t ON t.tenant_id=a.tenant_id AND t.id=a.teacher_id").
			Joins("JOIN uni_tenants u ON u.id=a.tenant_id").
			Where("s.token_hash=? AND s.revoked_at IS NULL AND s.expires_at>? AND a.status=1 AND t.status=1 AND u.status=1", studentHashToken(token), time.Now()).Take(&row).Error
		if err != nil {
			teacherPortalFail(c, 401, "教师登录已失效")
			c.Abort()
			return
		}
		c.Set(teacherTenantKey, row.TenantID)
		c.Set(teacherIDKey, row.TeacherID)
		c.Set(teacherAccountKey, row.AccountID)
		if row.MustChangePassword {
			path := c.FullPath()
			if strings.HasSuffix(path, "/me") {
				c.JSON(200, gin.H{"code": 0, "data": gin.H{"mustChangePassword": true}})
				c.Abort()
				return
			}
			if !strings.HasSuffix(path, "/change-password") && !strings.HasSuffix(path, "/logout") {
				teacherPortalFail(c, 403, "请先修改初始密码")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
func teacherPortalLogin(c *gin.Context) {
	var in struct {
		TenantID   uint64 `json:"tenantId"`
		TenantCode string `json:"tenantCode"`
		TeacherNo  string `json:"teacherNo"`
		Password   string `json:"password"`
	}
	if c.ShouldBindJSON(&in) != nil {
		teacherPortalFail(c, 400, "参数不合法")
		return
	}
	code := strings.TrimSpace(in.TenantCode)
	if (code == "" && in.TenantID == 0) || len(code) > 64 || len(in.TeacherNo) == 0 || len(in.TeacherNo) > 64 || len(in.Password) > 256 {
		teacherPortalFail(c, 400, "参数不合法")
		return
	}
	row, err := findPortalLogin("teacher", code, in.TenantID, in.TeacherNo)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		teacherPortalFail(c, 500, "登录失败")
		return
	}
	if err == nil && row.locked() {
		teacherPortalFail(c, 429, "登录暂时锁定，请15分钟后重试")
		return
	}
	if err != nil || !row.matches(in.Password) {
		if err == nil && recordLoginFailure("teacher", row.TenantID, row.ID, row.PasswordHash) != nil {
			teacherPortalFail(c, 500, "登录失败")
			return
		}
		teacherPortalFail(c, 401, "高校、工号或密码错误")
		return
	}
	data, err := issuePortalLogin("teacher", row)
	if err != nil {
		teacherPortalFail(c, 500, "创建会话失败")
		return
	}
	teacherPortalOK(c, data)
}
func teacherPortalMe(c *gin.Context) {
	var row uniTeacher
	if err := global.GVA_DB.Where("tenant_id=? AND id=?", c.GetUint64(teacherTenantKey), c.GetUint64(teacherIDKey)).Take(&row).Error; err != nil {
		teacherPortalFail(c, 404, "教师不存在")
		return
	}
	teacherPortalOK(c, row)
}
func teacherPortalChangePassword(c *gin.Context) {
	var in struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if c.ShouldBindJSON(&in) != nil || !studentPasswordValid(in.NewPassword) {
		teacherPortalFail(c, 400, "新密码必须为6位且不能为123456")
		return
	}
	tid := c.GetUint64(teacherTenantKey)
	aid := c.GetUint64(teacherAccountKey)
	var account struct {
		PasswordHash string `gorm:"column:password_hash"`
	}
	if err := global.GVA_DB.Table("uni_teacher_accounts").Select("password_hash").Where("tenant_id=? AND id=?", tid, aid).Take(&account).Error; err != nil {
		teacherPortalFail(c, 401, "账号不存在")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(in.OldPassword)) != nil {
		teacherPortalFail(c, 401, "旧密码错误")
		return
	}
	if in.OldPassword == in.NewPassword {
		teacherPortalFail(c, 400, "新旧密码不能相同")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		teacherPortalFail(c, 500, "密码处理失败")
		return
	}
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		r := tx.Table("uni_teacher_accounts").Where("tenant_id=? AND id=? AND password_hash=?", tid, aid, account.PasswordHash).Updates(map[string]interface{}{"password_hash": string(hash), "password_changed_at": time.Now(), "must_change_password": false, "failed_attempts": 0, "locked_until": nil})
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return tx.Exec("UPDATE uni_teacher_sessions SET revoked_at=NOW(3) WHERE tenant_id=? AND account_id=? AND revoked_at IS NULL", tid, aid).Error
	})
	if err != nil {
		teacherPortalFail(c, 409, "修改失败，请重试")
		return
	}
	teacherPortalOK(c, gin.H{"changed": true})
}
func teacherPortalLogout(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if err := global.GVA_DB.Table("uni_teacher_sessions").Where("tenant_id=? AND account_id=? AND token_hash=?", c.GetUint64(teacherTenantKey), c.GetUint64(teacherAccountKey), studentHashToken(token)).Update("revoked_at", time.Now()).Error; err != nil {
		teacherPortalFail(c, 500, "退出失败")
		return
	}
	teacherPortalOK(c, gin.H{"loggedOut": true})
}
func teacherAssignedClasses(c *gin.Context) {
	tid, teacher := c.GetUint64(teacherTenantKey), c.GetUint64(teacherIDKey)
	now := time.Now()
	var rows []struct {
		ID           uint64 `json:"id"`
		Name         string `json:"name"`
		Code         string `json:"code"`
		StudentCount int64  `json:"studentCount"`
		OnLeave      int64  `json:"onLeave"`
	}
	err := global.GVA_DB.Table("uni_students s").
		Select(`cl.id,cl.name,cl.code,COUNT(*) AS student_count,
		SUM(CASE WHEN EXISTS (SELECT 1 FROM uni_leave_requests l WHERE l.tenant_id=s.tenant_id AND l.student_id=s.id
		AND l.status='approved' AND l.start_at<=? AND l.end_at>?) THEN 1 ELSE 0 END) AS on_leave`, now, now).
		Joins("JOIN uni_classes cl ON cl.tenant_id=s.tenant_id AND cl.id=s.class_id").
		Where("s.tenant_id=? AND s.advisor_id=?", tid, teacher).
		Group("cl.id,cl.name,cl.code").Order("cl.id").Scan(&rows).Error
	if err != nil {
		teacherPortalFail(c, 500, "查询班级失败")
		return
	}
	teacherPortalOK(c, gin.H{"list": rows})
}
func teacherAssignedStudents(c *gin.Context) {
	tid, teacher := c.GetUint64(teacherTenantKey), c.GetUint64(teacherIDKey)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	if page > 1000000 {
		page = 1000000
	}
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	q := global.GVA_DB.Table("uni_students s").Joins("JOIN uni_classes cl ON cl.tenant_id=s.tenant_id AND cl.id=s.class_id").Where("s.tenant_id=? AND s.advisor_id=?", tid, teacher)
	if raw := c.Query("classId"); raw != "" {
		id, e := strconv.ParseUint(raw, 10, 64)
		if e != nil || id == 0 {
			teacherPortalFail(c, 400, "班级ID不合法")
			return
		}
		q = q.Where("s.class_id=?", id)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("(s.name LIKE ? OR s.student_no LIKE ?)", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		teacherPortalFail(c, 500, "查询失败")
		return
	}
	var rows []struct {
		ID             uint64 `json:"id"`
		StudentNo      string `json:"studentNo"`
		Name           string `json:"name"`
		ClassID        uint64 `json:"classId"`
		ClassName      string `json:"className"`
		EnrollmentYear int    `json:"enrollmentYear"`
		Status         int    `json:"status"`
	}
	if err := q.Select("s.id,s.student_no,s.name,s.class_id,cl.name class_name,s.enrollment_year,s.status").Order("s.id DESC").Limit(size).Offset((page - 1) * size).Scan(&rows).Error; err != nil {
		teacherPortalFail(c, 500, "查询失败")
		return
	}
	teacherPortalOK(c, gin.H{"list": rows, "total": total, "page": page, "pageSize": size})
}
func teacherLeaveList(c *gin.Context) {
	tid := c.GetUint64(teacherTenantKey)
	teacher := c.GetUint64(teacherIDKey)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	if page > 1000000 {
		page = 1000000
	}
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	scope := global.GVA_DB.WithContext(c.Request.Context()).Table("uni_leave_requests l").
		Joins("JOIN uni_students s ON s.tenant_id=l.tenant_id AND s.id=l.student_id").
		Where("s.advisor_id=?", teacher).
		Joins("JOIN uni_classes cl ON cl.tenant_id=s.tenant_id AND cl.id=s.class_id").
		Where("l.tenant_id=?", tid)
	q := scope.Session(&gorm.Session{})
	// Class filter is enforced in the database after the teacher/advisor join.
	// A teacher cannot query another class by guessing its ID.
	if raw := strings.TrimSpace(c.Query("classId")); raw != "" {
		classID, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || classID == 0 {
			teacherPortalFail(c, 400, "班级ID不合法")
			return
		}
		var assigned int64
		if err := global.GVA_DB.Table("uni_students").
			Where("tenant_id=? AND advisor_id=? AND class_id=?", tid, teacher, classID).
			Count(&assigned).Error; err != nil {
			teacherPortalFail(c, 500, "班级权限校验失败")
			return
		}
		if assigned == 0 {
			teacherPortalFail(c, 403, "无权查看该班级的请假记录")
			return
		}
		q = q.Where("s.class_id=?", classID)
	}
	if status := c.Query("status"); status != "" {
		if !leaveStatusValid(status) {
			teacherPortalFail(c, 400, "状态不合法")
			return
		}
		q = q.Where("l.status=?", status)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("(s.name LIKE ? OR s.student_no LIKE ?)", like, like)
	}
	from, to := c.Query("from"), c.Query("to")
	if from != "" || to != "" {
		start, startErr := time.Parse(time.DateOnly, from)
		end, endErr := time.Parse(time.DateOnly, to)
		if startErr != nil || endErr != nil || end.Before(start) || end.Year() >= 9999 {
			teacherPortalFail(c, 400, "日期范围不合法")
			return
		}
		q = q.Where("l.end_at>? AND l.start_at<?", start.Format(time.DateTime), end.AddDate(0, 0, 1).Format(time.DateTime))
	}
	if c.Query("format") == "xlsx" {
		exportTeacherLeaves(c, q)
		return
	}
	result := gin.H{}
	if c.Query("includeSummary") == "true" {
		summary, err := teacherLeaveSummary(scope)
		if err != nil {
			teacherPortalFail(c, 500, "统计请假失败")
			return
		}
		result["summary"] = summary
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		teacherPortalFail(c, 500, "查询失败")
		return
	}
	var rows []struct {
		ID          uint64    `json:"id"`
		StudentID   uint64    `json:"studentId"`
		StudentName string    `json:"studentName"`
		StudentNo   string    `json:"studentNo"`
		ClassID     uint64    `json:"classId"`
		ClassName   string    `json:"className"`
		LeaveType   string    `json:"leaveType"`
		Reason      string    `json:"reason"`
		StartAt     time.Time `json:"startAt"`
		EndAt       time.Time `json:"endAt"`
		Status      string    `json:"status"`
		ReviewNote  string    `json:"reviewNote"`
		HasProof    bool      `json:"hasProof"`
		ProofCount  int       `json:"proofCount"`
	}
	if err := q.Select("l.id,l.student_id,s.name student_name,s.student_no,s.class_id,cl.name class_name,l.leave_type,l.reason,l.start_at,l.end_at,l.status,l.review_note, " + leaveProofColumns).Order("l.id DESC").Limit(size).Offset((page - 1) * size).Scan(&rows).Error; err != nil {
		teacherPortalFail(c, 500, "查询失败")
		return
	}
	result["list"], result["total"] = rows, total
	teacherPortalOK(c, result)
}
func teacherLeaveSummary(scope *gorm.DB) (gin.H, error) {
	var pending int64
	if err := scope.Session(&gorm.Session{}).Where("l.status='pending'").Count(&pending).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	days := make([]time.Time, 7)
	counts := make([]int64, len(days))
	columns := make([]string, len(days))
	args := make([]any, 0, 2*len(days))
	dest := make([]any, len(days))
	for i := range days {
		days[i] = today.AddDate(0, 0, i-6)
		columns[i] = "COUNT(DISTINCT CASE WHEN l.start_at<? AND l.end_at>? THEN l.student_id END)"
		args = append(args, days[i].AddDate(0, 0, 1), days[i])
		dest[i] = &counts[i]
	}
	err := scope.Session(&gorm.Session{}).Where("l.status='approved' AND l.start_at<? AND l.end_at>?", today.AddDate(0, 0, 1), days[0]).
		Select(strings.Join(columns, ","), args...).Row().Scan(dest...)
	if err != nil {
		return nil, err
	}
	trend := make([]gin.H, len(days))
	for i, day := range days {
		trend[i] = gin.H{"date": day.Format(time.DateOnly), "count": counts[i]}
	}
	return gin.H{"pending": pending, "trend": trend}, nil
}

func exportTeacherLeaves(c *gin.Context, query *gorm.DB) {
	var rows []struct {
		StudentNo, Name, ClassName, LeaveType, Reason, Status, ReviewNote string
		StartAt, EndAt                                                    time.Time
	}
	if err := query.Select("s.student_no,s.name,cl.name AS class_name,l.leave_type,l.reason,l.status,l.review_note,l.start_at,l.end_at").
		Order("l.id DESC").Limit(10001).Scan(&rows).Error; err != nil {
		teacherPortalFail(c, 500, "查询导出数据失败")
		return
	}
	if len(rows) > 10000 {
		teacherPortalFail(c, 400, "导出最多10000条，请缩小筛选范围")
		return
	}
	file := excelize.NewFile()
	defer file.Close()
	sheet := file.GetSheetName(0)
	if err := file.SetSheetRow(sheet, "A1", &[]string{"学号", "姓名", "班级", "类型", "开始时间", "结束时间", "原因", "状态", "审批意见"}); err != nil {
		teacherPortalFail(c, 500, "生成报表失败")
		return
	}
	types := map[string]string{"sick": "病假", "personal": "事假", "other": "其他"}
	statuses := map[string]string{"pending": "待审批", "approved": "已通过", "rejected": "已驳回", "cancelled": "已撤销"}
	for i, row := range rows {
		values := []string{row.StudentNo, row.Name, row.ClassName, types[row.LeaveType], row.StartAt.Format(time.DateTime), row.EndAt.Format(time.DateTime), row.Reason, statuses[row.Status], row.ReviewNote}
		if err := file.SetSheetRow(sheet, "A"+strconv.Itoa(i+2), &values); err != nil {
			teacherPortalFail(c, 500, "生成报表失败")
			return
		}
	}
	content, err := file.WriteToBuffer()
	if err != nil {
		teacherPortalFail(c, 500, "生成报表失败")
		return
	}
	c.Header("Content-Disposition", "attachment; filename=teacher-leaves.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", content.Bytes())
}

func teacherLeaveProof(c *gin.Context) {
	downloadLeaveProof(c, func(q *gorm.DB) *gorm.DB {
		return q.Joins("JOIN uni_students s ON s.tenant_id=l.tenant_id AND s.id=l.student_id").
			Where("l.tenant_id=? AND s.advisor_id=?", c.GetUint64(teacherTenantKey), c.GetUint64(teacherIDKey))
	})
}

func teacherLeaveReview(c *gin.Context) {
	tid := c.GetUint64(teacherTenantKey)
	teacher := c.GetUint64(teacherIDKey)
	var in struct {
		ID     uint64   `json:"id"`
		IDs    []uint64 `json:"ids"`
		Action string   `json:"action"`
		Note   string   `json:"note"`
	}
	if c.ShouldBindJSON(&in) != nil || (in.Action != "approved" && in.Action != "rejected") {
		teacherPortalFail(c, 400, "参数不合法")
		return
	}
	in.Note = strings.TrimSpace(in.Note)
	if len(in.Note) > 1000 || (in.Action == "rejected" && in.Note == "") {
		teacherPortalFail(c, 400, "驳回必须填写原因，备注不超过1000字节")
		return
	}
	batch := in.IDs != nil
	ids := in.IDs
	if !batch {
		ids = []uint64{in.ID}
	} else if in.ID != 0 {
		teacherPortalFail(c, 400, "请只提交一组申请编号")
		return
	}
	seen := make(map[uint64]bool)
	if len(ids) == 0 || len(ids) > 50 {
		teacherPortalFail(c, 400, "每批可审批1至50条申请")
		return
	}
	for _, id := range ids {
		if id == 0 || seen[id] {
			teacherPortalFail(c, 400, "申请编号不合法或重复")
			return
		}
		seen[id] = true
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	err := global.GVA_DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		result := tx.Exec(`UPDATE uni_leave_requests AS l SET status=?,review_note=?,reviewed_at=?,updated_at=?,reviewer_id=NULL
		 WHERE l.id IN ? AND l.tenant_id=? AND l.status='pending' AND EXISTS (
		 SELECT 1 FROM uni_students s WHERE s.tenant_id=l.tenant_id AND s.id=l.student_id AND s.advisor_id=?)`,
			in.Action, in.Note, time.Now(), time.Now(), ids, tid, teacher)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != int64(len(ids)) {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		teacherPortalFail(c, 409, "所选申请中有记录已审批、不存在或不属于你负责的学生")
		return
	}
	if err != nil {
		teacherPortalFail(c, 500, "审批失败")
		return
	}
	teacherPortalOK(c, gin.H{"id": in.ID, "status": in.Action, "count": len(ids)})
}
func teacherAccountList(c *gin.Context) {
	tid, ok := middleware.GetTenantID(c)
	if !ok {
		teacherPortalFail(c, 403, "缺少高校身份")
		return
	}
	var rows []struct {
		TeacherID   uint64     `json:"teacherId"`
		TeacherNo   string     `json:"teacherNo"`
		Name        string     `json:"name"`
		Status      int        `json:"status"`
		LastLoginAt *time.Time `json:"lastLoginAt"`
	}
	err := global.GVA_DB.Table("uni_teacher_accounts a").Select("a.teacher_id,t.teacher_no,t.name,a.status,a.last_login_at").Joins("JOIN uni_teachers t ON t.tenant_id=a.tenant_id AND t.id=a.teacher_id").Where("a.tenant_id=?", tid).Order("a.id DESC").Scan(&rows).Error
	if err != nil {
		teacherPortalFail(c, 500, "查询失败")
		return
	}
	teacherPortalOK(c, gin.H{"list": rows})
}
func teacherAccountPassword(c *gin.Context, reset bool) {
	tid, ok := middleware.GetTenantID(c)
	if !ok {
		teacherPortalFail(c, 403, "缺少高校身份")
		return
	}
	var in struct {
		TeacherID uint64 `json:"teacherId"`
		Password  string `json:"password"`
	}
	if c.ShouldBindJSON(&in) != nil || in.TeacherID == 0 {
		teacherPortalFail(c, 400, "教师ID不合法")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		teacherPortalFail(c, 500, "密码处理失败")
		return
	}
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if e := tx.Table("uni_teachers").Where("tenant_id=? AND id=? AND status=1", tid, in.TeacherID).Count(&count).Error; e != nil {
			return e
		}
		if count != 1 {
			return gorm.ErrRecordNotFound
		}
		if reset {
			r := tx.Table("uni_teacher_accounts").Where("tenant_id=? AND teacher_id=?", tid, in.TeacherID).Updates(map[string]interface{}{"password_hash": string(hash), "password_changed_at": time.Now(), "must_change_password": true, "failed_attempts": 0, "locked_until": nil})
			if r.Error != nil {
				return r.Error
			}
			if r.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
			return tx.Exec("UPDATE uni_teacher_sessions s JOIN uni_teacher_accounts a ON a.tenant_id=s.tenant_id AND a.id=s.account_id SET s.revoked_at=NOW(3) WHERE a.tenant_id=? AND a.teacher_id=? AND s.revoked_at IS NULL", tid, in.TeacherID).Error
		}
		return tx.Table("uni_teacher_accounts").Create(map[string]interface{}{"tenant_id": tid, "teacher_id": in.TeacherID, "password_hash": string(hash), "status": 1, "must_change_password": true}).Error
	})
	if err != nil {
		teacherPortalFail(c, 409, "操作失败：教师不存在、已停用或账号状态不符合要求")
		return
	}
	teacherPortalOK(c, gin.H{"success": true})
}
func teacherAccountStatus(c *gin.Context) {
	tid, ok := middleware.GetTenantID(c)
	if !ok {
		teacherPortalFail(c, 403, "缺少高校身份")
		return
	}
	var in struct {
		TeacherID uint64 `json:"teacherId"`
		Status    int    `json:"status"`
	}
	if c.ShouldBindJSON(&in) != nil || in.TeacherID == 0 || (in.Status != 1 && in.Status != 2) {
		teacherPortalFail(c, 400, "参数不合法")
		return
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		r := tx.Table("uni_teacher_accounts").Where("tenant_id=? AND teacher_id=?", tid, in.TeacherID).Update("status", in.Status)
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		if in.Status == 2 {
			return tx.Exec("UPDATE uni_teacher_sessions s JOIN uni_teacher_accounts a ON a.tenant_id=s.tenant_id AND a.id=s.account_id SET s.revoked_at=NOW(3) WHERE a.tenant_id=? AND a.teacher_id=? AND s.revoked_at IS NULL", tid, in.TeacherID).Error
		}
		return nil
	})
	if err != nil {
		teacherPortalFail(c, 404, "账号不存在")
		return
	}
	teacherPortalOK(c, gin.H{"success": true})
}
func classAdvisorList(c *gin.Context) {
	tid, ok := middleware.GetTenantID(c)
	if !ok {
		teacherPortalFail(c, 403, "缺少高校身份")
		return
	}
	var rows []struct {
		ClassID     uint64 `json:"classId"`
		TeacherID   uint64 `json:"teacherId"`
		ClassName   string `json:"className"`
		TeacherName string `json:"teacherName"`
	}
	err := global.GVA_DB.Table("uni_class_advisors a").Select("a.class_id,a.teacher_id,c.name class_name,t.name teacher_name").Joins("JOIN uni_classes c ON c.tenant_id=a.tenant_id AND c.id=a.class_id").Joins("JOIN uni_teachers t ON t.tenant_id=a.tenant_id AND t.id=a.teacher_id").Where("a.tenant_id=?", tid).Order("a.class_id").Scan(&rows).Error
	if err != nil {
		teacherPortalFail(c, 500, "查询失败")
		return
	}
	teacherPortalOK(c, gin.H{"list": rows})
}
func classAdvisorSet(c *gin.Context) {
	tid, ok := middleware.GetTenantID(c)
	if !ok {
		teacherPortalFail(c, 403, "缺少高校身份")
		return
	}
	var in struct {
		ClassID   uint64 `json:"classId"`
		TeacherID uint64 `json:"teacherId"`
	}
	if c.ShouldBindJSON(&in) != nil || in.ClassID == 0 {
		teacherPortalFail(c, 400, "班级ID不合法")
		return
	}
	var count int64
	if err := global.GVA_DB.Table("uni_classes").Where("tenant_id=? AND id=?", tid, in.ClassID).Count(&count).Error; err != nil || count != 1 {
		teacherPortalFail(c, 400, "班级不存在")
		return
	}
	if in.TeacherID == 0 {
		if err := global.GVA_DB.Table("uni_class_advisors").Where("tenant_id=? AND class_id=?", tid, in.ClassID).Delete(&struct{ ID uint64 }{}).Error; err != nil {
			teacherPortalFail(c, 500, "取消分配失败")
			return
		}
		teacherPortalOK(c, gin.H{"success": true})
		return
	}
	if err := global.GVA_DB.Table("uni_teachers").Where("tenant_id=? AND id=? AND status=1", tid, in.TeacherID).Count(&count).Error; err != nil || count != 1 {
		teacherPortalFail(c, 400, "教师不存在或已停用")
		return
	}
	// One class has at most one advisor. The assignment can be changed by an administrator.
	err := global.GVA_DB.Exec("INSERT INTO uni_class_advisors(tenant_id,class_id,teacher_id) VALUES(?,?,?) ON DUPLICATE KEY UPDATE teacher_id=VALUES(teacher_id)", tid, in.ClassID, in.TeacherID).Error
	if err != nil {
		teacherPortalFail(c, 500, "分配失败")
		return
	}
	teacherPortalOK(c, gin.H{"success": true})
}
func InitTeacherAccountRouter(privateGroup *gin.RouterGroup) {
	g := privateGroup.Group("uni").Use(middleware.TenantAuth())
	g.GET("/teacher-accounts", teacherAccountList)
	g.POST("/teacher-accounts", func(c *gin.Context) { teacherAccountPassword(c, false) })
	g.POST("/teacher-accounts/reset-password", func(c *gin.Context) { teacherAccountPassword(c, true) })
	g.PUT("/teacher-accounts/status", teacherAccountStatus)
	g.GET("/class-advisors", classAdvisorList)
	g.PUT("/class-advisors", classAdvisorSet)
}
func InitTeacherPortalRouter(publicGroup *gin.RouterGroup) {
	g := publicGroup.Group("uni/teacher-portal")
	g.POST("/login", teacherPortalLogin)
	protected := g.Group("").Use(teacherPortalAuth())
	protected.GET("/me", teacherPortalMe)
	protected.POST("/change-password", teacherPortalChangePassword)
	protected.POST("/logout", teacherPortalLogout)
	protected.GET("/classes", teacherAssignedClasses)
	protected.GET("/students", teacherAssignedStudents)
	protected.GET("/leaves", teacherLeaveList)
	protected.GET("/leaves/:id/proof", teacherLeaveProof)
	protected.POST("/leaves/review", teacherLeaveReview)
}
