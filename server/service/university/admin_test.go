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

func TestDeletedAccountCanBeRecreatedWithSameUsername(t *testing.T) {
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
		"CREATE UNIQUE INDEX uk_uni_sys_users_username ON sys_users (username)",
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
	if _, err := service.CreateAccountWithAdvisor(1, "monitor", "123456", "Monitor", 2002, "T020"); !errors.Is(err, ErrAdvisorLink) {
		t.Fatalf("recreation should reject a cross-tenant advisor: %v", err)
	}
	var oldUser systemModel.SysUser
	if err := db.Unscoped().First(&oldUser, userID).Error; err != nil || oldUser.Username != "monitor" {
		t.Fatalf("failed recreation changed the old account: %v", err)
	}
	newID, err := service.CreateAccountWithAdvisor(1, "monitor", "123456789012", "Monitor", 2002, "T010")
	if err != nil || newID == userID {
		t.Fatalf("recreation did not create a new account: %d, %v", newID, err)
	}
	var newUser systemModel.SysUser
	if err := db.First(&newUser, newID).Error; err != nil || newUser.UUID == oldUser.UUID || !utils.BcryptCheck("123456789012", newUser.Password) {
		t.Fatalf("recreated account has incorrect credentials or identity: %v", err)
	}
	var recreated universityModel.Membership
	if err := db.Where("user_id = ?", newID).Take(&recreated).Error; err != nil {
		t.Fatal(err)
	}
	invalidRole := uint(888)
	if err := service.Edit(1, recreated.ID, "Monitor", "", "", nil, &invalidRole); !errors.Is(err, ErrInvalidAccountRole) {
		t.Fatalf("edit accepted a platform role: %v", err)
	}
	for _, target := range []struct {
		authority uint
		role      string
	}{{2001, "tenant_admin"}, {2002, "leave_monitor"}} {
		if err := service.Edit(1, recreated.ID, "Updated", "", "", nil, &target.authority); err != nil {
			t.Fatal(err)
		}
		var state struct {
			Role             string
			AdvisorTeacherID uint64
			SessionCutoff    int64
		}
		if err := db.Table("uni_memberships").Where("id = ?", recreated.ID).Take(&state).Error; err != nil ||
			state.Role != target.role || state.AdvisorTeacherID != 10 || state.SessionCutoff == 0 {
			t.Fatalf("unexpected membership after switching to %d: %+v, %v", target.authority, state, err)
		}
		if err := db.First(&newUser, newID).Error; err != nil || newUser.AuthorityId != target.authority {
			t.Fatalf("primary authority not updated: %v", err)
		}
		var authorities []systemModel.SysUserAuthority
		if err := db.Where("sys_user_id = ?", newID).Find(&authorities).Error; err != nil ||
			len(authorities) != 1 || authorities[0].SysAuthorityAuthorityId != target.authority {
			t.Fatalf("unexpected authority assignments: %+v, %v", authorities, err)
		}
	}
	if _, err := service.CreateAccount(2, "monitor", "123456", "Admin", 2001); !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("active username can be reused: %v", err)
	}
}
