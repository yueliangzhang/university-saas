package university

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func dashboardAnalytics(c *gin.Context) {
	tenant := c.MustGet("monitorTenant")
	days, e := strconv.Atoi(c.DefaultQuery("days", "7"))
	if e != nil || (days != 7 && days != 30 && days != 90) {
		leaveFail(c, 400, "日期范围错误")
		return
	}
	db := global.GVA_DB.WithContext(c.Request.Context())
	// Match the existing monitor's database clock and DATETIME convention.
	var clock struct {
		Now   string
		Today string
	}
	if db.Raw("SELECT DATE_FORMAT(NOW(),'%Y-%m-%d %H:%i:%s') AS `now`, DATE_FORMAT(CURDATE(),'%Y-%m-%d') AS today").Scan(&clock).Error != nil {
		leaveFail(c, 500, "读取时间失败")
		return
	}
	today, e := time.Parse("2006-01-02", clock.Today)
	if e != nil {
		leaveFail(c, 500, "时间错误")
		return
	}
	start := today.AddDate(0, 0, 1-days).Format("2006-01-02")
	end := today.AddDate(0, 0, 1).Format("2006-01-02")
	base := func() *gorm.DB {
		return db.Table("uni_students s").Joins("JOIN uni_classes cl ON cl.id=s.class_id AND cl.tenant_id=s.tenant_id").Joins("JOIN uni_majors ma ON ma.id=cl.major_id AND ma.tenant_id=s.tenant_id").Joins("JOIN uni_colleges co ON co.id=ma.college_id AND co.tenant_id=s.tenant_id").Where("s.tenant_id=?", tenant)
	}
	filters := map[string]uint64{}
	for k := range map[string]string{"collegeId": "co.id", "majorId": "ma.id", "classId": "cl.id"} {
		if v := c.Query(k); v != "" {
			id, err := strconv.ParseUint(v, 10, 64)
			if err != nil || id == 0 {
				leaveFail(c, 400, "筛选错误")
				return
			}
			filters[k] = id
		}
	}
	students := func() *gorm.DB {
		q := base()
		for k, col := range map[string]string{"collegeId": "co.id", "majorId": "ma.id", "classId": "cl.id"} {
			if id := filters[k]; id > 0 {
				q = q.Where(col+"=?", id)
			}
		}
		return q
	}
	leaves := func() *gorm.DB {
		return students().Joins("JOIN uni_leave_requests l ON l.student_id=s.id AND l.tenant_id=s.tenant_id")
	}
	period := func() *gorm.DB { return leaves().Where("l.start_at<? AND l.end_at>?", end, start) }
	var total, todayN, current, pending int64
	fail := func() { leaveFail(c, 500, "统计查询失败") }
	if students().Where("s.status=1").Count(&total).Error != nil || leaves().Where("s.status=1 AND l.status='approved' AND l.start_at<? AND l.end_at>?", end, clock.Today).Distinct("s.id").Count(&todayN).Error != nil || leaves().Where("s.status=1 AND l.status='approved' AND l.start_at<=? AND l.end_at>?", clock.Now, clock.Now).Distinct("s.id").Count(&current).Error != nil || leaves().Where("l.status='pending'").Count(&pending).Error != nil {
		fail()
		return
	}
	rate := float64(0)
	if total > 0 {
		rate = 100 * float64(todayN) / float64(total)
	}
	out := gin.H{"today": todayN, "current": current, "pending": pending, "rate": rate, "asOf": clock.Now}
	if utils.GetUserAuthorityId(c) == 2001 {
		counts := gin.H{}
		for k, t := range map[string]string{"students": "uni_students", "teachers": "uni_teachers", "colleges": "uni_colleges", "majors": "uni_majors", "classes": "uni_classes"} {
			var n int64
			if db.Table(t).Where("tenant_id=? AND status=1", tenant).Count(&n).Error != nil {
				fail()
				return
			}
			counts[k] = n
		}
		out["overview"] = counts
	}
	type point struct {
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}
	for k, col := range map[string]string{"types": "l.leave_type", "statuses": "l.status"} {
		rows := []point{}
		if period().Select(col+" AS name, COUNT(*) AS count").Group(col).Scan(&rows).Error != nil {
			fail()
			return
		}
		out[k] = rows
	}
	trend := []point{}
	for i := 0; i < days; i++ {
		d := today.AddDate(0, 0, i+1-days)
		trend = append(trend, point{Name: d.Format("2006-01-02")})
	}
	// One recursive date series query, bounded to 90 days.
	sub := period().Where("s.status=1 AND l.status='approved'").Select("s.id AS student_id,l.start_at,l.end_at")
	if db.Raw("WITH RECURSIVE dates AS (SELECT CAST(? AS DATE) AS d UNION ALL SELECT d+INTERVAL 1 DAY FROM dates WHERE d+INTERVAL 1 DAY<CAST(? AS DATE)) SELECT DATE_FORMAT(d,'%Y-%m-%d') AS name,COUNT(DISTINCT x.student_id) AS count FROM dates LEFT JOIN (?) x ON x.start_at<d+INTERVAL 1 DAY AND x.end_at>d GROUP BY d ORDER BY d", start, end, sub).Scan(&trend).Error != nil {
		fail()
		return
	}
	out["trend"] = trend
	type rank struct {
		Name    string `json:"name"`
		Total   int64  `json:"total"`
		Count   int64  `json:"count"`
		Pending int64  `json:"pending"`
	}
	for k, col := range map[string]string{"collegeRank": "co", "classRank": "cl", "advisorRank": "a"} {
		rows := []rank{}
		q := students().Where("s.status=1").Joins("LEFT JOIN uni_teachers a ON a.id=s.advisor_id AND a.tenant_id=s.tenant_id").Joins("LEFT JOIN uni_leave_requests l ON l.student_id=s.id AND l.tenant_id=s.tenant_id AND (l.status='pending' OR (l.status='approved' AND l.start_at<? AND l.end_at>?))", end, clock.Today)
		if q.Select("COALESCE("+col+".name,'未分配') AS name,COUNT(DISTINCT s.id) AS total,COUNT(DISTINCT CASE WHEN l.status='approved' THEN s.id END) AS count,COUNT(DISTINCT CASE WHEN l.status='pending' THEN l.id END) AS pending").Group(col+".id,"+col+".name").Order("count/total DESC, count DESC").Limit(20).Scan(&rows).Error != nil {
			fail()
			return
		}
		out[k] = rows
	}
	recent := []struct {
		ID        uint64    `json:"id"`
		Name      string    `json:"name"`
		ClassName string    `json:"className"`
		LeaveType string    `json:"leaveType"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"createdAt"`
	}{}
	if period().Select("l.id,s.name,cl.name AS class_name,l.leave_type,l.status,l.created_at").Order("l.created_at DESC,l.id DESC").Limit(10).Scan(&recent).Error != nil {
		fail()
		return
	}
	out["recent"] = recent
	var school struct{ Name string }
	if db.Table("uni_tenants").Select("name").Where("id=?", tenant).Scan(&school).Error != nil {
		fail()
		return
	}
	out["tenantName"] = school.Name
	leaveOK(c, out)
}
