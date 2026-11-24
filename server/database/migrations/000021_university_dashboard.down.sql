START TRANSACTION;
UPDATE sys_authorities SET default_router='universityLeaveMonitor' WHERE authority_id=2002 AND default_router='universityLeaveDashboard';
DELETE FROM casbin_rule WHERE ptype='p' AND v0 IN ('2001','2002') AND v1='/uni/monitor/analytics' AND v2='GET';
DELETE am FROM sys_authority_menus am JOIN sys_base_menus m ON m.id=am.sys_base_menu_id WHERE m.name='universityLeaveDashboard' AND am.sys_authority_authority_id=2002;
DELETE FROM sys_base_menus WHERE name='universityLeaveDashboard' AND component='view/university/leaveDashboard.vue';
COMMIT;
