package university

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestAdvisorWorkspaceRequiresLinkedActiveTeacher(t *testing.T) {
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
	if err := db.AutoMigrate(&uniLeave{}, &uniLeaveProof{}); err != nil {
		t.Fatal(err)
	}
	exec := func(sql string) {
		t.Helper()
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, sql := range []string{
		"CREATE TABLE sys_users (id INTEGER PRIMARY KEY, enable INTEGER, authority_id INTEGER, deleted_at DATETIME)",
		"CREATE TABLE uni_tenants (id INTEGER PRIMARY KEY, status INTEGER)",
		"CREATE TABLE uni_memberships (tenant_id INTEGER, user_id INTEGER, status INTEGER, role TEXT, advisor_teacher_id INTEGER)",
		"CREATE TABLE uni_teachers (id INTEGER PRIMARY KEY, tenant_id INTEGER, status INTEGER)",
		"CREATE TABLE uni_classes (id INTEGER PRIMARY KEY, tenant_id INTEGER, name TEXT, code TEXT)",
		"CREATE TABLE uni_students (id INTEGER PRIMARY KEY, tenant_id INTEGER, class_id INTEGER, advisor_id INTEGER, name TEXT, student_no TEXT)",
		"INSERT INTO sys_users VALUES (7, 1, 2001, NULL), (8, 1, 2002, NULL)",
		"INSERT INTO uni_tenants VALUES (1, 1), (2, 1)",
		"INSERT INTO uni_memberships VALUES (1, 7, 1, 'tenant_admin', 10), (1, 8, 1, 'leave_monitor', 11)",
		"INSERT INTO uni_teachers VALUES (10, 1, 1), (11, 1, 1), (20, 2, 1)",
		"INSERT INTO uni_classes VALUES (1, 1, 'Assigned', 'A'), (2, 1, 'OtherAdvisor', 'B'), (3, 2, 'OtherSchool', 'C')",
		"INSERT INTO uni_students VALUES (101,1,1,10,'Assigned Student','A001'),(102,1,2,11,'Other Student','B001'),(103,2,3,20,'Other School Student','C001')",
		"INSERT INTO uni_leave_requests (id,tenant_id,student_id,status,leave_type,start_at,end_at) VALUES (1,1,101,'pending','sick','2020-05-01 12:00:00','2020-05-02 12:00:00'),(2,1,102,'pending','sick','2020-05-01 12:00:00','2020-05-02 12:00:00'),(3,1,101,'pending','sick','2020-05-03 12:00:00','2020-05-04 12:00:00')",
	} {
		exec(sql)
	}
	router := gin.New()
	withClaims := router.Group("", func(c *gin.Context) {
		claims := systemReq.BaseClaims{ID: 7, AuthorityId: 2001}
		if c.Query("monitor") == "true" {
			claims = systemReq.BaseClaims{ID: 8, AuthorityId: 2002}
		}
		c.Set("claims", &systemReq.CustomClaims{BaseClaims: claims})
	})
	InitAdvisorWorkspaceRouter(withClaims)
	get := func(path string, status int) string {
		t.Helper()
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/uni/advisor-workspace/"+path, nil))
		if rec.Code != status {
			t.Fatalf("%s: %d %s", path, rec.Code, rec.Body)
		}
		return rec.Body.String()
	}
	if body := get("classes", http.StatusOK); !strings.Contains(body, "Assigned") ||
		strings.Contains(body, "OtherAdvisor") || strings.Contains(body, "OtherSchool") {
		t.Fatalf("unexpected class scope: %s", body)
	}
	postReview := func(body string, status int) {
		t.Helper()
		request := httptest.NewRequest(http.MethodPost, "/uni/advisor-workspace/leaves/review", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != status {
			t.Fatalf("approval status: %d %s", recorder.Code, recorder.Body)
		}
	}
	postReview(`{"ids":[1,2],"action":"approved"}`, http.StatusConflict)
	var pending int64
	if err := db.Model(&uniLeave{}).Where("status='pending'").Count(&pending).Error; err != nil || pending != 3 {
		t.Fatalf("mixed scope batch was not rolled back: %d %v", pending, err)
	}
	body := get("leaves?status=rejected&includeSummary=true", http.StatusOK)
	var response struct {
		Data struct {
			Total   int
			Summary struct {
				Pending int
				Trend   []struct{ Count int }
			}
		}
	}
	if err := json.Unmarshal([]byte(body), &response); err != nil || response.Data.Total != 0 ||
		response.Data.Summary.Pending != 2 || len(response.Data.Summary.Trend) != 7 {
		t.Fatalf("dashboard summary ignored scope or filtering: %v %s", err, body)
	}
	postReview(`{"ids":[1,3],"action":"approved"}`, http.StatusOK)
	export := get("leaves?format=xlsx&status=approved&keyword=A001", http.StatusOK)
	if !strings.HasPrefix(export, "PK") {
		t.Fatal("export did not return an XLSX archive")
	}
	if body := get("classes?monitor=true", http.StatusOK); !strings.Contains(body, "OtherAdvisor") ||
		strings.Contains(body, "Assigned") || strings.Contains(body, "OtherSchool") {
		t.Fatalf("unexpected monitor class scope: %s", body)
	}
	exec("UPDATE uni_memberships SET advisor_teacher_id=NULL WHERE user_id=8")
	if body := get("status?monitor=true", http.StatusOK); !strings.Contains(body, `"available":false`) {
		t.Fatalf("unexpected monitor advisor availability: %s", body)
	}
	get("classes?monitor=true", http.StatusForbidden)
	for _, teacherID := range []string{"NULL", "20", "10"} {
		exec("UPDATE uni_memberships SET advisor_teacher_id=" + teacherID + " WHERE user_id=7")
		if teacherID == "10" {
			exec("UPDATE uni_teachers SET status=2 WHERE id=10")
		}
		if body := get("status", http.StatusOK); !strings.Contains(body, `"available":false`) {
			t.Fatalf("unexpected advisor availability: %s", body)
		}
		get("classes", http.StatusForbidden)
	}
}
