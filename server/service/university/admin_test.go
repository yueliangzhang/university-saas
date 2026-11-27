package university

import (
	"errors"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	systemModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	universityModel "github.com/flipped-aurora/gin-vue-admin/server/model/university"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestDeleteAccountReleasesAdvisor(t *testing.T) {
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

	if err := db.AutoMigrate(&systemModel.SysUser{}, &universityModel.Membership{}); err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{
		"ALTER TABLE uni_memberships ADD COLUMN advisor_teacher_id INTEGER",
		"ALTER TABLE uni_memberships ADD COLUMN session_cutoff INTEGER NOT NULL DEFAULT 0",
		"CREATE UNIQUE INDEX advisor_link ON uni_memberships (tenant_id, advisor_teacher_id)",
		"CREATE TABLE uni_teachers (id INTEGER PRIMARY KEY, tenant_id INTEGER, status INTEGER, teacher_no TEXT)",
		"CREATE TABLE uni_tenants (id INTEGER PRIMARY KEY, status INTEGER)",
		"INSERT INTO uni_tenants VALUES (1, 1), (2, 1)",
		"INSERT INTO sys_authorities (authority_id) VALUES (2001), (2002)",
		"INSERT INTO sys_users (id, username, authority_id, enable) VALUES (1, 'platform', 888, 1)",
		"INSERT INTO uni_teachers VALUES (10, 1, 1, 'T010'), (20, 2, 1, 'T020')",
	} {
		if err := db.Exec(query).Error; err != nil {
			t.Fatal(err)
		}
	}

	service := AdminService{}
	for _, password := range []string{"12345", "1234567890123"} {
		if _, err := service.CreateAccount(1, "invalid-length", password, "Monitor", 2002); !errors.Is(err, utils.ErrUniversityAccountPasswordLength) {
			t.Fatalf("create accepted invalid password length: %v", err)
		}
	}
	_, err = service.CreateAccountWithAdvisor(1, "bad-link", "123456", "Monitor", 2002, "T020")
	if !errors.Is(err, ErrAdvisorLink) {
		t.Fatalf("cross-tenant association: %v", err)
	}
	var count int64
	if err := db.Model(&systemModel.SysUser{}).Where("username = ?", "bad-link").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("failed creation left a user: %d, %v", count, err)
	}

	userID, err := service.CreateAccountWithAdvisor(1, "monitor", "密码校验六位", "Monitor", 2002, "T010")
	if err != nil {
		t.Fatal(err)
	}
	var membership universityModel.Membership
	if err := db.Where("user_id = ?", userID).Take(&membership).Error; err != nil {
		t.Fatal(err)
	}
	if err := service.Reset(1, membership.ID, "123456"); err != nil {
		t.Fatalf("reset rejected six-character password: %v", err)
	}
	if err := service.Reset(1, membership.ID, "1234567890123"); !errors.Is(err, utils.ErrUniversityAccountPasswordLength) {
		t.Fatalf("reset accepted thirteen-character password: %v", err)
	}
	if err := service.DeleteAccount(2, membership.ID); !errors.Is(err, ErrMemberNotFound) {
		t.Fatalf("cross-tenant deletion: %v", err)
	}
	if err := service.DeleteAccount(1, membership.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&systemModel.SysUser{}, userID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("deleted account still active: %v", err)
	}
	if err := db.Model(&systemModel.SysUserAuthority{}).Where("sys_user_id = ?", userID).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("deleted account kept role assignments: %d, %v", count, err)
	}
	if _, err := service.CreateAccountWithAdvisor(1, "admin", "123456789012", "Admin", 2001, "T010"); err != nil {
		t.Fatalf("deleted account kept advisor association: %v", err)
	}
}
