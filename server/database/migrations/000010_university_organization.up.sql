-- V10: all child foreign keys include tenant_id to prevent cross-tenant references.
CREATE TABLE uni_colleges (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
 tenant_id BIGINT UNSIGNED NOT NULL,
 code VARCHAR(64) NOT NULL,
 name VARCHAR(191) NOT NULL,
 status TINYINT NOT NULL DEFAULT 1,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
 PRIMARY KEY (id),
 UNIQUE KEY uk_uni_colleges_tenant_code (tenant_id,code),
 UNIQUE KEY uk_uni_colleges_tenant_id (tenant_id,id),
 CONSTRAINT fk_uni_colleges_tenant FOREIGN KEY (tenant_id) REFERENCES uni_tenants(id) ON DELETE RESTRICT,
 CONSTRAINT chk_uni_colleges_status CHECK (status IN (1,2))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE uni_majors (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
 tenant_id BIGINT UNSIGNED NOT NULL,
 college_id BIGINT UNSIGNED NOT NULL,
 code VARCHAR(64) NOT NULL,
 name VARCHAR(191) NOT NULL,
 status TINYINT NOT NULL DEFAULT 1,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
 PRIMARY KEY (id),
 UNIQUE KEY uk_uni_majors_tenant_code (tenant_id,code),
 UNIQUE KEY uk_uni_majors_tenant_id (tenant_id,id),
 KEY idx_uni_majors_college (tenant_id,college_id),
 CONSTRAINT fk_uni_majors_college FOREIGN KEY (tenant_id,college_id) REFERENCES uni_colleges(tenant_id,id) ON DELETE RESTRICT,
 CONSTRAINT chk_uni_majors_status CHECK (status IN (1,2))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE uni_classes (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
 tenant_id BIGINT UNSIGNED NOT NULL,
 major_id BIGINT UNSIGNED NOT NULL,
 code VARCHAR(64) NOT NULL,
 name VARCHAR(191) NOT NULL,
 enrollment_year SMALLINT UNSIGNED NOT NULL,
 status TINYINT NOT NULL DEFAULT 1,
 created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
 updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
 PRIMARY KEY (id),
 UNIQUE KEY uk_uni_classes_tenant_code (tenant_id,code),
 KEY idx_uni_classes_major (tenant_id,major_id),
 CONSTRAINT fk_uni_classes_major FOREIGN KEY (tenant_id,major_id) REFERENCES uni_majors(tenant_id,id) ON DELETE RESTRICT,
 CONSTRAINT chk_uni_classes_status CHECK (status IN (1,2)),
 CONSTRAINT chk_uni_classes_year CHECK (enrollment_year BETWEEN 1900 AND 2200)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO sys_base_menus
 (created_at,updated_at,parent_id,path,name,component,sort,hidden,title,icon)
VALUES (NOW(3),NOW(3),0,'universityOrganization','universityOrganization',
 'view/university/organization.vue',30,0,'高校组织架构','school');
INSERT INTO sys_authority_menus(sys_base_menu_id,sys_authority_authority_id)
 SELECT id,2001 FROM sys_base_menus WHERE name='universityOrganization';

INSERT INTO casbin_rule(ptype,v0,v1,v2)
 SELECT 'p','2001',p.path,p.method
 FROM (
 SELECT '/uni/org/colleges' path,'GET' method UNION ALL
 SELECT '/uni/org/colleges','POST' UNION ALL
 SELECT '/uni/org/colleges','PUT' UNION ALL
 SELECT '/uni/org/colleges','DELETE' UNION ALL
 SELECT '/uni/org/majors','GET' UNION ALL
 SELECT '/uni/org/majors','POST' UNION ALL
 SELECT '/uni/org/majors','PUT' UNION ALL
 SELECT '/uni/org/majors','DELETE' UNION ALL
 SELECT '/uni/org/classes','GET' UNION ALL
 SELECT '/uni/org/classes','POST' UNION ALL
 SELECT '/uni/org/classes','PUT' UNION ALL
 SELECT '/uni/org/classes','DELETE'
 ) p
 WHERE NOT EXISTS (
 SELECT 1 FROM casbin_rule c WHERE c.ptype='p' AND c.v0='2001'
 AND c.v1=p.path AND c.v2=p.method);
