package university

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestRecordTablesSortAndBatchDelete(t *testing.T) {
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
	if err := db.AutoMigrate(&orgCollege{}, &orgClass{}, &uniStudent{}, &uniTeacher{}); err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{
		"PRAGMA foreign_keys = ON",
		"CREATE TABLE class_reference (class_id INTEGER REFERENCES uni_classes(id) ON DELETE RESTRICT)",
		"INSERT INTO uni_classes (id, tenant_id, name) VALUES (1, 11, 'A'), (2, 11, 'Z'), (3, 22, 'External'), (4, 11, 'B')",
		"INSERT INTO class_reference (class_id) VALUES (2)",
		"INSERT INTO uni_colleges (id, tenant_id, name) VALUES (1, 11, 'A'), (2, 11, 'Z')",
		"INSERT INTO uni_teachers (id, tenant_id, name, college_id) VALUES (1, 11, 'Z', 1), (2, 11, 'A', 2)",
		"INSERT INTO uni_students (id, tenant_id, name, class_id, advisor_id) VALUES (1, 11, 'B', 1, 1), (2, 11, 'A', 2, 2)",
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(middleware.TenantIDContextKey, uint64(11)) })
	router.GET("/students", studentList)
	router.GET("/teachers", teacherList)
	router.DELETE("/teachers", teacherDelete)
	router.DELETE("/classes", orgDelete("classes"))

	for _, test := range []struct {
		path string
		id   uint64
	}{
		{"/students?sortBy=className&sortOrder=asc&pageSize=1", 1},
		{"/students?sortBy=advisorName&sortOrder=asc&pageSize=1", 2},
		{"/students?sortBy=className&sortOrder=asc&pageSize=1&page=2", 2},
		{"/teachers?sortBy=collegeName&sortOrder=asc&pageSize=1", 1},
		{"/teachers?sortBy=collegeName&sortOrder=desc&pageSize=1", 2},
	} {
		result := httptest.NewRecorder()
		router.ServeHTTP(result, httptest.NewRequest(http.MethodGet, test.path, nil))
		var response struct {
			Data struct{ List []struct{ ID uint64 } }
		}
		err := json.Unmarshal(result.Body.Bytes(), &response)
		if err != nil || result.Code != http.StatusOK || len(response.Data.List) != 1 || response.Data.List[0].ID != test.id {
			t.Fatalf("%s: %d %s", test.path, result.Code, result.Body)
		}
	}

	for _, test := range []struct {
		path string
		code int
	}{
		{"/classes?ids=1,1", http.StatusBadRequest},
		{"/classes?ids=1,2", http.StatusConflict},
		{"/classes?ids=1,3", http.StatusNotFound},
		{"/teachers?ids=1,2", http.StatusConflict},
		{"/classes?ids=1,4", http.StatusOK},
	} {
		result := httptest.NewRecorder()
		router.ServeHTTP(result, httptest.NewRequest(http.MethodDelete, test.path, nil))
		if result.Code != test.code {
			t.Fatalf("%s: %d %s", test.path, result.Code, result.Body)
		}
	}

	var ids []uint64
	if err := db.Model(&orgClass{}).Order("id").Pluck("id", &ids).Error; err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != 2 || ids[1] != 3 {
		t.Fatalf("remaining classes: %v", ids)
	}
}
