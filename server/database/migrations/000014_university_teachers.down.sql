DELETE FROM casbin_rule WHERE ptype='p' AND v0='2001' AND v1='/uni/teachers' AND v2 IN ('GET','POST','PUT','DELETE');
DELETE am FROM sys_authority_menus am JOIN sys_base_menus bm ON bm.id=am.sys_base_menu_id WHERE bm.name='universityTeacher' AND am.sys_authority_authority_id=2001;
DELETE FROM sys_base_menus WHERE name='universityTeacher' AND component='view/university/teacher.vue';
DROP TABLE IF EXISTS uni_teachers;
