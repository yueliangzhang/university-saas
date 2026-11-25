package university

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// V13 student sessions are intentionally separate from GVA administrator JWTs.
// Never add student credentials to sys_users or assign the administrator role 2001.
type studentAccountRow struct {
	ID           uint64 `json:"id" gorm:"column:id"`
	TenantID     uint64 `json:"-" gorm:"column:tenant_id"`
	StudentID    uint64 `json:"studentId" gorm:"column:student_id"`
	PasswordHash string `json:"-" gorm:"column:password_hash"`
	Status       int    `json:"status" gorm:"column:status"`
}

func (studentAccountRow) TableName() string { return "uni_student_accounts" }

type studentAccountView struct {
	ID          uint64     `json:"id" gorm:"column:id"`
	StudentID   uint64     `json:"studentId" gorm:"column:student_id"`
	StudentNo   string     `json:"studentNo" gorm:"column:student_no"`
	Name        string     `json:"name" gorm:"column:name"`
	Status      int        `json:"status" gorm:"column:status"`
	LastLoginAt *time.Time `json:"lastLoginAt" gorm:"column:last_login_at"`
}

func studentAuthError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"code": 7, "msg": message})
}
func studentAuthOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "成功", "data": data})
}
func studentPasswordValid(s string) bool {
	return utf8.ValidString(s) && utf8.RuneCountInString(s) == 6 && strings.TrimSpace(s) == s && s != "123456"
}
func studentHashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func studentAccountList(c *gin.Context) {
	tid, ok := middleware.GetTenantID(c)
	if !ok {
		studentAuthError(c, 403, "无高校身份")
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
	q := global.GVA_DB.Table("uni_students s").Joins("LEFT JOIN uni_student_accounts a ON a.tenant_id=s.tenant_id AND a.student_id=s.id").Where("s.tenant_id=?", tid)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("s.name LIKE ? OR s.student_no LIKE ?", like, like)
	}
	if status := c.Query("status"); status != "" {
		switch status {
		case "0":
			q = q.Where("a.id IS NULL")
		case "1", "2":
			q = q.Where("a.status=?", status)
		default:
			studentAuthError(c, 400, "账号状态不合法")
			return
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		studentAuthError(c, 500, "查询失败")
		return
	}
	var rows []studentAccountView
	q = recordOrder(c, q, "s.id", map[string]string{
		"studentNo": "s.student_no", "name": "s.name",
		"status": "COALESCE(a.status,0)", "lastLoginAt": "a.last_login_at",
	})
	err := q.Select("COALESCE(a.id,0) id,s.id student_id,s.student_no,s.name,COALESCE(a.status,0) status,a.last_login_at").Limit(size).Offset((page - 1) * size).Scan(&rows).Error
	if err != nil {
		studentAuthError(c, 500, "查询失败")
		return
	}
	studentAuthOK(c, gin.H{"list": rows, "total": total})
}
func studentAccountCreate(c *gin.Context) {
	tid, ok := middleware.GetTenantID(c)
	if !ok {
		studentAuthError(c, 403, "无高校身份")
		return
	}
	var in struct {
		StudentID uint64 `json:"studentId"`
		Password  string `json:"password"`
	}
	if c.ShouldBindJSON(&in) != nil || in.StudentID == 0 {
		studentAuthError(c, 400, "学生ID不合法")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		studentAuthError(c, 500, "密码处理失败")
		return
	}
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if e := tx.Table("uni_students").Where("tenant_id=? AND id=? AND status=1", tid, in.StudentID).Count(&count).Error; e != nil {
			return e
		}
		if count != 1 {
			return gorm.ErrRecordNotFound
		}
		return tx.Table("uni_student_accounts").Create(map[string]interface{}{"tenant_id": tid, "student_id": in.StudentID, "password_hash": string(hash), "status": 1, "must_change_password": true}).Error
	})
	if err != nil {
		studentAuthError(c, 400, "学生不存在、已停用或账号已创建")
		return
	}
	studentAuthOK(c, gin.H{"created": true})
}
func studentAccountReset(c *gin.Context) {
	tid, ok := middleware.GetTenantID(c)
	if !ok {
		studentAuthError(c, 403, "无高校身份")
		return
	}
	var in struct {
		StudentID uint64 `json:"studentId"`
		Password  string `json:"password"`
	}
	if c.ShouldBindJSON(&in) != nil || in.StudentID == 0 {
		studentAuthError(c, 400, "参数不合法")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		studentAuthError(c, 500, "密码处理失败")
		return
	}
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		r := tx.Table("uni_student_accounts").Where("tenant_id=? AND student_id=?", tid, in.StudentID).Updates(map[string]interface{}{"password_hash": string(hash), "password_changed_at": time.Now(), "must_change_password": true, "failed_attempts": 0, "locked_until": nil})
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return tx.Exec("UPDATE uni_student_sessions SET revoked_at=NOW(3) WHERE tenant_id=? AND account_id=(SELECT id FROM uni_student_accounts WHERE tenant_id=? AND student_id=?) AND revoked_at IS NULL", tid, tid, in.StudentID).Error
	})
	if err != nil {
		studentAuthError(c, 404, "账号不存在")
		return
	}
	studentAuthOK(c, gin.H{"reset": true})
}
func studentAccountStatus(c *gin.Context) {
	tid, ok := middleware.GetTenantID(c)
	if !ok {
		studentAuthError(c, 403, "无高校身份")
		return
	}
	var in struct {
		StudentID uint64 `json:"studentId"`
		Status    int    `json:"status"`
	}
	if c.ShouldBindJSON(&in) != nil || in.StudentID == 0 || (in.Status != 1 && in.Status != 2) {
		studentAuthError(c, 400, "参数不合法")
		return
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		r := tx.Table("uni_student_accounts").Where("tenant_id=? AND student_id=?", tid, in.StudentID).Update("status", in.Status)
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		if in.Status == 2 {
			return tx.Exec("UPDATE uni_student_sessions SET revoked_at=NOW(3) WHERE tenant_id=? AND account_id=(SELECT id FROM uni_student_accounts WHERE tenant_id=? AND student_id=?) AND revoked_at IS NULL", tid, tid, in.StudentID).Error
		}
		return nil
	})
	if err != nil {
		studentAuthError(c, 404, "账号不存在")
		return
	}
	studentAuthOK(c, gin.H{"updated": true})
}
func studentPortalLogin(c *gin.Context) {
	var in struct {
		TenantID   uint64 `json:"tenantId"`
		TenantCode string `json:"tenantCode"`
		StudentNo  string `json:"studentNo"`
		Password   string `json:"password"`
	}
	if c.ShouldBindJSON(&in) != nil {
		studentAuthError(c, 400, "参数不合法")
		return
	}
	code := strings.TrimSpace(in.TenantCode)
	if (code == "" && in.TenantID == 0) || len(code) > 64 || len(in.StudentNo) == 0 || len(in.StudentNo) > 64 || len(in.Password) > 256 {
		studentAuthError(c, 400, "参数不合法")
		return
	}
	row, err := findPortalLogin("student", code, in.TenantID, in.StudentNo)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		studentAuthError(c, 500, "登录失败")
		return
	}
	if err == nil && row.locked() {
		studentAuthError(c, 429, "登录暂时锁定，请15分钟后重试")
		return
	}
	if err != nil || !row.matches(in.Password) {
		if err == nil && recordLoginFailure("student", row.TenantID, row.ID, row.PasswordHash) != nil {
			studentAuthError(c, 500, "登录失败")
			return
		}
		studentAuthError(c, 401, "高校、学号或密码错误")
		return
	}
	data, err := issuePortalLogin("student", row)
	if err != nil {
		studentAuthError(c, 500, "创建会话失败")
		return
	}
	data["studentId"] = row.ProfileID
	studentAuthOK(c, data)
}

const studentPortalStudentID = "uniPortalStudentID"
const studentPortalTenantID = "uniPortalTenantID"
const studentPortalAccountID = "uniPortalAccountID"

func studentPortalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			studentAuthError(c, 401, "请先登录学生端")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if len(token) != 43 {
			studentAuthError(c, 401, "登录凭证无效")
			c.Abort()
			return
		}
		var row struct {
			MustChangePassword bool
			AccountID          uint64 `gorm:"column:account_id"`
			StudentID          uint64 `gorm:"column:student_id"`
			TenantID           uint64 `gorm:"column:tenant_id"`
		}
		err := global.GVA_DB.Table("uni_student_sessions sess").Select("sess.account_id,a.student_id,a.tenant_id,a.must_change_password").
			Joins("JOIN uni_student_accounts a ON a.tenant_id=sess.tenant_id AND a.id=sess.account_id").
			Joins("JOIN uni_students s ON s.tenant_id=a.tenant_id AND s.id=a.student_id").
			Joins("JOIN uni_tenants t ON t.id=a.tenant_id").
			Where("sess.token_hash=? AND sess.revoked_at IS NULL AND sess.expires_at>? AND a.status=1 AND s.status=1 AND t.status=1", studentHashToken(token), time.Now()).Take(&row).Error
		if err != nil {
			studentAuthError(c, 401, "登录已失效，请重新登录")
			c.Abort()
			return
		}
		c.Set(studentPortalStudentID, row.StudentID)
		c.Set(studentPortalTenantID, row.TenantID)
		c.Set(studentPortalAccountID, row.AccountID)
		if row.MustChangePassword {
			path := c.FullPath()
			if strings.HasSuffix(path, "/me") {
				c.JSON(200, gin.H{"code": 0, "data": gin.H{"mustChangePassword": true}})
				c.Abort()
				return
			}
			if !strings.HasSuffix(path, "/change-password") && !strings.HasSuffix(path, "/logout") {
				studentAuthError(c, 403, "请先修改初始密码")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
func studentPortalMe(c *gin.Context) {
	tid := c.GetUint64(studentPortalTenantID)
	sid := c.GetUint64(studentPortalStudentID)
	var row struct {
		StudentNo      string  `json:"studentNo"`
		Name           string  `json:"name"`
		ClassID        uint64  `json:"classId"`
		AdvisorID      *uint64 `json:"advisorId"`
		AdvisorName    string  `json:"advisorName"`
		EnrollmentYear int     `json:"enrollmentYear"`
	}
	err := global.GVA_DB.Table("uni_students s").Select("s.student_no,s.name,s.class_id,s.enrollment_year,s.advisor_id,COALESCE(t.name, '') advisor_name").Joins("LEFT JOIN uni_teachers t ON t.tenant_id=s.tenant_id AND t.id=s.advisor_id").Where("s.tenant_id=? AND s.id=?", tid, sid).Take(&row).Error
	if err != nil {
		studentAuthError(c, 404, "学生档案不存在")
		return
	}
	studentAuthOK(c, row)
}
func studentPortalChangePassword(c *gin.Context) {
	aid := c.GetUint64(studentPortalAccountID)
	tid := c.GetUint64(studentPortalTenantID)
	var in struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if c.ShouldBindJSON(&in) != nil || !studentPasswordValid(in.NewPassword) {
		studentAuthError(c, 400, "新密码必须为6位且不能为123456")
		return
	}
	var account studentAccountRow
	if err := global.GVA_DB.Where("tenant_id=? AND id=?", tid, aid).Take(&account).Error; err != nil {
		studentAuthError(c, 401, "账号不存在")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(in.OldPassword)) != nil {
		studentAuthError(c, 401, "旧密码错误")
		return
	}
	if in.OldPassword == in.NewPassword {
		studentAuthError(c, 400, "新旧密码不能相同")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		studentAuthError(c, 500, "密码处理失败")
		return
	}
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		r := tx.Table("uni_student_accounts").Where("tenant_id=? AND id=? AND password_hash=?", tid, aid, account.PasswordHash).Updates(map[string]interface{}{"password_hash": string(hash), "password_changed_at": time.Now(), "must_change_password": false, "failed_attempts": 0, "locked_until": nil})
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return tx.Exec("UPDATE uni_student_sessions SET revoked_at=NOW(3) WHERE tenant_id=? AND account_id=? AND revoked_at IS NULL", tid, aid).Error
	})
	if err != nil {
		studentAuthError(c, 409, "密码已变化，请重新登录")
		return
	}
	studentAuthOK(c, gin.H{"changed": true})
}
func studentPortalLogout(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	err := global.GVA_DB.Table("uni_student_sessions").Where("token_hash=? AND tenant_id=? AND account_id=?", studentHashToken(token), c.GetUint64(studentPortalTenantID), c.GetUint64(studentPortalAccountID)).Update("revoked_at", time.Now()).Error
	if err != nil {
		studentAuthError(c, 500, "退出失败")
		return
	}
	studentAuthOK(c, gin.H{"loggedOut": true})
}

// Register under GVA's PRIVATE group (JWT + Casbin); student sessions cannot call these.
func InitStudentAccountRouter(privateGroup *gin.RouterGroup) {
	g := privateGroup.Group("uni").Use(middleware.TenantAuth())
	g.GET("/student-accounts", studentAccountList)
	g.POST("/student-accounts", studentAccountCreate)
	g.POST("/student-accounts/reset-password", studentAccountReset)
	g.PUT("/student-accounts/status", studentAccountStatus)
}

// Register under GVA's PUBLIC group, not the JWT-protected private group.
func InitStudentPortalRouter(publicGroup *gin.RouterGroup) {
	publicGroup.POST("/uni/portal/login", unifiedPortalLogin)
	g := publicGroup.Group("uni/student-portal")
	g.POST("/login", studentPortalLogin)
	protected := g.Group("").Use(studentPortalAuth())
	protected.GET("/me", studentPortalMe)
	protected.POST("/change-password", studentPortalChangePassword)
	protected.POST("/logout", studentPortalLogout)
}
