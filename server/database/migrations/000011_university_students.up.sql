-- V10 uni_classes only has PRIMARY KEY(id); add composite parent key for tenant-safe FK.
ALTER TABLE uni_classes ADD UNIQUE KEY uk_uni_classes_tenant_id (tenant_id,id);
CREATE TABLE uni_students (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
 tenant_id BIGINT UNSIGNED NOT NULL,
 student_no VARCHAR(64) NOT NULL,
 name VARCHAR(191) NOT NULL,
 gender TINYINT NOT NULL DEFAULT 0,
 phone VARCHAR(32) NOT NULL DEFAULT '',
 email VARCHAR(191) NOT NULL DEFAULT '',
 class_id BIGINT UNSIGNED NOT NULL,
 enrollment_year SMALLINT UNSIGNED NOT NULL,
 status TINYINT NOT NULL DEFAULT 1,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
 PRIMARY KEY(id),
 UNIQUE KEY uk_uni_students_tenant_no(tenant_id,student_no),
 KEY idx_uni_students_class(tenant_id,class_id),
 KEY idx_uni_students_name(tenant_id,name),
 CONSTRAINT fk_uni_students_class FOREIGN KEY(tenant_id,class_id) REFERENCES uni_classes(tenant_id,id) ON DELETE RESTRICT,
 CONSTRAINT chk_uni_students_gender CHECK(gender IN (0,1,2)),
 CONSTRAINT chk_uni_students_status CHECK(status IN (1,2)),
 CONSTRAINT chk_uni_students_year CHECK(enrollment_year BETWEEN 1900 AND 2200)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO sys_base_menus
 (created_at,updated_at,parent_id,path,name,component,sort,hidden,title,icon)
VALUES (NOW(3),NOW(3),0,'universityStudent','universityStudent',
 'view/university/student.vue',40,0,'学生管理','user');
INSERT INTO sys_authority_menus(sys_base_menu_id,sys_authority_authority_id)
 SELECT id,2001 FROM sys_base_menus WHERE name='universityStudent';
INSERT INTO casbin_rule(ptype,v0,v1,v2)
 SELECT 'p','2001',p.path,p.method FROM (
 SELECT '/uni/students' path,'GET' method UNION ALL
 SELECT '/uni/students','POST' UNION ALL
 SELECT '/uni/students','PUT' UNION ALL
 SELECT '/uni/students','DELETE'
 ) p WHERE NOT EXISTS (
 SELECT 1 FROM casbin_rule c WHERE c.ptype='p' AND c.v0='2001' AND c.v1=p.path AND c.v2=p.method);
