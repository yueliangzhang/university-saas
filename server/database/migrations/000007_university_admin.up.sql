-- Preflight: verify no duplicate usernames, including soft-deleted accounts.
-- A unique index on username intentionally prevents reuse after soft delete.
-- NULL usernames remain permitted by MySQL UNIQUE semantics; API rejects empty usernames.
ALTER TABLE sys_users
  ADD UNIQUE KEY uk_uni_sys_users_username (username);

-- Platform administrator role 888 is permitted to call the new routes.
INSERT INTO casbin_rule (ptype, v0, v1, v2)
SELECT 'p', '888', '/uni/admin/create', 'POST'
WHERE NOT EXISTS (SELECT 1 FROM casbin_rule WHERE ptype='p' AND v0='888' AND v1='/uni/admin/create' AND v2='POST');
INSERT INTO casbin_rule (ptype, v0, v1, v2)
SELECT 'p', '888', '/uni/admin/list', 'POST'
WHERE NOT EXISTS (SELECT 1 FROM casbin_rule WHERE ptype='p' AND v0='888' AND v1='/uni/admin/list' AND v2='POST');
INSERT INTO casbin_rule (ptype, v0, v1, v2)
SELECT 'p', '888', '/uni/admin/status', 'PUT'
WHERE NOT EXISTS (SELECT 1 FROM casbin_rule WHERE ptype='p' AND v0='888' AND v1='/uni/admin/status' AND v2='PUT');
