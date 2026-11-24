package university

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestStudentPortalLoginThenMe(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	previous := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() {
		global.GVA_DB = previous
		_ = sqlDB.Close()
	})

	for _, statement := range []string{
		"CREATE TABLE uni_tenants (id INTEGER PRIMARY KEY, code TEXT UNIQUE, status INTEGER)",
		"CREATE TABLE uni_students (id INTEGER PRIMARY KEY, tenant_id INTEGER, status INTEGER, student_no TEXT, name TEXT, class_id INTEGER, enrollment_year INTEGER, advisor_id INTEGER)",
		"CREATE TABLE uni_teachers (id INTEGER PRIMARY KEY, tenant_id INTEGER, name TEXT, teacher_no TEXT, status INTEGER)",
		"CREATE TABLE uni_student_accounts (id INTEGER PRIMARY KEY, tenant_id INTEGER, student_id INTEGER, password_hash TEXT, status INTEGER, must_change_password BOOLEAN, failed_attempts INTEGER, locked_until DATETIME, last_login_at DATETIME)",
		"CREATE TABLE uni_student_sessions (tenant_id INTEGER, account_id INTEGER, token_hash TEXT, expires_at DATETIME, revoked_at DATETIME)",
		"CREATE TABLE uni_teacher_accounts (id INTEGER PRIMARY KEY, tenant_id INTEGER, teacher_id INTEGER, password_hash TEXT, status INTEGER, must_change_password BOOLEAN, failed_attempts INTEGER, locked_until DATETIME, last_login_at DATETIME)",
		"CREATE TABLE uni_teacher_sessions (tenant_id INTEGER, account_id INTEGER, token_hash TEXT, expires_at DATETIME, revoked_at DATETIME)",
		"INSERT INTO uni_tenants VALUES (1, '0007', 1), (9, 'OTHER', 1)",
		"INSERT INTO uni_students VALUES (2, 1, 1, 'S001', 'Test Student', 3, 2026, NULL)",
		"INSERT INTO uni_teachers VALUES (4, 1, 'Test Teacher', 'T001', 1)",
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("testpw"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO uni_student_accounts VALUES (3, 1, 2, ?, 1, false, 0, NULL, NULL)", string(hash)).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO uni_teacher_accounts VALUES (5, 1, 4, ?, 1, false, 0, NULL, NULL)", string(hash)).Error; err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	InitStudentPortalRouter(router.Group(""))
	InitTeacherPortalRouter(router.Group(""))
	login := func(portal, payload string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/uni/"+portal+"-portal/login", strings.NewReader(payload))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		return response
	}
	for _, tc := range []struct {
		portal, payload string
		status          int
	}{
		{"student", `{"tenantId":1,"studentNo":"S001","password":"testpw"}`, http.StatusOK},
		{"student", `{"tenantCode":"OTHER","studentNo":"S001","password":"testpw"}`, http.StatusUnauthorized},
		{"teacher", `{"tenantCode":"0007","teacherNo":"T001","password":"testpw"}`, http.StatusOK},
		{"teacher", `{"tenantCode":"7","teacherNo":"T001","password":"testpw"}`, http.StatusUnauthorized},
		{"teacher", `{"tenantId":1,"tenantCode":"OTHER","teacherNo":"T001","password":"testpw"}`, http.StatusUnauthorized},
	} {
		if result := login(tc.portal, tc.payload); result.Code != tc.status {
			t.Fatalf("%s login returned %d, want %d: %s", tc.portal, result.Code, tc.status, result.Body)
		}
	}
	loginResponse := login("student", `{"tenantCode":" 0007 ","studentNo":"S001","password":"testpw"}`)

	var response struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login returned %d: %s", loginResponse.Code, loginResponse.Body)
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &response); err != nil || response.Data.Token == "" {
		t.Fatalf("missing session token: %v", err)
	}

	getMe := func() *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, "/uni/student-portal/me", nil)
		request.Header.Set("Authorization", "Bearer "+response.Data.Token)
		result := httptest.NewRecorder()
		router.ServeHTTP(result, request)
		return result
	}

	profile := getMe()
	if profile.Code != http.StatusOK || !strings.Contains(profile.Body.String(), `"studentNo":"S001"`) {
		t.Fatalf("profile returned %d: %s", profile.Code, profile.Body)
	}

	if err := db.Exec("UPDATE uni_student_accounts SET must_change_password=true WHERE id=3").Error; err != nil {
		t.Fatal(err)
	}
	profile = getMe()
	if profile.Code != http.StatusOK || !strings.Contains(profile.Body.String(), `"mustChangePassword":true`) || strings.Contains(profile.Body.String(), `"studentNo"`) {
		t.Fatalf("initial password status returned %d: %s", profile.Code, profile.Body)
	}
}
