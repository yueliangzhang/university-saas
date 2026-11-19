-- Rollback assumes V10 created these menu and policy rows.
DELETE FROM casbin_rule WHERE ptype='p' AND v0='2001'
 AND v1 IN ('/uni/org/colleges','/uni/org/majors','/uni/org/classes')
 AND v2 IN ('GET','POST','PUT','DELETE');
DELETE am FROM sys_authority_menus am JOIN sys_base_menus bm
 ON bm.id=am.sys_base_menu_id
 WHERE bm.name='universityOrganization' AND am.sys_authority_authority_id=2001;
DELETE FROM sys_base_menus WHERE name='universityOrganization'
 AND component='view/university/organization.vue';
DROP TABLE IF EXISTS uni_classes;
DROP TABLE IF EXISTS uni_majors;
DROP TABLE IF EXISTS uni_colleges;
