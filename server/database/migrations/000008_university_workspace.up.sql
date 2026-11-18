-- Preconditions: role 2001 exists, menu name universityWorkspace unused.
-- A dedicated menu; never grant platform menu universityTenant to role 2001.
INSERT INTO sys_base_menus
  (created_at, updated_at, parent_id, path, name, component, sort, hidden, title, icon)
VALUES
  (NOW(3), NOW(3), 0, 'universityWorkspace', 'universityWorkspace',
   'view/university/workspace.vue', 20, 0, '高校工作台', 'school');

INSERT INTO sys_authority_menus (sys_base_menu_id, sys_authority_authority_id)
SELECT id, 2001 FROM sys_base_menus WHERE name = 'universityWorkspace';

UPDATE sys_authorities SET default_router = 'universityWorkspace', updated_at = NOW(3)
WHERE authority_id = 2001;

INSERT INTO casbin_rule (ptype, v0, v1, v2)
SELECT 'p', '2001', '/uni/workspace/context', 'GET'
WHERE NOT EXISTS (
 SELECT 1 FROM casbin_rule
 WHERE ptype = 'p' AND v0 = '2001' AND v1 = '/uni/workspace/context' AND v2 = 'GET'
);
