-- V5: 创建高校管理菜单，并授权平台管理员角色 888

INSERT INTO sys_base_menus (
    created_at,
    updated_at,
    menu_level,
    parent_id,
    path,
    name,
    hidden,
    component,
    sort,
    active_name,
    keep_alive,
    default_menu,
    title,
    icon,
    close_tab
)
SELECT
    NOW(3),
    NOW(3),
    0,
    0,
    'universityTenant',
    'universityTenant',
    0,
    'view/university/tenant.vue',
    20,
    '',
    0,
    0,
    '高校管理',
    'school',
    0
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_base_menus
    WHERE name = 'universityTenant'
      AND deleted_at IS NULL
);

-- 为平台管理员分配高校管理菜单
INSERT INTO sys_authority_menus (
    sys_base_menu_id,
    sys_authority_authority_id
)
SELECT
    id,
    888
FROM sys_base_menus
WHERE name = 'universityTenant'
  AND component = 'view/university/tenant.vue'
  AND deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM sys_authority_menus AS am
      WHERE am.sys_base_menu_id = sys_base_menus.id
        AND am.sys_authority_authority_id = 888
  );
