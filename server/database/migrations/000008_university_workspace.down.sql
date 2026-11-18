-- Only use this rollback if the menu/policy were created by V8.
DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = '2001'
 AND v1 = '/uni/workspace/context' AND v2 = 'GET';
UPDATE sys_authorities SET default_router = 'dashboard', updated_at = NOW(3)
 WHERE authority_id = 2001 AND default_router = 'universityWorkspace';
DELETE am FROM sys_authority_menus am
 JOIN sys_base_menus bm ON bm.id = am.sys_base_menu_id
 WHERE bm.name = 'universityWorkspace' AND am.sys_authority_authority_id = 2001;
DELETE FROM sys_base_menus WHERE name = 'universityWorkspace'
 AND component = 'view/university/workspace.vue';
