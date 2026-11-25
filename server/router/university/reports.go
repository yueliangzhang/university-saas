package university

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type reportDimension struct{ key, name string }

var reportDimensions = map[string]reportDimension{
	"college": {"co.id", "co.name"},
	"major":   {"ma.id", "ma.name"},
	"class":   {"cl.id", "cl.name"},
	"advisor": {"COALESCE(t.id,0)", "COALESCE(t.name,'未分配')"},
	"student": {"s.id", "CONCAT(s.name,' · ',s.student_no)"},
}

type reportGroup struct {
	ID              uint64  `json:"id"`
	Name            string  `json:"name"`
	Roster          int     `json:"roster"`
	Requests        int     `json:"requests"`
	Students        int     `json:"students"`
	Approved        int     `json:"approved"`
	Reviewed        int     `json:"reviewed"`
	ApprovedMinutes int64   `json:"approvedMinutes"`
	Rate            float64 `json:"rate"`
}

type reportDay struct {
	Date  string         `json:"date"`
	Count int            `json:"count"`
	Rate  float64        `json:"rate"`
	Types map[string]int `json:"types"`
}

type reportRecord struct {
	GroupID, StudentID uint64
	LeaveType, Status  string
	StartAt, EndAt     time.Time
}

type reportResult struct {
	From            string         `json:"from"`
	To              string         `json:"to"`
	SchoolName      string         `json:"schoolName"`
	Dimension       string         `json:"dimension"`
	Roster          int            `json:"roster"`
	Requests        int            `json:"requests"`
	Students        int            `json:"students"`
	AverageDaily    float64        `json:"averageDaily"`
	ApprovedMinutes int64          `json:"approvedMinutes"`
	Trend           []reportDay    `json:"trend"`
	Types           map[string]int `json:"types"`
	Groups          []reportGroup  `json:"groups"`
}

func studentReportScope(c *gin.Context, db *gorm.DB) (*gorm.DB, bool) {
	q := db.Table("uni_students s").
		Joins("JOIN uni_classes cl ON cl.tenant_id=s.tenant_id AND cl.id=s.class_id").
		Joins("JOIN uni_majors ma ON ma.tenant_id=cl.tenant_id AND ma.id=cl.major_id").
		Joins("JOIN uni_colleges co ON co.tenant_id=ma.tenant_id AND co.id=ma.college_id").
		Joins("LEFT JOIN uni_teachers t ON t.tenant_id=s.tenant_id AND t.id=s.advisor_id").
		Where("s.tenant_id=?", c.MustGet("monitorTenant"))
	for key, column := range map[string]string{
		"collegeId": "co.id", "majorId": "ma.id", "classId": "cl.id", "advisorId": "t.id",
	} {
		if raw := c.Query(key); raw != "" {
			id, err := strconv.ParseUint(raw, 10, 64)
			if err != nil || id == 0 {
				return nil, false
			}
			q = q.Where(column+"=?", id)
		}
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		if len(keyword) > 191 {
			return nil, false
		}
		like := "%" + keyword + "%"
		q = q.Where("(s.name LIKE ? OR s.student_no LIKE ?)", like, like)
	}
	return q, true
}

func universityLeaveReport(c *gin.Context) {
	from, to := c.Query("from"), c.Query("to")
	start, startErr := time.Parse(time.DateOnly, from)
	last, endErr := time.Parse(time.DateOnly, to)
	end := last.AddDate(0, 0, 1)
	days := int(end.Sub(start) / (24 * time.Hour))
	if startErr != nil || endErr != nil || start.Year() < 1000 || last.Year() >= 9999 || days < 1 || days > 366 {
		leaveFail(c, 400, "请选择不超过366天的有效日期范围")
		return
	}
	key := c.DefaultQuery("dimension", "college")
	dimension, ok := reportDimensions[key]
	kind := c.Query("leaveType")
	if !ok || (kind != "" && !slices.Contains([]string{"sick", "personal", "other"}, kind)) {
		leaveFail(c, 400, "报表筛选不合法")
		return
	}
	format := c.Query("format")
	if format != "" && format != "xlsx" {
		leaveFail(c, 400, "报表格式不合法")
		return
	}
	db := global.GVA_DB.WithContext(c.Request.Context())
	scope, ok := studentReportScope(c, db)
	if !ok {
		leaveFail(c, 400, "组织或学生筛选不合法")
		return
	}
	result := reportResult{From: from, To: to, Dimension: key}
	if err := db.Table("uni_tenants").Select("name").Where("id=?", c.MustGet("monitorTenant")).Scan(&result.SchoolName).Error; err != nil {
		leaveFail(c, 500, "读取高校名称失败")
		return
	}
	var roster []reportGroup
	err := scope.Session(&gorm.Session{}).Select(dimension.key + " AS id," + dimension.name + " AS name,COUNT(*) AS roster").
		Group(dimension.key + "," + dimension.name).Order("id").Limit(10001).Scan(&roster).Error
	if err != nil {
		leaveFail(c, 500, "查询报表分组失败")
		return
	}
	if len(roster) > 10000 {
		leaveFail(c, 400, "分组超过10000个，请缩小筛选范围")
		return
	}
	query := scope.Session(&gorm.Session{}).Joins("JOIN uni_leave_requests l ON l.tenant_id=s.tenant_id AND l.student_id=s.id").
		Where("l.start_at<? AND l.end_at>?", end.Format(time.DateTime), start.Format(time.DateTime))
	if kind != "" {
		query = query.Where("l.leave_type=?", kind)
	}
	var records []reportRecord
	err = query.Select(dimension.key + " AS group_id,s.id AS student_id,l.leave_type,l.status,l.start_at,l.end_at").
		Order("s.id,l.start_at,l.id").Limit(50001).Scan(&records).Error
	if err != nil {
		leaveFail(c, 500, "查询请假明细失败")
		return
	}
	if len(records) > 50000 {
		leaveFail(c, 400, "请假记录超过50000条，请缩小筛选范围")
		return
	}
	buildLeaveReport(&result, roster, records, start, end, days)
	if format == "xlsx" {
		writeLeaveReport(c, result)
		return
	}
	leaveOK(c, result)
}

// DATETIME values are wall-clock timestamps. Normalize their location without shifting them.
func reportWallTime(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), value.Hour(), value.Minute(), value.Second(), value.Nanosecond(), time.UTC)
}

func buildLeaveReport(result *reportResult, roster []reportGroup, records []reportRecord, start, end time.Time, days int) {
	result.Groups = make([]reportGroup, 0, len(roster))
	result.Types = map[string]int{"sick": 0, "personal": 0, "other": 0}
	result.Trend = make([]reportDay, days)
	groups := make(map[uint64]*reportGroup, len(roster))
	for i := range roster {
		group := &roster[i]
		groups[group.ID] = group
		result.Roster += group.Roster
	}
	type intervals struct {
		deltas             []int
		first, last, count int
	}
	buckets := make(map[string]*intervals)
	for _, key := range []string{"all", "sick", "personal", "other"} {
		buckets[key] = &intervals{deltas: make([]int, days+1), first: -1, last: -1}
	}
	// Records arrive in student and start-time order, so overlapping days can be merged.
	var student uint64
	approved := false
	flush := func(bucket *intervals) {
		if bucket.first >= 0 {
			bucket.deltas[bucket.first]++
			bucket.deltas[bucket.last+1]--
		}
		bucket.first, bucket.last = -1, -1
	}
	flushStudent := func() {
		for _, bucket := range buckets {
			flush(bucket)
		}
	}
	addDays := func(bucket *intervals, left, right int) {
		if bucket.first < 0 || left > bucket.last+1 {
			flush(bucket)
			bucket.first, bucket.last = left, right
		} else {
			bucket.last = max(bucket.last, right)
		}
	}
	for _, record := range records {
		group := groups[record.GroupID]
		if group == nil {
			continue
		}
		result.Requests++
		result.Types[record.LeaveType]++
		group.Requests++
		if student != record.StudentID {
			flushStudent()
			student, approved = record.StudentID, false
			result.Students++
			group.Students++
		}
		if record.Status == "approved" || record.Status == "rejected" {
			group.Reviewed++
		}
		if record.Status != "approved" {
			continue
		}
		if !approved {
			approved = true
			group.Approved++
		}
		begin, finish := reportWallTime(record.StartAt), reportWallTime(record.EndAt)
		if begin.Before(start) {
			begin = start
		}
		if finish.After(end) {
			finish = end
		}
		if !finish.After(begin) {
			continue
		}
		minutes := int64(finish.Sub(begin) / time.Minute)
		group.ApprovedMinutes += minutes
		result.ApprovedMinutes += minutes
		left := int(begin.Sub(start) / (24 * time.Hour))
		right := int((finish.Sub(start) - time.Nanosecond) / (24 * time.Hour))
		addDays(buckets["all"], left, right)
		if bucket := buckets[record.LeaveType]; bucket != nil && record.LeaveType != "all" {
			addDays(bucket, left, right)
		}
	}
	flushStudent()
	var total int
	for i := range result.Trend {
		byType := make(map[string]int, 3)
		for key, bucket := range buckets {
			bucket.count += bucket.deltas[i]
			if key != "all" {
				byType[key] = bucket.count
			}
		}
		count := buckets["all"].count
		total += count
		rate := float64(0)
		if result.Roster > 0 {
			rate = 100 * float64(count) / float64(result.Roster)
		}
		result.Trend[i] = reportDay{Date: start.AddDate(0, 0, i).Format(time.DateOnly), Count: count, Rate: rate, Types: byType}
	}
	result.AverageDaily = float64(total) / float64(days)
	for _, group := range roster {
		if group.Roster > 0 {
			group.Rate = 100 * float64(group.Approved) / float64(group.Roster)
		}
		result.Groups = append(result.Groups, group)
	}
	slices.SortStableFunc(result.Groups, func(a, b reportGroup) int {
		return b.Requests - a.Requests
	})
}

func writeLeaveReport(c *gin.Context, result reportResult) {
	file, err := excelBook([]string{"分组", "学生数", "请假人次", "涉及学生", "已批准学生", "已处理申请", "已批准时长（分钟）", "批准涉及比例（%）"})
	if err != nil {
		file.Close()
		leaveFail(c, http.StatusInternalServerError, "生成报表失败")
		return
	}
	defer file.Close()
	fail := func() { leaveFail(c, http.StatusInternalServerError, "生成报表失败") }
	set := func(sheet string, row int, values []any) bool {
		return file.SetSheetRow(sheet, "A"+strconv.Itoa(row), &values) == nil
	}
	for i, g := range result.Groups {
		if !set("Sheet1", i+2, []any{g.Name, g.Roster, g.Requests, g.Students, g.Approved, g.Reviewed, g.ApprovedMinutes, g.Rate}) {
			fail()
			return
		}
	}
	if _, err = file.NewSheet("汇总"); err != nil {
		fail()
		return
	}
	if !set("汇总", 1, []any{result.SchoolName, result.From, result.To}) ||
		!set("汇总", 2, []any{"请假人次", "涉及学生", "日均已批准请假学生", "已批准时长（分钟）", "病假", "事假", "其他"}) ||
		!set("汇总", 3, []any{result.Requests, result.Students, result.AverageDaily, result.ApprovedMinutes, result.Types["sick"], result.Types["personal"], result.Types["other"]}) ||
		!set("汇总", 5, []any{"日期", "已批准请假学生", "请假率（%）", "病假人数", "事假人数", "其他人数"}) {
		fail()
		return
	}
	for i, day := range result.Trend {
		if !set("汇总", i+6, []any{day.Date, day.Count, day.Rate, day.Types["sick"], day.Types["personal"], day.Types["other"]}) {
			fail()
			return
		}
	}
	excelSend(c, file, "university-leave-report")
}
