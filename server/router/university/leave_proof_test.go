package university

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
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
	if err := db.AutoMigrate(&uniLeave{}, &uniLeaveProof{}, &uniLeaveNotification{}); err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"CREATE TABLE uni_students (id INTEGER, tenant_id INTEGER, status INTEGER, advisor_id INTEGER, name TEXT, student_no TEXT)",
		"INSERT INTO uni_students VALUES (2, 1, 1, 3, 'Student One', 'S001'), (2, 2, 1, 3, 'Other tenant', 'S001')",
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.TenantIDContextKey, uint64(1))
		c.Set(studentPortalTenantID, uint64(1))
		c.Set(studentPortalStudentID, uint64(2))
		c.Set(teacherTenantKey, uint64(1))
		advisor := uint64(3)
		if c.GetHeader("X-Test-Advisor") == "unassigned" {
			advisor = 4
		}
		c.Set(teacherIDKey, advisor)
	})
	router.GET("/admin/leaves", leaveList)
	router.POST("/student/leaves", studentLeaveCreate)
	router.GET("/student/leaves", studentLeaveList)
	router.GET("/student/leaves/:id/proof", studentLeaveProof)
	router.GET("/teacher/leaves/:id/proof", teacherLeaveProof)
	router.POST("/teacher/leaves/review", teacherLeaveReview)
	router.POST("/student/leaves/cancel", studentLeaveCancel)
	registerLeaveNotifications(router.Group("/student"), notificationStudent)
	registerLeaveNotifications(router.Group("/teacher"), notificationAdvisor)

	send := func(method, url, contentType string, body []byte, advisor string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, url, bytes.NewReader(body))
		request.Header.Set("Content-Type", contentType)
		request.Header.Set("X-Test-Advisor", advisor)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		return response
	}
	const layout = "2006-01-02 15:04:05"
	now := time.Now().UTC()
	fields := map[string]string{"leaveType": "sick", "reason": "test", "startAt": now.Add(time.Hour).Format(layout), "endAt": now.Add(2 * time.Hour).Format(layout)}
	for _, tc := range []struct {
		start string
		end   string
		msg   string
	}{
		{"invalid", fields["endAt"], "请选择有效的开始时间和结束时间"},
		{fields["startAt"] + ".1", fields["endAt"], "请选择有效的开始时间和结束时间"},
		{now.Add(-time.Hour).Format(layout), fields["endAt"], "开始时间必须晚于当前时间"},
		{fields["startAt"], fields["startAt"], "结束时间必须晚于开始时间"},
		{fields["startAt"], now.Format(layout), "结束时间必须晚于开始时间"},
	} {
		input, err := json.Marshal(map[string]string{"leaveType": "sick", "reason": "test", "startAt": tc.start, "endAt": tc.end})
		if err != nil {
			t.Fatal(err)
		}
		response := send(http.MethodPost, "/student/leaves", "application/json", input, "")
		if response.Code != http.StatusBadRequest || !bytes.Contains(response.Body.Bytes(), []byte(tc.msg)) {
			t.Fatalf("validate time: %d %s", response.Code, response.Body)
		}
	}
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
	if err := db.Exec(`INSERT INTO uni_leave_requests (tenant_id,student_id,status,start_at,end_at)
		VALUES (1,2,'approved','2020-05-01 12:00:00','2020-05-02 12:00:00'),
		(1,2,'approved','2020-04-30 12:00:00','2020-05-01 00:00:00'),
		(2,2,'approved','2020-05-01 12:00:00','2020-05-02 12:00:00')`).Error; err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{"?status=approved", "?status=approved&keyword=S001"} {
		adminList := send(http.MethodGet, "/admin/leaves"+query, "", nil, "")
		body := adminList.Body.String()
		if adminList.Code != http.StatusOK || !strings.Contains(body, `"total":2`) ||
			!strings.Contains(body, `"studentName":"Student One"`) ||
			!strings.Contains(body, `"studentNo":"S001"`) || !strings.Contains(body, `"studentId":2`) ||
			strings.Contains(body, "Other tenant") {
			t.Fatalf("unexpected admin leave listing: %s", body)
		}
	}

	filtered := send(http.MethodGet, "/student/leaves?status=approved&from=2020-05-01&to=2020-05-01&includeSummary=true", "", nil, "")
	if filtered.Code != http.StatusOK || !strings.Contains(filtered.Body.String(), `"total":1`) ||
		!strings.Contains(filtered.Body.String(), `"approved":2`) {
		t.Fatalf("filtered records and unfiltered summary: %d %s", filtered.Code, filtered.Body)
	}
	badRange := send(http.MethodGet, "/student/leaves?from=2020-05-02&to=2020-05-01", "", nil, "")
	if badRange.Code != http.StatusBadRequest {
		t.Fatalf("invalid range returned %d", badRange.Code)
	}

	checkInbox := func(path, advisor string, unread int, events ...string) {
		t.Helper()
		response := send(http.MethodGet, path, "", nil, advisor)
		var inbox struct {
			Data struct {
				List   []uniLeaveNotification
				Unread int
			}
		}
		if err := json.Unmarshal(response.Body.Bytes(), &inbox); err != nil || response.Code != http.StatusOK ||
			inbox.Data.Unread != unread || len(inbox.Data.List) != len(events) {
			t.Fatalf("unexpected inbox at %s: %d %s", path, response.Code, response.Body)
		}
		for index, event := range events {
			if inbox.Data.List[index].Event != event {
				t.Fatalf("unexpected inbox event at %s: %s", path, response.Body)
			}
		}
	}
	post := func(path, body, advisor string, status int) {
		t.Helper()
		response := send(http.MethodPost, path, "application/json", []byte(body), advisor)
		if response.Code != status {
			t.Fatalf("%s: %d %s", path, response.Code, response.Body)
		}
	}
	checkInbox("/teacher/notifications", "", 2, "submitted", "submitted")
	checkInbox("/teacher/notifications", "unassigned", 0)
	post("/teacher/notifications/read", `{"id":1}`, "unassigned", http.StatusOK)
	checkInbox("/teacher/notifications", "", 2, "submitted", "submitted")
	post("/teacher/notifications/read", `{"id":1}`, "", http.StatusOK)
	checkInbox("/teacher/notifications?unreadOnly=true", "", 1)
	post("/teacher/leaves/review", `{"id":1,"action":"approved"}`, "", http.StatusOK)
	post("/teacher/leaves/review", `{"id":1,"action":"approved"}`, "", http.StatusConflict)
	post("/student/leaves/cancel", `{"id":2}`, "", http.StatusOK)
	post("/student/leaves/cancel", `{"id":2}`, "", http.StatusConflict)
	checkInbox("/student/notifications", "", 1, "approved")
	checkInbox("/teacher/notifications", "", 2, "cancelled", "submitted", "submitted")
	post("/teacher/notifications/read", `{"all":true}`, "", http.StatusOK)
	checkInbox("/teacher/notifications", "", 0, "cancelled", "submitted", "submitted")
	if err := db.Exec("UPDATE uni_students SET advisor_id=4 WHERE id=2 AND tenant_id=1").Error; err != nil {
		t.Fatal(err)
	}
	checkInbox("/teacher/notifications", "", 0)
	checkInbox("/teacher/notifications", "unassigned", 0)
	detail := send(http.MethodGet, "/student/leaves?id=1", "", nil, "")
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), `"total":1`) {
		t.Fatalf("could not open linked leave: %s", detail.Body)
	}

}
