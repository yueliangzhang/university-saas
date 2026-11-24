DELETE FROM sys_authority_menus
WHERE sys_authority_authority_id IN (2001, 2002)
  AND sys_base_menu_id IN (
    SELECT id FROM sys_base_menus
    WHERE name = 'person' AND component = 'view/person/person.vue'
  );

DELETE FROM casbin_rule
WHERE ptype = 'p' AND v0 IN ('2001', '2002')
  AND (
    (v1 = '/user/setSelfInfo' AND v2 = 'PUT')
    OR (v1 = '/user/changePassword' AND v2 = 'POST')
  );
