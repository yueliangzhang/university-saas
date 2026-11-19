CREATE TABLE uni_teacher_accounts (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 tenant_id BIGINT UNSIGNED NOT NULL,
 teacher_id BIGINT UNSIGNED NOT NULL,
 password_hash VARCHAR(255) NOT NULL,
 status TINYINT NOT NULL DEFAULT 1,
 password_changed_at DATETIME(3) NULL,
 last_login_at DATETIME(3) NULL,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
 UNIQUE KEY uk_teacher_account(tenant_id,teacher_id),
 UNIQUE KEY uk_teacher_account_tenant_id(tenant_id,id),
 CONSTRAINT fk_teacher_account_teacher FOREIGN KEY(teacher_id) REFERENCES uni_teachers(id) ON DELETE RESTRICT,
 CONSTRAINT chk_teacher_account_status CHECK(status IN (1,2))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE uni_teacher_sessions (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 tenant_id BIGINT UNSIGNED NOT NULL,
 account_id BIGINT UNSIGNED NOT NULL,
 token_hash CHAR(64) NOT NULL,
 expires_at DATETIME(3) NOT NULL,
 revoked_at DATETIME(3) NULL,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 UNIQUE KEY uk_teacher_session_token(token_hash),
 KEY idx_teacher_session_account(tenant_id,account_id),
 CONSTRAINT fk_teacher_session_account FOREIGN KEY(tenant_id,account_id) REFERENCES uni_teacher_accounts(tenant_id,id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE uni_class_advisors (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 tenant_id BIGINT UNSIGNED NOT NULL,
 class_id BIGINT UNSIGNED NOT NULL,
 teacher_id BIGINT UNSIGNED NOT NULL,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 UNIQUE KEY uk_class_advisor(tenant_id,class_id),
 KEY idx_class_advisor_teacher(tenant_id,teacher_id),
 CONSTRAINT fk_advisor_class FOREIGN KEY(tenant_id,class_id) REFERENCES uni_classes(tenant_id,id) ON DELETE RESTRICT,
 CONSTRAINT fk_advisor_teacher FOREIGN KEY(teacher_id) REFERENCES uni_teachers(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
INSERT INTO sys_base_menus(created_at,updated_at,parent_id,path,name,component,sort,hidden,title,icon)
VALUES(NOW(3),NOW(3),0,'universityTeacherAccount','universityTeacherAccount','view/university/teacherAccount.vue',46,0,'教师账号与辅导员','user');
INSERT INTO sys_authority_menus(sys_base_menu_id,sys_authority_authority_id)
 SELECT id,2001 FROM sys_base_menus WHERE name='universityTeacherAccount';
INSERT INTO casbin_rule(ptype,v0,v1,v2)
 SELECT 'p','2001',x.path,x.method FROM (
 SELECT '/uni/teacher-accounts' path,'GET' method UNION ALL
 SELECT '/uni/teacher-accounts','POST' UNION ALL
 SELECT '/uni/teacher-accounts/reset-password','POST' UNION ALL
 SELECT '/uni/teacher-accounts/status','PUT' UNION ALL
 SELECT '/uni/class-advisors','GET' UNION ALL
 SELECT '/uni/class-advisors','PUT'
 ) x WHERE NOT EXISTS(SELECT 1 FROM casbin_rule c WHERE c.ptype='p' AND c.v0='2001' AND c.v1=x.path AND c.v2=x.method);
