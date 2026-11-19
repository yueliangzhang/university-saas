-- Student credentials are separate from GVA sys_users. Student login uses its own scoped session.
CREATE TABLE uni_student_accounts (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
 tenant_id BIGINT UNSIGNED NOT NULL,
 student_id BIGINT UNSIGNED NOT NULL,
 password_hash VARCHAR(255) NOT NULL,
 status TINYINT NOT NULL DEFAULT 1,
 password_changed_at DATETIME(3) NULL,
 last_login_at DATETIME(3) NULL,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
 PRIMARY KEY(id),
 UNIQUE KEY uk_uni_student_account_student(tenant_id,student_id),
 UNIQUE KEY uk_uni_student_account_tenant_id(tenant_id,id),
 CONSTRAINT fk_uni_student_account_student FOREIGN KEY(tenant_id,student_id) REFERENCES uni_students(tenant_id,id) ON DELETE RESTRICT,
 CONSTRAINT chk_uni_student_account_status CHECK(status IN (1,2))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE uni_student_sessions (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
 tenant_id BIGINT UNSIGNED NOT NULL,
 account_id BIGINT UNSIGNED NOT NULL,
 token_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 expires_at DATETIME(3) NOT NULL,
 revoked_at DATETIME(3) NULL,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 PRIMARY KEY(id),
 UNIQUE KEY uk_uni_student_session_token(token_hash),
 KEY idx_uni_student_session_expiry(expires_at),
 KEY idx_uni_student_session_account(tenant_id,account_id,expires_at),
 CONSTRAINT fk_uni_student_session_account FOREIGN KEY(tenant_id,account_id) REFERENCES uni_student_accounts(tenant_id,id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
INSERT INTO sys_base_menus (created_at,updated_at,parent_id,path,name,component,sort,hidden,title,icon)
VALUES (NOW(3),NOW(3),0,'universityStudentAccount','universityStudentAccount','view/university/studentAccount.vue',60,0,'学生账号管理','lock');
INSERT INTO sys_authority_menus(sys_base_menu_id,sys_authority_authority_id)
 SELECT id,2001 FROM sys_base_menus WHERE name='universityStudentAccount';
INSERT INTO casbin_rule(ptype,v0,v1,v2)
 SELECT 'p','2001',p.path,p.method FROM (
 SELECT '/uni/student-accounts' path,'GET' method UNION ALL
 SELECT '/uni/student-accounts','POST' UNION ALL
 SELECT '/uni/student-accounts/reset-password','POST' UNION ALL
 SELECT '/uni/student-accounts/status','PUT'
 ) p WHERE NOT EXISTS (SELECT 1 FROM casbin_rule c WHERE c.ptype='p' AND c.v0='2001' AND c.v1=p.path AND c.v2=p.method);
