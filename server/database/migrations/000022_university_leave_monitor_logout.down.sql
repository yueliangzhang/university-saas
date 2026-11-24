DELETE FROM casbin_rule
WHERE ptype = 'p'
  AND v0 = '2002'
  AND v1 = '/jwt/jsonInBlacklist'
  AND v2 = 'POST';
