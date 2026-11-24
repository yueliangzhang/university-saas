-- Before rollback, verify no accounts or memberships use role 2002. Do not delete accounts automatically.
START TRANSACTION;
DELETE FROM casbin_rule WHERE ptype='p' AND v0 IN ('2001','2002') AND v1 IN ('/uni/monitor/leaves','/uni/monitor/options','/uni/monitor/export','/uni/monitor/dashboard','/uni/monitor/detail');
DELETE FROM casbin_rule WHERE ptype='p' AND v0='2002' AND v1 IN ('/menu/getMenu','/user/getUserInfo');
DELETE am FROM sys_authority_menus am JOIN sys_base_menus m ON m.id=am.sys_base_menu_id WHERE m.name='universityLeaveMonitor';
DELETE FROM sys_base_menus WHERE name='universityLeaveMonitor';
UPDATE sys_base_menus SET title='高校管理员管理' WHERE name='universityAdmin';
DELETE FROM sys_authorities WHERE authority_id=2002;
COMMIT;
