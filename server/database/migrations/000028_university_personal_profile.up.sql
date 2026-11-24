INSERT INTO sys_authority_menus (sys_base_menu_id, sys_authority_authority_id)
SELECT m.id, a.authority_id
FROM sys_base_menus m
CROSS JOIN sys_authorities a
WHERE m.name = 'person' AND m.component = 'view/person/person.vue'
  AND m.deleted_at IS NULL AND a.deleted_at IS NULL AND a.authority_id IN (2001, 2002)
  AND NOT EXISTS (
    SELECT 1 FROM sys_authority_menus am
    WHERE am.sys_base_menu_id = m.id AND am.sys_authority_authority_id = a.authority_id
);

INSERT INTO casbin_rule (ptype, v0, v1, v2)
SELECT 'p', a.authority_id, permissions.path, permissions.method
FROM sys_authorities a
CROSS JOIN (
    SELECT '/user/setSelfInfo' AS path, 'PUT' AS method
    UNION ALL SELECT '/user/changePassword', 'POST'
) permissions
WHERE a.deleted_at IS NULL AND a.authority_id IN (2001, 2002)
  AND NOT EXISTS (
    SELECT 1 FROM casbin_rule c
    WHERE c.ptype = 'p' AND c.v0 = CAST(a.authority_id AS CHAR)
      AND c.v1 = permissions.path AND c.v2 = permissions.method
);
