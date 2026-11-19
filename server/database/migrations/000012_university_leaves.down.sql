DELETE FROM casbin_rule WHERE ptype='p' AND v0='2001' AND
 ((v1='/uni/leaves' AND v2 IN ('GET','POST','PUT')) OR (v1 IN ('/uni/leaves/review','/uni/leaves/cancel') AND v2='POST'));
DELETE am FROM sys_authority_menus am JOIN sys_base_menus bm ON bm.id=am.sys_base_menu_id
 WHERE bm.name='universityLeave' AND am.sys_authority_authority_id=2001;
DELETE FROM sys_base_menus WHERE name='universityLeave' AND component='view/university/leave.vue';
DROP TABLE IF EXISTS uni_leave_requests;
ALTER TABLE uni_students DROP INDEX uk_uni_students_tenant_id;
