-- V5 回滚：删除高校管理菜单的角色关联
DELETE am
FROM sys_authority_menus AS am
INNER JOIN sys_base_menus AS m
    ON m.id = am.sys_base_menu_id
WHERE m.name = 'universityTenant'
  AND m.component = 'view/university/tenant.vue'
  AND am.sys_authority_authority_id = 888;

-- 删除高校管理菜单
DELETE FROM sys_base_menus
WHERE name = 'universityTenant'
  AND component = 'view/university/tenant.vue';
