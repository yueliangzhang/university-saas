package university

import (
	"net/http"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type orgCollege struct {
	ID       uint64 `json:"id" gorm:"primaryKey"`
	TenantID uint64 `json:"-"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Status   int    `json:"status"`
}

func (orgCollege) TableName() string { return "uni_colleges" }

type orgMajor struct {
	ID        uint64 `json:"id" gorm:"primaryKey"`
	TenantID  uint64 `json:"-"`
	CollegeID uint64 `json:"collegeId"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Status    int    `json:"status"`
}

func (orgMajor) TableName() string { return "uni_majors" }

type orgClass struct {
	ID             uint64 `json:"id" gorm:"primaryKey"`
	TenantID       uint64 `json:"-"`
	MajorID        uint64 `json:"majorId"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	EnrollmentYear int    `json:"enrollmentYear"`
	Status         int    `json:"status"`
}

func (orgClass) TableName() string { return "uni_classes" }

type orgInput struct {
	ID             uint64 `json:"id"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	Status         int    `json:"status"`
	CollegeID      uint64 `json:"collegeId"`
	MajorID        uint64 `json:"majorId"`
	EnrollmentYear int    `json:"enrollmentYear"`
}

func orgOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": data, "msg": "成功"})
}
func orgFail(c *gin.Context, status int, msg string) { c.JSON(status, gin.H{"code": 7, "msg": msg}) }
func orgTenant(c *gin.Context) (uint64, bool)        { return middleware.GetTenantID(c) }
func orgInputValid(kind string, in orgInput) bool {
	in.Code = strings.TrimSpace(in.Code)
	in.Name = strings.TrimSpace(in.Name)
	if len(in.Code) == 0 || len(in.Code) > 64 || len(in.Name) == 0 || len(in.Name) > 191 || (in.Status != 1 && in.Status != 2) {
		return false
	}
	switch kind {
	case "majors":
		return in.CollegeID != 0
	case "classes":
		return in.MajorID != 0 && in.EnrollmentYear >= 1900 && in.EnrollmentYear <= 2200
	default:
		return true
	}
}
func orgTable(kind string) string {
	switch kind {
	case "colleges":
		return "uni_colleges"
	case "majors":
		return "uni_majors"
	case "classes":
		return "uni_classes"
	}
	return ""
}
func orgParentValid(db *gorm.DB, tenantID uint64, kind string, in orgInput) (bool, error) {
	var n int64
	switch kind {
	case "majors":
		err := db.Table("uni_colleges").Where("tenant_id=? AND id=? AND status=1", tenantID, in.CollegeID).Count(&n).Error
		return n == 1, err
	case "classes":
		err := db.Table("uni_majors m").Joins("JOIN uni_colleges c ON c.tenant_id=m.tenant_id AND c.id=m.college_id").Where("m.tenant_id=? AND m.id=? AND m.status=1 AND c.status=1", tenantID, in.MajorID).Count(&n).Error
		return n == 1, err
	}
	return true, nil
}
func orgList(kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tid, ok := orgTenant(c)
		if !ok {
			orgFail(c, 403, "缺少高校身份")
			return
		}
		db := global.GVA_DB
		switch kind {
		case "colleges":
			var rows []orgCollege
			if err := db.Where("tenant_id=?", tid).Order("id DESC").Find(&rows).Error; err != nil {
				orgFail(c, 500, "查询学院失败")
				return
			}
			orgOK(c, gin.H{"list": rows})
		case "majors":
			var rows []orgMajor
			if err := db.Where("tenant_id=?", tid).Order("id DESC").Find(&rows).Error; err != nil {
				orgFail(c, 500, "查询专业失败")
				return
			}
			orgOK(c, gin.H{"list": rows})
		case "classes":
			var rows []orgClass
			if err := db.Where("tenant_id=?", tid).Order("id DESC").Find(&rows).Error; err != nil {
				orgFail(c, 500, "查询班级失败")
				return
			}
			orgOK(c, gin.H{"list": rows})
		}
	}
}
func orgWrite(kind string, update bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tid, ok := orgTenant(c)
		if !ok {
			orgFail(c, 403, "缺少高校身份")
			return
		}
		var in orgInput
		if err := c.ShouldBindJSON(&in); err != nil || !orgInputValid(kind, in) || (update && in.ID == 0) {
			orgFail(c, 400, "参数不合法")
			return
		}
		in.Code = strings.TrimSpace(in.Code)
		in.Name = strings.TrimSpace(in.Name)
		if !update {
			in.ID = 0
		}
		db := global.GVA_DB
		valid, err := orgParentValid(db, tid, kind, in)
		if err != nil {
			orgFail(c, 500, "校验上级失败")
			return
		}
		if !valid {
			orgFail(c, 400, "上级不存在、已停用或不属于当前高校")
			return
		}
		table := orgTable(kind)
		for _, field := range []struct{ key, value, label string }{
			{"code", in.Code, "组织编码"},
			{"name", in.Name, "组织名称"},
		} {
			exists, err := universityValueExists(db, table, field.key, tid, in.ID, field.value)
			if err != nil {
				orgFail(c, 500, "组织数据查重失败")
				return
			}
			if exists {
				orgFail(c, 409, field.label+"已存在")
				return
			}
		}
		if update {
			if in.Status == 2 && (kind == "colleges" || kind == "majors") {
				var count int64
				var q *gorm.DB
				if kind == "colleges" {
					q = db.Table("uni_majors").Where("tenant_id=? AND college_id=? AND status=1", tid, in.ID)
				} else {
					q = db.Table("uni_classes").Where("tenant_id=? AND major_id=? AND status=1", tid, in.ID)
				}
				if err := q.Count(&count).Error; err != nil {
					orgFail(c, 500, "校验下级数据失败")
					return
				}
				if count > 0 {
					orgFail(c, 409, "请先停用下级数据")
					return
				}
			}
			fields := map[string]interface{}{"code": in.Code, "name": in.Name, "status": in.Status}
			if kind == "majors" {
				fields["college_id"] = in.CollegeID
			}
			if kind == "classes" {
				fields["major_id"] = in.MajorID
				fields["enrollment_year"] = in.EnrollmentYear
			}
			tx := db.Table(table).Where("tenant_id=? AND id=?", tid, in.ID).Updates(fields)
			if tx.Error != nil {
				orgFail(c, 409, "更新失败，请检查组织编码或名称是否重复、上级是否有效")
				return
			}
			if tx.RowsAffected == 0 {
				orgFail(c, 404, "记录不存在或未发生变化")
				return
			}
			orgOK(c, gin.H{"id": in.ID})
			return
		}
		var result *gorm.DB
		switch kind {
		case "colleges":
			result = db.Create(&orgCollege{TenantID: tid, Code: in.Code, Name: in.Name, Status: in.Status})
		case "majors":
			result = db.Create(&orgMajor{TenantID: tid, CollegeID: in.CollegeID, Code: in.Code, Name: in.Name, Status: in.Status})
		case "classes":
			result = db.Create(&orgClass{TenantID: tid, MajorID: in.MajorID, Code: in.Code, Name: in.Name, EnrollmentYear: in.EnrollmentYear, Status: in.Status})
		}
		if result == nil || result.Error != nil {
			orgFail(c, 409, "创建失败，请检查组织编码或名称是否重复、上级是否有效")
			return
		}
		orgOK(c, gin.H{"created": true})
	}
}
func orgDelete(kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tid, ok := orgTenant(c)
		if !ok {
			orgFail(c, 403, "缺少高校身份")
			return
		}
		ids, ok := deletionIDs(c, kind == "classes")
		if !ok {
			orgFail(c, 400, "记录 ID 不合法")
			return
		}
		err := deleteRecords(func(tx *gorm.DB) *gorm.DB {
			return tx.Table(orgTable(kind)).Where("tenant_id=? AND id IN ?", tid, ids).Delete(&struct{}{})
		}, len(ids))
		if err == gorm.ErrRecordNotFound {
			orgFail(c, 404, "部分记录不存在，请刷新后重试")
			return
		}
		if err != nil {
			orgFail(c, 409, "删除失败：请先删除关联的下级数据")
			return
		}
		orgOK(c, gin.H{"deleted": true})
	}
}

func InitOrganizationRouter(privateGroup *gin.RouterGroup) {
	group := privateGroup.Group("uni/org").Use(middleware.TenantAuth())
	group.GET("classes/template", excelTemplateFor(classImportHeaders, []string{"major"}))
	group.POST("classes/import/preview", classExcelImport(false))
	group.POST("classes/import/confirm", classExcelImport(true))
	for _, kind := range []string{"colleges", "majors", "classes"} {
		k := kind
		group.GET(k, orgList(k))
		group.POST(k, orgWrite(k, false))
		group.PUT(k, orgWrite(k, true))
		group.DELETE(k, orgDelete(k))
	}
}
