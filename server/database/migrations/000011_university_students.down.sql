DELETE FROM casbin_rule WHERE ptype='p' AND v0='2001' AND v1='/uni/students' AND v2 IN ('GET','POST','PUT','DELETE');
DELETE am FROM sys_authority_menus am JOIN sys_base_menus bm ON bm.id=am.sys_base_menu_id
 WHERE bm.name='universityStudent' AND am.sys_authority_authority_id=2001;
DELETE FROM sys_base_menus WHERE name='universityStudent' AND component='view/university/student.vue';
DROP TABLE IF EXISTS uni_students;
ALTER TABLE uni_classes DROP INDEX uk_uni_classes_tenant_id;
