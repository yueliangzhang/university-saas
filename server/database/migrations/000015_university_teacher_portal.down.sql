DELETE FROM casbin_rule WHERE ptype='p' AND v0='2001' AND ((v1='/uni/teacher-accounts' AND v2 IN ('GET','POST')) OR (v1='/uni/teacher-accounts/reset-password' AND v2='POST') OR (v1='/uni/teacher-accounts/status' AND v2='PUT') OR (v1='/uni/class-advisors' AND v2 IN ('GET','PUT')));
DELETE am FROM sys_authority_menus am JOIN sys_base_menus bm ON bm.id=am.sys_base_menu_id WHERE bm.name='universityTeacherAccount' AND am.sys_authority_authority_id=2001;
DELETE FROM sys_base_menus WHERE name='universityTeacherAccount' AND component='view/university/teacherAccount.vue';
DROP TABLE IF EXISTS uni_class_advisors;
DROP TABLE IF EXISTS uni_teacher_sessions;
DROP TABLE IF EXISTS uni_teacher_accounts;
