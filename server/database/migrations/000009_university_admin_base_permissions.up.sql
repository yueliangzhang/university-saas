INSERT INTO casbin_rule (ptype, v0, v1, v2)
SELECT 'p', '2001', '/menu/getMenu', 'POST'
WHERE NOT EXISTS (
    SELECT 1 FROM casbin_rule
    WHERE ptype = 'p'
      AND v0 = '2001'
      AND v1 = '/menu/getMenu'
      AND v2 = 'POST'
);

INSERT INTO casbin_rule (ptype, v0, v1, v2)
SELECT 'p', '2001', '/user/getUserInfo', 'GET'
WHERE NOT EXISTS (
    SELECT 1 FROM casbin_rule
    WHERE ptype = 'p'
      AND v0 = '2001'
      AND v1 = '/user/getUserInfo'
      AND v2 = 'GET'
);
