package university

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type uniStudent struct {
	AdvisorName    string  `json:"advisorName" gorm:"->;column:advisor_name"`
	ID             uint64  `json:"id" gorm:"primaryKey"`
	TenantID       uint64  `json:"-"`
	StudentNo      string  `json:"studentNo"`
	Name           string  `json:"name"`
	Gender         int     `json:"gender"`
	Phone          string  `json:"phone"`
	Email          string  `json:"email"`
	ClassID        uint64  `json:"classId"`
	AdvisorID      *uint64 `json:"advisorId"`
	EnrollmentYear int     `json:"enrollmentYear"`
	Status         int     `json:"status"`
}

func (uniStudent) TableName() string { return "uni_students" }

type studentInput struct {
	AdvisorProvided bool    `json:"-"`
	ID              uint64  `json:"id"`
	StudentNo       string  `json:"studentNo"`
	Name            string  `json:"name"`
	Gender          int     `json:"gender"`
	Phone           string  `json:"phone"`
	Email           string  `json:"email"`
	ClassID         uint64  `json:"classId"`
	AdvisorID       *uint64 `json:"advisorId"`
	EnrollmentYear  int     `json:"enrollmentYear"`
	Status          int     `json:"status"`
}

// Preserve existing assignment when older clients omit advisorId; null explicitly clears it.
func (in *studentInput) UnmarshalJSON(data []byte) error {
	type plain studentInput
	var v plain
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	*in = studentInput(v)
	_, in.AdvisorProvided = fields["advisorId"]
	return nil
}
func studentError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"code": 7, "msg": msg})
}
func studentOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": data, "msg": "成功"})
}
func studentTenant(c *gin.Context) (uint64, bool) { return middleware.GetTenantID(c) }
func studentValid(in studentInput) bool {
	return in.StudentNo != "" && len(in.StudentNo) <= 64 && in.Name != "" && len(in.Name) <= 191 &&
		len(in.Phone) <= 32 && len(in.Email) <= 191 && in.ClassID != 0 &&
		in.Gender >= 0 && in.Gender <= 2 && in.Status >= 1 && in.Status <= 2 &&
		in.EnrollmentYear >= 1900 && in.EnrollmentYear <= 2200
}
func studentClassValid(db *gorm.DB, tenantID uint64, classID uint64) (bool, error) {
	var count int64
	err := db.Table("uni_classes cl").Joins("JOIN uni_majors m ON m.tenant_id=cl.tenant_id AND m.id=cl.major_id").
		Joins("JOIN uni_colleges co ON co.tenant_id=m.tenant_id AND co.id=m.college_id").
		Where("cl.tenant_id=? AND cl.id=? AND cl.status=1 AND m.status=1 AND co.status=1", tenantID, classID).Count(&count).Error
	return count == 1, err
}
func studentList(c *gin.Context) {
	tid, ok := studentTenant(c)
	if !ok {
		studentError(c, 403, "缺少高校身份")
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
	q := global.GVA_DB.Table("uni_students s").Where("s.tenant_id=?", tid)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("s.student_no LIKE ? OR s.name LIKE ?", like, like)
	}
	if classStr := c.Query("classId"); classStr != "" {
		classID, err := strconv.ParseUint(classStr, 10, 64)
		if err != nil || classID == 0 {
			studentError(c, 400, "classId 不合法")
			return
		}
		q = q.Where("s.class_id=?", classID)
	}
	if raw := c.Query("advisorId"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			studentError(c, 400, "辅导员ID不合法")
			return
		}
		if id == 0 {
			q = q.Where("s.advisor_id IS NULL")
		} else {
			q = q.Where("s.advisor_id=?", id)
		}
	}
	if status := c.Query("status"); status != "" {
		if status != "1" && status != "2" {
			studentError(c, 400, "状态不合法")
			return
		}
		q = q.Where("s.status=?", status)
	}
	if raw := c.Query("enrollmentYear"); raw != "" {
		year, err := strconv.Atoi(raw)
		if err != nil || year < 1900 || year > 2100 {
			studentError(c, 400, "入学年份不合法")
			return
		}
		q = q.Where("s.enrollment_year=?", year)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		studentError(c, 500, "统计学生失败")
		return
	}
	var rows []uniStudent
	q = q.Joins("LEFT JOIN uni_teachers t ON t.tenant_id=s.tenant_id AND t.id=s.advisor_id")
	if err := recordOrder(c, q, "s.id", map[string]string{
		"studentNo": "s.student_no", "name": "s.name", "gender": "s.gender",
		"className":   "(SELECT cl.name FROM uni_classes cl WHERE cl.tenant_id=s.tenant_id AND cl.id=s.class_id)",
		"advisorName": "advisor_name", "enrollmentYear": "s.enrollment_year", "phone": "s.phone", "status": "s.status",
	}).Select("s.*, COALESCE(t.name, '') AS advisor_name").Limit(size).Offset((page - 1) * size).Scan(&rows).Error; err != nil {
		studentError(c, 500, "查询学生失败")
		return
	}
	studentOK(c, gin.H{"list": rows, "total": total, "page": page, "pageSize": size})
}
func studentWrite(update bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tid, ok := studentTenant(c)
		if !ok {
			studentError(c, 403, "缺少高校身份")
			return
		}
		var in studentInput
		if err := c.ShouldBindJSON(&in); err != nil {
			studentError(c, 400, "JSON 参数不合法")
			return
		}
		in.StudentNo = strings.TrimSpace(in.StudentNo)
		in.Name = strings.TrimSpace(in.Name)
		in.Phone = strings.TrimSpace(in.Phone)
		in.Email = strings.TrimSpace(in.Email)
		if !studentValid(in) || (update && in.ID == 0) {
			studentError(c, 400, "学生参数不合法")
			return
		}
		if !validUniversityEmail(in.Email) {
			studentError(c, 400, "邮箱格式不合法")
			return
		}
		if !update {
			in.ID = 0
		}
		db := global.GVA_DB
		exists, err := universityValueExists(db, "uni_students", "student_no", tid, in.ID, in.StudentNo)
		if err != nil {
			studentError(c, 500, "学号查重失败")
			return
		}
		if exists {
			studentError(c, 409, "学号已存在")
			return
		}
		valid, err := studentClassValid(db, tid, in.ClassID)
		if err != nil {
			studentError(c, 500, "班级校验失败")
			return
		}
		if !valid {
			studentError(c, 400, "班级不存在、已停用或不属于当前高校")
			return
		}
		if !update && !in.AdvisorProvided {
			var ids []uint64
			if err := db.Table("uni_class_advisors a").Where("a.tenant_id=? AND a.class_id=?", tid, in.ClassID).Distinct("a.teacher_id").Pluck("a.teacher_id", &ids).Error; err != nil {
				studentError(c, 500, "默认辅导员查询失败")
				return
			}
			if len(ids) == 1 {
				var n int64
				if err := db.Table("uni_teachers").Where("tenant_id=? AND id=? AND status=1", tid, ids[0]).Count(&n).Error; err != nil {
					studentError(c, 500, "辅导员校验失败")
					return
				}
				if n == 1 {
					in.AdvisorID = &ids[0]
				}
			}
		}
		if in.AdvisorID != nil {
			var n int64
			if err := db.Table("uni_teachers").Where("tenant_id=? AND id=? AND status=1", tid, *in.AdvisorID).Count(&n).Error; err != nil {
				studentError(c, 500, "辅导员校验失败")
				return
			}
			if n != 1 {
				studentError(c, 400, "辅导员不存在、已停用或不属于当前高校")
				return
			}
		}
		if update {
			fields := map[string]interface{}{"student_no": in.StudentNo, "name": in.Name, "gender": in.Gender, "phone": in.Phone, "email": in.Email, "class_id": in.ClassID, "enrollment_year": in.EnrollmentYear, "status": in.Status}
			if in.AdvisorProvided {
				fields["advisor_id"] = in.AdvisorID
			}
			tx := db.Model(&uniStudent{}).Where("tenant_id=? AND id=?", tid, in.ID).Updates(fields)
			if tx.Error != nil {
				studentError(c, 409, "更新失败，请检查学号是否重复或班级关联")
				return
			}
			if tx.RowsAffected == 0 {
				studentError(c, 404, "学生不存在或未发生变化")
				return
			}
			studentOK(c, gin.H{"id": in.ID})
			return
		}
		row := uniStudent{TenantID: tid, StudentNo: in.StudentNo, Name: in.Name, Gender: in.Gender, Phone: in.Phone, Email: in.Email, ClassID: in.ClassID, AdvisorID: in.AdvisorID, EnrollmentYear: in.EnrollmentYear, Status: in.Status}
		if err := db.Transaction(func(tx *gorm.DB) error {
			if e := tx.Create(&row).Error; e != nil {
				return e
			}
			return provisionAccount(tx, "student", tid, row.ID)
		}); err != nil {
			studentError(c, 409, "创建失败，请检查学号是否重复或班级关联")
			return
		}
		studentOK(c, gin.H{"id": row.ID})
	}
}
func studentDelete(c *gin.Context) {
	tid, ok := studentTenant(c)
	if !ok {
		studentError(c, 403, "缺少高校身份")
		return
	}
	ids, ok := deletionIDs(c, true)
	if !ok {
		studentError(c, 400, "学生ID不合法")
		return
	}
	err := deleteRecords(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("tenant_id=? AND id IN ?", tid, ids).Delete(&uniStudent{})
	}, len(ids))
	if err == gorm.ErrRecordNotFound {
		studentError(c, 404, "部分学生不存在，请刷新后重试")
		return
	}
	if err != nil {
		studentError(c, 409, "删除失败：请先处理所选学生关联的业务记录")
		return
	}
	studentOK(c, gin.H{"deleted": true})
}
func InitStudentRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("uni").Use(middleware.TenantAuth())

	group.GET("/students", studentList)
	group.GET("/students/template", excelTemplate(true))
	group.GET("/students/export", excelExport(true))
	group.POST("/students/import/preview", excelImport(true, false))
	group.POST("/students/import/confirm", excelImport(true, true))
	group.POST("/students", studentWrite(false))
	group.PUT("/students", studentWrite(true))
	group.DELETE("/students", studentDelete)
}
