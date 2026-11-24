package university

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type uniTeacher struct {
	ID        uint64 `json:"id" gorm:"primaryKey"`
	TenantID  uint64 `json:"-"`
	TeacherNo string `json:"teacherNo"`
	Name      string `json:"name"`
	Gender    int    `json:"gender"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	CollegeID uint64 `json:"collegeId"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
}

func (uniTeacher) TableName() string { return "uni_teachers" }

type teacherInput struct {
	ID        uint64 `json:"id"`
	TeacherNo string `json:"teacherNo"`
	Name      string `json:"name"`
	Gender    int    `json:"gender"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	CollegeID uint64 `json:"collegeId"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
}

func teacherError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"code": 7, "msg": msg})
}
func teacherOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "成功", "data": data})
}
func teacherCollegeValid(tid, id uint64) (bool, error) {
	var count int64
	err := global.GVA_DB.Table("uni_colleges").Where("tenant_id=? AND id=? AND status=1", tid, id).Count(&count).Error
	return count == 1, err
}
func teacherList(c *gin.Context) {
	tid, ok := middleware.GetTenantID(c)
	if !ok {
		teacherError(c, 403, "缺少高校身份")
		return
	}
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
	q := global.GVA_DB.Model(&uniTeacher{}).Where("tenant_id=?", tid)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("teacher_no LIKE ? OR name LIKE ?", like, like)
	}
	if raw := c.Query("collegeId"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			teacherError(c, 400, "学院ID不合法")
			return
		}
		q = q.Where("college_id=?", id)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		teacherError(c, 500, "查询失败")
		return
	}
	var rows []uniTeacher
	if err := q.Order("id DESC").Limit(size).Offset((page - 1) * size).Find(&rows).Error; err != nil {
		teacherError(c, 500, "查询失败")
		return
	}
	teacherOK(c, gin.H{"list": rows, "total": total, "page": page, "pageSize": size})
}
func teacherWrite(update bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tid, ok := middleware.GetTenantID(c)
		if !ok {
			teacherError(c, 403, "缺少高校身份")
			return
		}
		var in teacherInput
		if c.ShouldBindJSON(&in) != nil {
			teacherError(c, 400, "请求参数不合法")
			return
		}
		in.TeacherNo = strings.TrimSpace(in.TeacherNo)
		in.Name = strings.TrimSpace(in.Name)
		in.Phone = strings.TrimSpace(in.Phone)
		in.Email = strings.TrimSpace(in.Email)
		in.Title = strings.TrimSpace(in.Title)
		if in.TeacherNo == "" || len(in.TeacherNo) > 64 || in.Name == "" || len(in.Name) > 191 || len(in.Phone) > 32 || len(in.Email) > 191 || len(in.Title) > 100 || in.CollegeID == 0 || in.Gender < 0 || in.Gender > 2 || (in.Status != 1 && in.Status != 2) || (update && in.ID == 0) {
			teacherError(c, 400, "教师资料不合法")
			return
		}
		valid, err := teacherCollegeValid(tid, in.CollegeID)
		if err != nil {
			teacherError(c, 500, "学院校验失败")
			return
		}
		if !valid {
			teacherError(c, 400, "学院不存在、不属于当前高校或已停用")
			return
		}
		if update {
			tx := global.GVA_DB.Model(&uniTeacher{}).Where("tenant_id=? AND id=?", tid, in.ID).Updates(map[string]interface{}{"teacher_no": in.TeacherNo, "name": in.Name, "gender": in.Gender, "phone": in.Phone, "email": in.Email, "college_id": in.CollegeID, "title": in.Title, "status": in.Status})
			if tx.Error != nil {
				teacherError(c, 409, "更新失败：工号可能重复")
				return
			}
			if tx.RowsAffected == 0 {
				teacherError(c, 404, "教师不存在")
				return
			}
			teacherOK(c, gin.H{"id": in.ID})
			return
		}
		row := uniTeacher{TenantID: tid, TeacherNo: in.TeacherNo, Name: in.Name, Gender: in.Gender, Phone: in.Phone, Email: in.Email, CollegeID: in.CollegeID, Title: in.Title, Status: in.Status}
		if err := global.GVA_DB.Create(&row).Error; err != nil {
			teacherError(c, 409, "创建失败：工号可能重复")
			return
		}
		teacherOK(c, gin.H{"id": row.ID})
	}
}
func teacherDelete(c *gin.Context) {
	tid, ok := middleware.GetTenantID(c)
	if !ok {
		teacherError(c, 403, "缺少高校身份")
		return
	}
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		teacherError(c, 400, "教师ID不合法")
		return
	}
	var n int64
	if err := global.GVA_DB.Table("uni_students").Where("tenant_id=? AND advisor_id=?", tid, id).Count(&n).Error; err != nil {
		teacherError(c, 500, "校验关联学生失败")
		return
	}
	if n > 0 {
		teacherError(c, 409, "请先解除或转移该教师负责的学生")
		return
	}
	tx := global.GVA_DB.Where("tenant_id=? AND id=?", tid, id).Delete(&uniTeacher{})
	if tx.Error != nil {
		teacherError(c, 409, "删除失败：教师可能已被业务引用")
		return
	}
	if tx.RowsAffected == 0 {
		teacherError(c, 404, "教师不存在")
		return
	}
	teacherOK(c, gin.H{"deleted": true})
}
func InitTeacherRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("uni").Use(middleware.TenantAuth())
	group.GET("/teachers", teacherList)
	group.POST("/teachers", teacherWrite(false))
	group.PUT("/teachers", teacherWrite(true))
	group.DELETE("/teachers", teacherDelete)
}
