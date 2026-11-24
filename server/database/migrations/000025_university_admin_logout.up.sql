INSERT INTO casbin_rule (ptype, v0, v1, v2)
SELECT 'p', '2001', '/jwt/jsonInBlacklist', 'POST'
WHERE NOT EXISTS (
    SELECT 1 FROM casbin_rule
    WHERE ptype = 'p'
      AND v0 = '2001'
      AND v1 = '/jwt/jsonInBlacklist'
      AND v2 = 'POST'
);
