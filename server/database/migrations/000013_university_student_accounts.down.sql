DELETE FROM casbin_rule WHERE ptype='p' AND v0='2001' AND
 ((v1='/uni/student-accounts' AND v2 IN ('GET','POST')) OR
 (v1='/uni/student-accounts/reset-password' AND v2='POST') OR
 (v1='/uni/student-accounts/status' AND v2='PUT'));
DELETE am FROM sys_authority_menus am JOIN sys_base_menus bm ON bm.id=am.sys_base_menu_id
 WHERE bm.name='universityStudentAccount' AND am.sys_authority_authority_id=2001;
DELETE FROM sys_base_menus WHERE name='universityStudentAccount' AND component='view/university/studentAccount.vue';
DROP TABLE IF EXISTS uni_student_sessions;
DROP TABLE IF EXISTS uni_student_accounts;
