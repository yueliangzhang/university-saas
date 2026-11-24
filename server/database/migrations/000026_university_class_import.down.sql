DELETE FROM casbin_rule
WHERE ptype = 'p' AND v0 = '2001'
  AND (
    (v1 = '/uni/org/classes/template' AND v2 = 'GET')
    OR (v1 IN ('/uni/org/classes/import/preview', '/uni/org/classes/import/confirm') AND v2 = 'POST')
  );
