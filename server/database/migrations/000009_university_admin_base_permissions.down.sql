DELETE FROM casbin_rule
WHERE ptype = 'p'
  AND v0 = '2001'
  AND (
    (v1 = '/menu/getMenu' AND v2 = 'POST')
    OR
    (v1 = '/user/getUserInfo' AND v2 = 'GET')
  );
