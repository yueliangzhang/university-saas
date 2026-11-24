INSERT INTO casbin_rule (ptype, v0, v1, v2)
SELECT 'p', '2001', p.path, p.method
FROM (
    SELECT '/uni/org/classes/template' AS path, 'GET' AS method
    UNION ALL SELECT '/uni/org/classes/import/preview', 'POST'
    UNION ALL SELECT '/uni/org/classes/import/confirm', 'POST'
) p
WHERE NOT EXISTS (
    SELECT 1 FROM casbin_rule c
    WHERE c.ptype = 'p' AND c.v0 = '2001' AND c.v1 = p.path AND c.v2 = p.method
);
