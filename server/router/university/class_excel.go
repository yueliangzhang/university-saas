package university

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var classImportHeaders = []string{"班级编码", "班级名称", "所属专业", "入学年份", "状态"}

func classExcelImport(commit bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, ok := orgTenant(c)
		if !ok {
			orgFail(c, http.StatusForbidden, "缺少高校身份")
			return
		}
		rows, err := excelRead(c, classImportHeaders)
		if err != nil {
			orgFail(c, http.StatusBadRequest, err.Error())
			return
		}

		issues := []excelIssue{}
		preview := make([][]string, 0, len(rows))
		run := func(db *gorm.DB) error {
			majors, err := excelOptions(db, tenantID, "major", commit)
			if err != nil {
				return err
			}
			classes := make([]orgClass, 0, len(rows))
			seen := make(map[string]bool, len(rows))
			for index, raw := range rows {
				var fields [5]string
				for column, value := range raw {
					if column >= len(fields) {
						break
					}
					fields[column] = strings.TrimSpace(value)
				}
				bad := func(message string) {
					issues = append(issues, excelIssue{Row: index + 2, Number: fields[0], Message: message})
				}
				if len(raw) > len(fields) {
					bad("存在额外列")
					continue
				}
				if strings.Join(fields[:], "") == "" {
					continue
				}
				major, ok := excelMatch(majors, fields[2])
				if !ok {
					bad("所属专业不存在、已停用或名称不唯一，请使用专业编码")
					continue
				}
				year, _ := strconv.Atoi(fields[3])
				status := map[string]int{"": 1, "启用": 1, "停用": 2}[fields[4]]
				input := orgInput{Code: fields[0], Name: fields[1], MajorID: major.ID, EnrollmentYear: year, Status: status}
				if !orgInputValid("classes", input) {
					bad("班级编码、名称、入学年份（1900–2200）或状态不合法")
					continue
				}
				key := strings.ToLower(input.Code)
				if seen[key] {
					bad("文件内班级编码重复")
					continue
				}
				seen[key] = true
				var existing int64
				if err := db.Model(&orgClass{}).Where("tenant_id=? AND code=?", tenantID, input.Code).Count(&existing).Error; err != nil {
					return err
				}
				if existing > 0 {
					bad("班级编码已存在")
					continue
				}
				classes = append(classes, orgClass{
					TenantID: tenantID, MajorID: input.MajorID, Code: input.Code,
					Name: input.Name, EnrollmentYear: input.EnrollmentYear, Status: input.Status,
				})
				preview = append(preview, []string{input.Code, input.Name, excelLabel(major.Name, major.Code), fields[3]})
			}
			if commit && len(issues) == 0 && len(classes) > 0 {
				return db.CreateInBatches(&classes, 200).Error
			}
			return nil
		}

		db := global.GVA_DB.WithContext(c.Request.Context())
		if commit {
			err = db.Transaction(run)
		} else {
			err = run(db)
		}
		if err != nil {
			orgFail(c, http.StatusConflict, "处理失败，未提交导入；请检查重复编码或专业是否变化")
			return
		}
		orgOK(c, gin.H{
			"valid": len(preview), "errors": issues, "preview": preview,
			"imported": commit && len(issues) == 0 && len(preview) > 0,
		})
	}
}
