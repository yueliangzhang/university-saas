-- Roll back only when these rules were introduced by V7 and not pre-existing.
DELETE FROM casbin_rule
WHERE ptype='p' AND v0='888' AND (
 (v1='/uni/admin/create' AND v2='POST') OR
 (v1='/uni/admin/list' AND v2='POST') OR
 (v1='/uni/admin/status' AND v2='PUT')
);
ALTER TABLE sys_users DROP INDEX uk_uni_sys_users_username;
