package university

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type monitorRow struct {
	ID         uint64     `json:"id"`
	StudentNo  string     `json:"studentNo"`
	Name       string     `json:"name"`
	College    string     `json:"college"`
	Major      string     `json:"major"`
	ClassName  string     `json:"className"`
	Advisor    string     `json:"advisor"`
	LeaveType  string     `json:"leaveType"`
	Status     string     `json:"status"`
	Reason     string     `json:"reason"`
	ReviewNote string     `json:"reviewNote"`
	ReviewerID *uint64    `json:"reviewerId"`
	ReviewedAt *time.Time `json:"reviewedAt"`
	StartAt    time.Time  `json:"startAt"`
	EndAt      time.Time  `json:"endAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

func InitMonitorRouter(g *gin.RouterGroup) {
	r := g.Group("uni/monitor").Use(func(c *gin.Context) {
		aid := utils.GetUserAuthorityId(c)
		role := "tenant_admin"
		if aid == 2002 {
			role = "leave_monitor"
		} else if aid != 2001 {
			c.AbortWithStatus(403)
			return
		}
		var ids []uint64
		e := global.GVA_DB.Table("uni_memberships m").Joins("JOIN uni_tenants t ON t.id=m.tenant_id AND t.status=1").Joins("JOIN sys_users u ON u.id=m.user_id AND u.deleted_at IS NULL AND u.enable=1").Where("m.user_id=? AND m.role=? AND m.status=1 AND u.authority_id=?", utils.GetUserID(c), role, aid).Pluck("m.tenant_id", &ids).Error
		if e != nil || len(ids) != 1 {
			c.AbortWithStatus(403)
			return
		}
		c.Set("monitorTenant", ids[0])
		c.Next()
	})
	r.GET("leaves", monitorList)
	r.GET("options", monitorOptions)
	r.GET("export", monitorExport)
	r.GET("dashboard", monitorStats)
	r.GET("detail", monitorDetail)
	r.GET("analytics", dashboardAnalytics)
}
func monitorQuery(c *gin.Context) *gorm.DB {
	q := global.GVA_DB.Table("uni_leave_requests l").Joins("JOIN uni_students s ON s.id=l.student_id AND s.tenant_id=l.tenant_id").Joins("JOIN uni_classes cl ON cl.id=s.class_id AND cl.tenant_id=l.tenant_id").Joins("JOIN uni_majors ma ON ma.id=cl.major_id AND ma.tenant_id=l.tenant_id").Joins("JOIN uni_colleges co ON co.id=ma.college_id AND co.tenant_id=l.tenant_id").Joins("LEFT JOIN uni_teachers a ON a.id=s.advisor_id AND a.tenant_id=l.tenant_id").Where("l.tenant_id=?", c.MustGet("monitorTenant"))
	for k, col := range map[string]string{"collegeId": "co.id", "majorId": "ma.id", "classId": "cl.id", "advisorId": "a.id", "id": "l.id"} {
		if v := c.Query(k); v != "" {
			id, e := strconv.ParseUint(v, 10, 64)
			if e != nil || id == 0 {
				return nil
			}
			q = q.Where(col+"=?", id)
		}
	}
	for k, col := range map[string]string{"status": "l.status", "leaveType": "l.leave_type"} {
		if v := c.Query(k); v != "" {
			q = q.Where(col+"=?", v)
		}
	}
	if v := c.Query("keyword"); v != "" {
		v = "%" + v + "%"
		q = q.Where("(s.name LIKE ? OR s.student_no LIKE ? OR a.name LIKE ?)", v, v, v)
	}
	for _, k := range []string{"start", "end"} {
		if v := c.Query(k); v != "" {
			d, e := time.Parse("2006-01-02", v)
			if e != nil {
				return nil
			}
			if k == "start" {
				q = q.Where("l.end_at>?", d.Format("2006-01-02"))
			} else {
				q = q.Where("l.start_at<?", d.AddDate(0, 0, 1).Format("2006-01-02"))
			}
		}
	}
	if c.Query("start") != "" && c.Query("end") != "" && c.Query("start") > c.Query("end") {
		return nil
	}
	if c.Query("current") == "1" {
		q = q.Where("l.status='approved' AND l.start_at<=NOW() AND l.end_at>NOW()")
	}
	return q
}

const monitorSelect = "l.id,l.leave_type,l.status,l.reason,l.review_note,l.reviewer_id,l.reviewed_at,l.start_at,l.end_at,l.created_at,s.student_no,s.name,co.name AS college,ma.name AS major,cl.name AS class_name,COALESCE(a.name,'') AS advisor"

func monitorList(c *gin.Context) {
	q := monitorQuery(c)
	if q == nil {
		leaveFail(c, 400, "筛选参数错误")
		return
	}
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	n, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if p < 1 || p > 1000000 || n < 1 || n > 100 {
		leaveFail(c, 400, "分页错误")
		return
	}
	var total int64
	rows := []monitorRow{}
	if q.Count(&total).Error != nil || q.Select(monitorSelect).Order("l.id DESC").Limit(n).Offset((p-1)*n).Scan(&rows).Error != nil {
		leaveFail(c, 500, "查询失败")
		return
	}
	leaveOK(c, gin.H{"list": rows, "total": total})
}
func monitorDetail(c *gin.Context) {
	if c.Query("id") == "" {
		leaveFail(c, 400, "缺少ID")
		return
	}
	q := monitorQuery(c)
	if q == nil {
		leaveFail(c, 400, "参数错误")
		return
	}
	var r monitorRow
	res := q.Select(monitorSelect).Limit(1).Scan(&r)
	if res.Error != nil || res.RowsAffected != 1 {
		leaveFail(c, 404, "记录不存在")
		return
	}
	leaveOK(c, r)
}
func monitorStats(c *gin.Context) {
	q := monitorQuery(c)
	if q == nil {
		leaveFail(c, 400, "参数错误")
		return
	}
	var r struct {
		Total   int64 `json:"total"`
		Today   int64 `json:"today"`
		Current int64 `json:"current"`
		Pending int64 `json:"pending"`
	}
	if q.Select("COUNT(*) total,COALESCE(SUM(l.status='approved' AND l.start_at<CURDATE()+INTERVAL 1 DAY AND l.end_at>CURDATE()),0) today,COALESCE(SUM(l.status='approved' AND l.start_at<=NOW() AND l.end_at>NOW()),0) current,COALESCE(SUM(l.status='pending'),0) pending").Scan(&r).Error != nil {
		leaveFail(c, 500, "统计失败")
		return
	}
	leaveOK(c, r)
}
func monitorOptions(c *gin.Context) {
	out := gin.H{}
	for k, t := range map[string]string{"colleges": "uni_colleges", "majors": "uni_majors", "classes": "uni_classes", "advisors": "uni_teachers"} {
		var rows []map[string]interface{}
		cols := "id,name"
		if k == "majors" {
			cols += ",college_id"
		}
		if k == "classes" {
			cols += ",major_id"
		}
		if global.GVA_DB.Table(t).Select(cols).Where("tenant_id=?", c.MustGet("monitorTenant")).Order("name,id").Find(&rows).Error != nil {
			leaveFail(c, 500, "查询失败")
			return
		}
		out[k] = rows
	}
	leaveOK(c, out)
}
func monitorExport(c *gin.Context) {
	q := monitorQuery(c)
	if q == nil {
		leaveFail(c, 400, "参数错误")
		return
	}
	var rows []monitorRow
	if q.Select(monitorSelect).Order("l.id DESC").Limit(50001).Scan(&rows).Error != nil || len(rows) > 50000 {
		leaveFail(c, 400, "导出失败或超过50000条，请缩小范围")
		return
	}
	x := excelize.NewFile()
	defer x.Close()
	sheet := "Sheet1"
	headers := []string{"学号", "姓名", "学院", "专业", "班级", "辅导员", "类型", "状态", "开始", "结束"}
	for i, v := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if x.SetCellStr(sheet, cell, v) != nil {
			leaveFail(c, 500, "导出失败")
			return
		}
	}
	for j, r := range rows {
		for i, v := range []string{r.StudentNo, r.Name, r.College, r.Major, r.ClassName, r.Advisor, r.LeaveType, r.Status, r.StartAt.Format("2006-01-02 15:04:05"), r.EndAt.Format("2006-01-02 15:04:05")} {
			cell, _ := excelize.CoordinatesToCellName(i+1, j+2)
			if x.SetCellStr(sheet, cell, v) != nil {
				leaveFail(c, 500, "导出失败")
				return
			}
		}
	}
	b, e := x.WriteToBuffer()
	if e != nil {
		leaveFail(c, 500, "导出失败")
		return
	}
	c.Header("Content-Disposition", "attachment; filename=leave-monitor.xlsx")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", b.Bytes())
}
