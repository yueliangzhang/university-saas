package university

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestStudentLeaveOptionalProof(t *testing.T) {
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
	for _, statement := range []string{
		"CREATE TABLE uni_students (id INTEGER, tenant_id INTEGER, status INTEGER, advisor_id INTEGER)",
		"INSERT INTO uni_students VALUES (2, 1, 1, 3)",
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(studentPortalTenantID, uint64(1))
		c.Set(studentPortalStudentID, uint64(2))
		c.Set(teacherTenantKey, uint64(1))
		advisor := uint64(3)
		if c.GetHeader("X-Test-Advisor") == "unassigned" {
			advisor = 4
		}
		c.Set(teacherIDKey, advisor)
	})
	router.POST("/student/leaves", studentLeaveCreate)
	router.GET("/student/leaves", studentLeaveList)
	router.GET("/student/leaves/:id/proof", studentLeaveProof)
	router.GET("/teacher/leaves/:id/proof", teacherLeaveProof)

	send := func(method, url, contentType string, body []byte, advisor string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, url, bytes.NewReader(body))
		request.Header.Set("Content-Type", contentType)
		request.Header.Set("X-Test-Advisor", advisor)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		return response
	}
	fields := map[string]string{"leaveType": "sick", "reason": "test", "startAt": "2026-11-24 10:00:00", "endAt": "2026-11-24 11:00:00"}
	legacy, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	response := send(http.MethodPost, "/student/leaves", "application/json", legacy, "")
	if response.Code != http.StatusOK {
		t.Fatalf("submit without proof: %d %s", response.Code, response.Body)
	}

	pdf := []byte("%PDF-1.7\nexample")
	response = send(http.MethodPost, "/student/leaves", "multipart/form-data", pdf, "")
	if response.Code != http.StatusBadRequest || !bytes.Contains(response.Body.Bytes(), []byte("无法解析证明上传表单")) {
		t.Fatalf("malformed upload: %d %s", response.Code, response.Body)
	}
	for _, tc := range []struct {
		content []byte
		count   int
		status  int
	}{
		{[]byte("fake PDF"), 1, http.StatusBadRequest},
		{pdf, 6, http.StatusBadRequest},
		{pdf, 5, http.StatusOK},
	} {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		for field, value := range fields {
			if err := writer.WriteField(field, value); err != nil {
				t.Fatal(err)
			}
		}
		for i := 1; i <= tc.count; i++ {
			file, err := writer.CreateFormFile("proof", "证明.pdf")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := file.Write(tc.content); err != nil {
				t.Fatal(err)
			}
			if _, err := file.Write([]byte{byte(i)}); err != nil {
				t.Fatal(err)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		response = send(http.MethodPost, "/student/leaves", writer.FormDataContentType(), body.Bytes(), "")
		if response.Code != tc.status {
			t.Fatalf("submit with proof: %d %s", response.Code, response.Body)
		}
	}

	response = send(http.MethodGet, "/student/leaves", "", nil, "")
	var result struct {
		Data struct {
			List []struct {
				ID         uint64 `json:"id"`
				HasProof   bool   `json:"hasProof"`
				ProofCount int    `json:"proofCount"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if rows := result.Data.List; len(rows) != 2 || rows[0].ID != 2 || !rows[0].HasProof || rows[0].ProofCount != 5 || rows[1].HasProof || rows[1].ProofCount != 0 {
		t.Fatalf("unexpected leave list: %s", response.Body)
	}
	for _, tc := range []struct {
		url   string
		index byte
	}{
		{"/student/leaves/2/proof", 1},
		{"/teacher/leaves/2/proof?index=5", 5},
	} {
		response = send(http.MethodGet, tc.url, "", nil, "")
		if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), append(bytes.Clone(pdf), tc.index)) {
			t.Fatalf("download from %s: %d %s", tc.url, response.Code, response.Body)
		}
	}
	response = send(http.MethodGet, "/teacher/leaves/2/proof?index=5", "", nil, "unassigned")
	if response.Code != http.StatusNotFound {
		t.Fatalf("unassigned teacher download: %d", response.Code)
	}
}
