package university

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func TestUniversityLeaveReport(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	pool.SetMaxOpenConns(1)
	previous := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() {
		global.GVA_DB = previous
		_ = pool.Close()
	})
	if err := db.AutoMigrate(&orgCollege{}, &orgMajor{}, &orgClass{}, &uniStudent{}, &uniTeacher{}, &uniLeave{}); err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{
		"CREATE TABLE uni_tenants (id INTEGER PRIMARY KEY,name TEXT)",
		"INSERT INTO uni_tenants VALUES (1,'First School'),(2,'Second School')",
		"INSERT INTO uni_colleges (id,tenant_id,name) VALUES (1,1,'First College'),(2,2,'Other College')",
		"INSERT INTO uni_majors (id,tenant_id,college_id,name) VALUES (1,1,1,'Nursing'),(2,2,2,'Other Major')",
		"INSERT INTO uni_classes (id,tenant_id,major_id,name) VALUES (1,1,1,'Class One'),(2,2,2,'Other Class')",
		"INSERT INTO uni_students (id,tenant_id,class_id,name,student_no) VALUES (1,1,1,'First','A01'),(2,1,1,'Second','A02'),(3,2,2,'Other','B01')",
		`INSERT INTO uni_leave_requests (id,tenant_id,student_id,leave_type,status,start_at,end_at) VALUES
		(1,1,1,'sick','approved','2026-05-01 22:00:00','2026-05-02 05:00:00'),
		(2,1,1,'personal','approved','2026-05-02 03:00:00','2026-05-02 08:00:00'),
		(3,1,2,'sick','pending','2026-05-02 12:00:00','2026-05-02 14:00:00'),
		(4,2,3,'sick','approved','2026-05-02 12:00:00','2026-05-02 14:00:00'),
		(5,1,2,'sick','approved','2026-05-01 00:00:00','2026-05-02 00:00:00')`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	router := gin.New()
	router.GET("/analytics", func(c *gin.Context) {
		c.Set("monitorTenant", uint64(1))
		dashboardAnalytics(c)
	})
	get := func(query string) *httptest.ResponseRecorder {
		t.Helper()
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/analytics?mode=report&"+query, nil))
		return response
	}
	query := "from=2026-05-02&to=2026-05-03&dimension=class"
	response := get(query)
	var parsed struct{ Data reportResult }
	if err := json.Unmarshal(response.Body.Bytes(), &parsed); err != nil || response.Code != http.StatusOK {
		t.Fatalf("report: %d %s (%v)", response.Code, response.Body, err)
	}
	report := parsed.Data
	if report.Roster != 2 || report.Requests != 3 || report.Students != 2 ||
		report.ApprovedMinutes != 600 || report.AverageDaily != 0.5 || len(report.Groups) != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}
	if first := report.Trend[0]; first.Count != 1 || first.Rate != 50 || first.Types["sick"] != 1 || first.Types["personal"] != 1 {
		t.Fatalf("unexpected daily trend: %+v", first)
	}
	if group := report.Groups[0]; group.ID != 1 || group.Rate != 50 || group.Reviewed != 2 {
		t.Fatalf("unexpected class group: %+v", group)
	}
	ex := get(query + "&format=xlsx")
	if ex.Code != http.StatusOK {
		t.Fatalf("export: %d %s", ex.Code, ex.Body)
	}
	book, err := excelize.OpenReader(bytes.NewReader(ex.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer book.Close()
	rows, err := book.GetRows("汇总")
	if err != nil || len(rows) < 6 || rows[5][1] != "1" {
		t.Fatalf("unexpected export: %v %v", rows, err)
	}
	for _, invalid := range []string{"from=2026-05-02&to=2028-05-03", query + "&leaveType=invalid"} {
		if result := get(invalid); result.Code != http.StatusBadRequest {
			t.Fatalf("invalid filter %s: %d", invalid, result.Code)
		}
	}
}
