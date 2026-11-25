package university

import (
	"errors"
	"fmt"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func excelHeaders(student bool) []string {
	if student {
		return []string{"学号", "姓名", "班级", "辅导员", "性别", "入学年份", "电话", "邮箱", "学籍状态"}
	}
	return []string{"工号", "姓名", "所属学院", "性别", "职称", "电话", "邮箱", "状态"}
}

type excelOption struct {
	ID             uint64
	Code, Name     string
	EnrollmentYear int
}

func excelOptions(db *gorm.DB, tid uint64, kind string, lock bool) ([]excelOption, error) {
	var rows []excelOption
	var q *gorm.DB
	switch kind {
	case "class":
		q = db.Table("uni_classes cl").Select("cl.id,cl.code,cl.name,cl.enrollment_year").Joins("JOIN uni_majors m ON m.tenant_id=cl.tenant_id AND m.id=cl.major_id").Joins("JOIN uni_colleges co ON co.tenant_id=m.tenant_id AND co.id=m.college_id").Where("cl.tenant_id=? AND cl.status=1 AND m.status=1 AND co.status=1", tid)
	case "advisor":
		q = db.Table("uni_teachers").Select("id,teacher_no AS code,name").Where("tenant_id=? AND status=1", tid)
	case "major":
		q = db.Table("uni_majors m").Select("m.id,m.code,m.name").
			Joins("JOIN uni_colleges co ON co.tenant_id=m.tenant_id AND co.id=m.college_id").
			Where("m.tenant_id=? AND m.status=1 AND co.status=1", tid)
	default:
		q = db.Table("uni_colleges").Select("id,code,name").Where("tenant_id=? AND status=1", tid)
	}
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.Scan(&rows).Error
	return rows, err
}
func excelLabel(name, code string) string { return name + "（" + code + "）" }
func excelMatch(rows []excelOption, s string) (excelOption, bool) {
	var found excelOption
	n := 0
	for _, r := range rows {
		if s == r.Code || s == excelLabel(r.Name, r.Code) || s == r.Name {
			found = r
			n++
		}
	}
	return found, n == 1 && s != ""
}
func excelSend(c *gin.Context, f *excelize.File, name string) {
	b, e := f.WriteToBuffer()
	if e != nil {
		studentError(c, 500, "Excel生成失败")
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+name+`.xlsx"`)
	c.Header("Cache-Control", "no-store")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", b.Bytes())
}
func excelBook(h []string) (*excelize.File, error) {
	f := excelize.NewFile()
	for i, v := range h {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if e := f.SetCellStr("Sheet1", cell, v); e != nil {
			return f, e
		}
	}
	style, e := f.NewStyle(&excelize.Style{NumFmt: 49})
	if e != nil {
		return f, e
	}
	last, _ := excelize.ColumnNumberToName(len(h))
	if e = f.SetColStyle("Sheet1", "A:"+last, style); e != nil {
		return f, e
	}
	e = f.SetColWidth("Sheet1", "A", last, 24)
	return f, e
}
func excelTemplate(student bool) gin.HandlerFunc {
	if student {
		return excelTemplateFor(excelHeaders(true), []string{"class", "advisor"})
	}
	return excelTemplateFor(excelHeaders(false), []string{"college"})
}

func excelTemplateFor(headers []string, kinds []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tid, ok := studentTenant(c)
		if !ok {
			studentError(c, 403, "缺少高校身份")
			return
		}
		f, e := excelBook(headers)
		defer f.Close()
		if e != nil {
			studentError(c, 500, "模板生成失败")
			return
		}
		_, e = f.NewSheet("选项")
		if e != nil {
			studentError(c, 500, "模板生成失败")
			return
		}
		for i, k := range kinds {
			opts, e := excelOptions(global.GVA_DB, tid, k, false)
			if e != nil {
				studentError(c, 500, "选项查询失败")
				return
			}
			col, _ := excelize.ColumnNumberToName(i + 1)
			for j, o := range opts {
				if e = f.SetCellStr("选项", fmt.Sprintf("%s%d", col, j+1), excelLabel(o.Name, o.Code)); e != nil {
					studentError(c, 500, "模板生成失败")
					return
				}
			}
			if len(opts) > 0 {
				target, _ := excelize.ColumnNumberToName(i + 3)
				dv := excelize.NewDataValidation(true)
				dv.Sqref = target + "2:" + target + "5001"
				dv.SetSqrefDropList(fmt.Sprintf("'选项'!$%s$1:$%s$%d", col, col, len(opts)))
				if e = f.AddDataValidation("Sheet1", dv); e != nil {
					studentError(c, 500, "下拉选项生成失败")
					return
				}
			}
		}
		excelSend(c, f, "import-template")
	}
}
func excelRead(c *gin.Context, headers []string) ([][]string, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20+65536)
	if e := c.Request.ParseMultipartForm(10 << 20); e != nil {
		return nil, errors.New("上传失败或文件超过10MB")
	}
	defer c.Request.MultipartForm.RemoveAll()
	file, h, e := c.Request.FormFile("file")
	if e != nil {
		return nil, errors.New("请选择文件")
	}
	defer file.Close()
	if h.Size > 10<<20 || !strings.HasSuffix(strings.ToLower(h.Filename), ".xlsx") {
		return nil, errors.New("仅支持10MB以内xlsx")
	}
	f, e := excelize.OpenReader(file, excelize.Options{UnzipSizeLimit: 128 << 20, UnzipXMLSizeLimit: 16 << 20})
	if e != nil {
		return nil, errors.New("无法读取Excel")
	}
	defer f.Close()
	it, e := f.Rows("Sheet1")
	if e != nil {
		return nil, errors.New("缺少Sheet1")
	}
	defer it.Close()
	var rows [][]string
	for it.Next() {
		r, e := it.Columns()
		if e != nil {
			return nil, e
		}
		rows = append(rows, r)
		if len(rows) > 5001 {
			return nil, errors.New("最多5000行数据")
		}
	}
	if e = it.Error(); e != nil {
		return nil, e
	}
	if len(rows) < 2 || len(rows[0]) != len(headers) {
		return nil, errors.New("模板不正确或无数据")
	}
	for i, h := range headers {
		if strings.TrimSpace(rows[0][i]) != h {
			return nil, errors.New("表头不匹配，请下载最新模板")
		}
	}
	return rows[1:], nil
}

type excelIssue struct {
	Row     int    `json:"row"`
	Number  string `json:"number"`
	Message string `json:"message"`
}

func excelValidate(db *gorm.DB, tid uint64, student, commit bool, rows [][]string) ([]uniStudent, []uniTeacher, []excelIssue, error) {
	var ss []uniStudent
	var ts []uniTeacher
	issues := []excelIssue{}
	kind := "college"
	table, no := "uni_teachers", "teacher_no"
	if student {
		kind = "class"
		table, no = "uni_students", "student_no"
	}
	opts, e := excelOptions(db, tid, kind, commit)
	if e != nil {
		return nil, nil, nil, e
	}
	var advisors []excelOption
	if student {
		advisors, e = excelOptions(db, tid, "advisor", commit)
		if e != nil {
			return nil, nil, nil, e
		}
	}
	seen := map[string]bool{}
	for i, raw := range rows {
		r := make([]string, len(excelHeaders(student)))
		for j, v := range raw {
			if j < len(r) {
				r[j] = strings.TrimSpace(v)
			}
		}
		if strings.Join(r, "") == "" {
			continue
		}
		bad := func(s string) { issues = append(issues, excelIssue{i + 2, r[0], s}) }
		if len(raw) > len(r) {
			bad("存在额外列")
			continue
		}
		o, ok := excelMatch(opts, r[2])
		if !ok {
			bad("班级或学院不存在、停用或名称不唯一")
			continue
		}
		key := strings.ToLower(r[0])
		if seen[key] {
			bad("文件内编号重复")
			continue
		}
		seen[key] = true
		exists, err := universityValueExists(db, table, no, tid, 0, r[0])
		if err != nil {
			return nil, nil, nil, err
		}
		if exists {
			if student {
				bad("学号已存在")
			} else {
				bad("工号已存在")
			}
			continue
		}
		gidx, pidx, eidx, sidx := 3, 5, 6, 7
		if student {
			gidx, pidx, eidx, sidx = 4, 6, 7, 8
		}
		gender, ok := map[string]int{"": 0, "未填写": 0, "男": 1, "女": 2}[r[gidx]]
		if !ok {
			bad("性别不合法")
			continue
		}
		status := 1
		switch r[sidx] {
		case "", "在读", "在职", "启用":
		case "停用":
			status = 2
		default:
			bad("状态不合法")
			continue
		}
		if !validUniversityEmail(r[eidx]) {
			bad("邮箱不合法")
			continue
		}
		if student {
			a, ok := excelMatch(advisors, r[3])
			if !ok {
				bad("必须指定有效且唯一的辅导员")
				continue
			}
			year := o.EnrollmentYear
			if r[5] != "" {
				year, e = strconv.Atoi(r[5])
				if e != nil {
					bad("入学年份不合法")
					continue
				}
			}
			in := studentInput{StudentNo: r[0], Name: r[1], ClassID: o.ID, Gender: gender, EnrollmentYear: year, Phone: r[pidx], Email: r[eidx], Status: status}
			if !studentValid(in) {
				bad("必填字段、长度或年份不合法")
				continue
			}
			ss = append(ss, uniStudent{TenantID: tid, StudentNo: r[0], Name: r[1], ClassID: o.ID, AdvisorID: &a.ID, Gender: gender, EnrollmentYear: year, Phone: r[pidx], Email: r[eidx], Status: status})
		} else {
			if r[0] == "" || r[1] == "" || len(r[0]) > 64 || len(r[1]) > 191 || len(r[4]) > 100 || len(r[pidx]) > 32 || len(r[eidx]) > 191 {
				bad("必填字段或长度不合法")
				continue
			}
			ts = append(ts, uniTeacher{TenantID: tid, TeacherNo: r[0], Name: r[1], CollegeID: o.ID, Gender: gender, Title: r[4], Phone: r[pidx], Email: r[eidx], Status: status})
		}
	}
	return ss, ts, issues, nil
}
func excelImport(student, commit bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tid, ok := studentTenant(c)
		if !ok {
			studentError(c, 403, "缺少高校身份")
			return
		}
		rows, e := excelRead(c, excelHeaders(student))
		if e != nil {
			studentError(c, 400, e.Error())
			return
		}
		var issues []excelIssue
		var preview [][]string
		count := 0
		stage := "validate"
		run := func(db *gorm.DB) error {
			ss, ts, errs, e := excelValidate(db, tid, student, commit, rows)
			if e != nil {
				return e
			}
			issues = errs
			count = len(ss) + len(ts)
			for _, s := range ss {
				preview = append(preview, []string{s.StudentNo, s.Name})
			}
			for _, t := range ts {
				preview = append(preview, []string{t.TeacherNo, t.Name})
			}
			if !commit || len(issues) > 0 || count == 0 {
				return nil
			}
			stage = "insert_profiles"
			if student {
				if e := db.CreateInBatches(&ss, 200).Error; e != nil {
					return e
				}
				stage = "provision_accounts"
				for _, s := range ss {
					if e := provisionAccount(db, "student", tid, s.ID); e != nil {
						return e
					}
				}
				return nil
			}
			if e := db.CreateInBatches(&ts, 200).Error; e != nil {
				return e
			}
			stage = "provision_accounts"
			for _, t := range ts {
				if e := provisionAccount(db, "teacher", tid, t.ID); e != nil {
					return e
				}
			}
			return nil
		}
		if commit {
			e = global.GVA_DB.WithContext(c.Request.Context()).Transaction(run)
		} else {
			e = run(global.GVA_DB.WithContext(c.Request.Context()))
		}
		if e != nil {
			fields := []zap.Field{zap.Uint64("tenantId", tid), zap.Bool("student", student), zap.Bool("confirm", commit), zap.String("stage", stage)}
			status := http.StatusInternalServerError
			message := "导入处理失败，请联系管理员查看服务日志"
			var mysqlError *mysql.MySQLError
			if errors.As(e, &mysqlError) {
				fields = append(fields, zap.Uint16("mysqlCode", mysqlError.Number))
				switch mysqlError.Number {
				case 1062, 1451, 1452:
					status = http.StatusConflict
					message = "重复编号或关联数据已变化，请重新校验后导入"
				case 1054, 1146:
					message = "导入所需的数据库表或字段缺失，请管理员检查数据库迁移"
				}
			} else {
				fields = append(fields, zap.String("errorType", fmt.Sprintf("%T", e)))
			}
			zap.L().Error("Excel 批量导入失败", fields...)
			studentError(c, status, message)
			return
		}
		studentOK(c, gin.H{"valid": count, "errors": issues, "preview": preview, "imported": commit && len(issues) == 0 && count > 0})
	}
}
func excelExport(student bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tid, ok := studentTenant(c)
		if !ok {
			studentError(c, 403, "缺少高校身份")
			return
		}
		headers := excelHeaders(student)
		var indices []int
		if fields := c.Query("fields"); fields != "" {
			for _, field := range strings.Split(fields, ",") {
				index, err := strconv.Atoi(field)
				if err != nil || index < 0 || index >= len(headers) || slices.Contains(indices, index) {
					studentError(c, 400, "导出字段不合法")
					return
				}
				indices = append(indices, index)
			}
		} else {
			for i := range headers {
				indices = append(indices, i)
			}
		}
		selectedHeaders := make([]string, 0, len(indices))
		for _, index := range indices {
			selectedHeaders = append(selectedHeaders, headers[index])
		}
		f, e := excelBook(selectedHeaders)
		if e != nil {
			studentError(c, 500, "导出失败")
			return
		}
		defer f.Close()

		type row struct {
			Number, Name, OrgName, OrgCode, AdvisorName, AdvisorCode, Phone, Email, Title string
			Gender, Status, EnrollmentYear                                                int
		}
		var rows []row
		var q *gorm.DB
		if student {
			q = global.GVA_DB.Table("uni_students s").Select("s.student_no AS number,s.name,s.gender,s.phone,s.email,s.status,s.enrollment_year,cl.name AS org_name,cl.code AS org_code,COALESCE(t.name,'') AS advisor_name,COALESCE(t.teacher_no,'') AS advisor_code").Joins("JOIN uni_classes cl ON cl.tenant_id=s.tenant_id AND cl.id=s.class_id").Joins("LEFT JOIN uni_teachers t ON t.tenant_id=s.tenant_id AND t.id=s.advisor_id").Where("s.tenant_id=?", tid)
		} else {
			q = global.GVA_DB.Table("uni_teachers s").Select("s.teacher_no AS number,s.name,s.gender,s.phone,s.email,s.status,s.title,co.name AS org_name,co.code AS org_code").Joins("JOIN uni_colleges co ON co.tenant_id=s.tenant_id AND co.id=s.college_id").Where("s.tenant_id=?", tid)
		}
		if k := strings.TrimSpace(c.Query("keyword")); k != "" {
			col := "s.teacher_no"
			if student {
				col = "s.student_no"
			}
			q = q.Where("("+col+" LIKE ? OR s.name LIKE ?)", "%"+k+"%", "%"+k+"%")
		}
		filters := map[string]string{"collegeId": "s.college_id"}
		if student {
			filters = map[string]string{"classId": "s.class_id", "advisorId": "s.advisor_id"}
		}
		for k, col := range filters {
			if raw := c.Query(k); raw != "" {
				id, e := strconv.ParseUint(raw, 10, 64)
				if e != nil || (id == 0 && k != "advisorId") {
					studentError(c, 400, "筛选参数不合法")
					return
				}
				if id == 0 {
					q = q.Where(col + " IS NULL")
				} else {
					q = q.Where(col+"=?", id)
				}
			}
		}
		if student {
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
		}
		if e = q.Order("s.id DESC").Limit(50001).Scan(&rows).Error; e != nil || len(rows) > 50000 {
			studentError(c, 400, "导出失败或超过50000条，请缩小筛选范围")
			return
		}
		for i, r := range rows {
			g := map[int]string{0: "未填写", 1: "男", 2: "女"}[r.Gender]
			status := "在职"
			if student {
				status = "在读"
			}
			if r.Status == 2 {
				status = "停用"
			}
			org := excelLabel(r.OrgName, r.OrgCode)
			a := ""
			if r.AdvisorCode != "" {
				a = excelLabel(r.AdvisorName, r.AdvisorCode)
			}
			v := []string{r.Number, r.Name, org, g, r.Title, r.Phone, r.Email, status}
			if student {
				v = []string{r.Number, r.Name, org, a, g, strconv.Itoa(r.EnrollmentYear), r.Phone, r.Email, status}
			}
			for j, index := range indices {
				cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
				if e = f.SetCellStr("Sheet1", cell, v[index]); e != nil {
					studentError(c, 500, "导出失败")
					return
				}
			}
		}
		excelSend(c, f, "records")
	}
}
