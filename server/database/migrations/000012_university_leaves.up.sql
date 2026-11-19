-- V12: immutable student identity linkage; existing V11 physical deletion is blocked when referenced.
ALTER TABLE uni_students ADD UNIQUE KEY uk_uni_students_tenant_id (tenant_id,id);
CREATE TABLE uni_leave_requests (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
 tenant_id BIGINT UNSIGNED NOT NULL,
 student_id BIGINT UNSIGNED NOT NULL,
 leave_type VARCHAR(20) NOT NULL,
 reason VARCHAR(2000) NOT NULL,
 start_at DATETIME NOT NULL,
 end_at DATETIME NOT NULL,
 status VARCHAR(20) NOT NULL DEFAULT 'pending',
 review_note VARCHAR(1000) NOT NULL DEFAULT '',
 reviewer_id BIGINT UNSIGNED NULL,
 reviewed_at DATETIME NULL,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
 PRIMARY KEY(id),
 KEY idx_uni_leave_tenant_status (tenant_id,status,created_at),
 KEY idx_uni_leave_student (tenant_id,student_id,created_at),
 CONSTRAINT fk_uni_leave_student FOREIGN KEY(tenant_id,student_id) REFERENCES uni_students(tenant_id,id) ON DELETE RESTRICT,
 CONSTRAINT chk_uni_leave_time CHECK(end_at>start_at),
 CONSTRAINT chk_uni_leave_type CHECK(leave_type IN ('sick','personal','other')),
 CONSTRAINT chk_uni_leave_status CHECK(status IN ('pending','approved','rejected','cancelled'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
INSERT INTO sys_base_menus (created_at,updated_at,parent_id,path,name,component,sort,hidden,title,icon)
VALUES (NOW(3),NOW(3),0,'universityLeave','universityLeave','view/university/leave.vue',50,0,'学生请假管理','calendar');
INSERT INTO sys_authority_menus(sys_base_menu_id,sys_authority_authority_id)
 SELECT id,2001 FROM sys_base_menus WHERE name='universityLeave';
INSERT INTO casbin_rule(ptype,v0,v1,v2)
 SELECT 'p','2001',p.path,p.method FROM (
 SELECT '/uni/leaves' path,'GET' method UNION ALL
 SELECT '/uni/leaves','POST' UNION ALL
 SELECT '/uni/leaves','PUT' UNION ALL
 SELECT '/uni/leaves/review','POST' UNION ALL
 SELECT '/uni/leaves/cancel','POST'
 ) p WHERE NOT EXISTS (SELECT 1 FROM casbin_rule c WHERE c.ptype='p' AND c.v0='2001' AND c.v1=p.path AND c.v2=p.method);
