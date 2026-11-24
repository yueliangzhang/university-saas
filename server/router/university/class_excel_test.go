package university

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func TestClassExcelImport(t *testing.T) {
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
	if err := db.AutoMigrate(&orgCollege{}, &orgMajor{}, &orgClass{}); err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"INSERT INTO uni_colleges (id, tenant_id, status) VALUES (1, 11, 1), (2, 22, 1)",
		"INSERT INTO uni_majors (id, tenant_id, college_id, code, name, status) VALUES (1, 11, 1, 'CS', '计算机', 1), (2, 22, 2, 'OTHER', '外校专业', 1), (3, 11, 1, 'OFF', '停用专业', 2)",
		"CREATE UNIQUE INDEX class_tenant_code ON uni_classes (tenant_id, code)",
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(middleware.TenantIDContextKey, uint64(11)) })
	router.GET("/template", excelTemplateFor(classImportHeaders, []string{"major"}))
	router.POST("/preview", classExcelImport(false))
	router.POST("/confirm", classExcelImport(true))
	router.POST("/classes", orgWrite("classes", false))
	router.PUT("/classes", orgWrite("classes", true))

	template := httptest.NewRecorder()
	router.ServeHTTP(template, httptest.NewRequest(http.MethodGet, "/template", nil))
	if template.Code != http.StatusOK {
		t.Fatalf("template returned %d: %s", template.Code, template.Body)
	}
	book, err := excelize.OpenReader(template.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer book.Close()
	options, err := book.GetRows("选项")
	if err != nil || len(options) != 1 || len(options[0]) != 1 || options[0][0] != "计算机（CS）" {
		t.Fatalf("unexpected major options: %v %v", options, err)
	}

	type result struct {
		Valid    int          `json:"valid"`
		Errors   []excelIssue `json:"errors"`
		Imported bool         `json:"imported"`
	}
	importRows := func(endpoint string, rows ...[]string) result {
		t.Helper()
		f, err := excelBook(classImportHeaders)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		for index, row := range rows {
			cell, _ := excelize.CoordinatesToCellName(1, index+2)
			if err := f.SetSheetRow("Sheet1", cell, &row); err != nil {
				t.Fatal(err)
			}
		}
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "classes.xlsx")
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Write(part); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequest(http.MethodPost, endpoint, &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		var decoded struct {
			Data result `json:"data"`
		}
		if response.Code != http.StatusOK {
			t.Fatalf("import returned %d: %s", response.Code, response.Body)
		}
		if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
			t.Fatal(err)
		}
		return decoded.Data
	}
	checkCount := func(want int64) {
		t.Helper()
		var count int64
		if err := db.Model(&orgClass{}).Count(&count).Error; err != nil || count != want {
			t.Fatalf("class count = %d, want %d: %v", count, want, err)
		}
	}

	first := []string{"0001", "一班", "计算机（CS）", "2026"}
	second := []string{"0002", "二班", "CS", "2026", "停用"}
	if got := importRows("/preview", first, second); got.Valid != 2 || got.Imported || len(got.Errors) != 0 {
		t.Fatalf("unexpected preview: %+v", got)
	}
	checkCount(0)
	if got := importRows("/confirm", first, []string{"0003", "三班", "OTHER", "2026"}, []string{"0004", "四班", "OFF", "2026"}); got.Valid != 1 || got.Imported || len(got.Errors) != 2 || got.Errors[0].Row != 3 {
		t.Fatalf("unexpected cross-tenant or disabled major result: %+v", got)
	}
	checkCount(0)
	if got := importRows("/confirm", first, second); got.Valid != 2 || !got.Imported || len(got.Errors) != 0 {
		t.Fatalf("unexpected successful import: %+v", got)
	}
	checkCount(2)
	if got := importRows("/confirm", first); got.Imported || len(got.Errors) != 1 {
		t.Fatalf("existing code should be rejected: %+v", got)
	}
	duplicateName := []string{"0005", " 一班 ", "CS", "2026"}
	if got := importRows("/preview", duplicateName); got.Valid != 0 || len(got.Errors) != 1 || got.Errors[0].Message != "班级名称已存在" {
		t.Fatalf("existing class name should be rejected: %+v", got)
	}
	newClass := []string{"0006", "三班", "CS", "2026"}
	repeatedName := []string{"0007", "三班", "CS", "2026"}
	if got := importRows("/confirm", newClass, repeatedName); got.Imported || len(got.Errors) != 1 || got.Errors[0].Message != "文件内班级名称重复" {
		t.Fatalf("repeated name should prevent entire import: %+v", got)
	}

	for _, tc := range []struct {
		method string
		body   string
		status int
		msg    string
	}{
		{http.MethodPost, `{"code":"0003","name":"一班","majorId":1,"enrollmentYear":2026,"status":1}`, http.StatusConflict, "组织名称已存在"},
		{http.MethodPost, `{"id":1,"code":"0001","name":"三班","majorId":1,"enrollmentYear":2026,"status":1}`, http.StatusConflict, "组织编码已存在"},
		{http.MethodPut, `{"id":1,"code":"0001","name":"二班","majorId":1,"enrollmentYear":2026,"status":1}`, http.StatusConflict, "组织名称已存在"},
		{http.MethodPut, `{"id":1,"code":"0001","name":"一班","majorId":1,"enrollmentYear":2027,"status":1}`, http.StatusOK, "成功"},
	} {
		request := httptest.NewRequest(tc.method, "/classes", bytes.NewBufferString(tc.body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != tc.status || !bytes.Contains(response.Body.Bytes(), []byte(tc.msg)) {
			t.Fatalf("%s returned %d: %s", tc.method, response.Code, response.Body)
		}
	}
	checkCount(2)
}
