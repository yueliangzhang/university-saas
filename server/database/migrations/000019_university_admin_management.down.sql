START TRANSACTION;
DELETE FROM casbin_rule WHERE ptype='p' AND v0='888' AND v1 IN ('/uni/admin/edit','/uni/admin/password') AND v2='PUT';
UPDATE sys_base_menus SET parent_id=0,menu_level=0,title='高校管理',sort=20 WHERE name='universityTenant' AND deleted_at IS NULL;
DELETE am FROM sys_authority_menus am JOIN sys_base_menus m ON m.id=am.sys_base_menu_id WHERE m.name IN ('universityManagement','universityAdmin');
DELETE FROM sys_base_menus WHERE name IN ('universityManagement','universityAdmin');
COMMIT;
ALTER TABLE uni_memberships DROP COLUMN session_cutoff;
