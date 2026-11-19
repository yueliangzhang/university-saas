CREATE TABLE uni_teachers (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
 tenant_id BIGINT UNSIGNED NOT NULL,
 teacher_no VARCHAR(64) NOT NULL,
 name VARCHAR(191) NOT NULL,
 gender TINYINT NOT NULL DEFAULT 0,
 phone VARCHAR(32) NOT NULL DEFAULT '',
 email VARCHAR(191) NOT NULL DEFAULT '',
 college_id BIGINT UNSIGNED NOT NULL,
 title VARCHAR(100) NOT NULL DEFAULT '',
 status TINYINT NOT NULL DEFAULT 1,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
 PRIMARY KEY(id),
 UNIQUE KEY uk_uni_teachers_tenant_no(tenant_id,teacher_no),
 KEY idx_uni_teachers_college(tenant_id,college_id),
 KEY idx_uni_teachers_name(tenant_id,name),
 CONSTRAINT fk_uni_teachers_college FOREIGN KEY(tenant_id,college_id) REFERENCES uni_colleges(tenant_id,id) ON DELETE RESTRICT,
 CONSTRAINT chk_uni_teachers_gender CHECK(gender IN (0,1,2)),
 CONSTRAINT chk_uni_teachers_status CHECK(status IN (1,2))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
INSERT INTO sys_base_menus (created_at,updated_at,parent_id,path,name,component,sort,hidden,title,icon)
VALUES (NOW(3),NOW(3),0,'universityTeacher','universityTeacher','view/university/teacher.vue',45,0,'教师管理','user');
INSERT INTO sys_authority_menus(sys_base_menu_id,sys_authority_authority_id)
 SELECT id,2001 FROM sys_base_menus WHERE name='universityTeacher';
INSERT INTO casbin_rule(ptype,v0,v1,v2)
 SELECT 'p','2001',p.path,p.method FROM (
 SELECT '/uni/teachers' path,'GET' method UNION ALL
 SELECT '/uni/teachers','POST' UNION ALL
 SELECT '/uni/teachers','PUT' UNION ALL
 SELECT '/uni/teachers','DELETE'
 ) p WHERE NOT EXISTS (SELECT 1 FROM casbin_rule c WHERE c.ptype='p' AND c.v0='2001' AND c.v1=p.path AND c.v2=p.method);
