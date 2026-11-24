INSERT INTO casbin_rule(ptype,v0,v1,v2) SELECT 'p','2001',p.path,p.method FROM (
SELECT '/uni/students/template' path,'GET' method UNION ALL
SELECT '/uni/students/export' path,'GET' method UNION ALL
SELECT '/uni/students/import/preview' path,'POST' method UNION ALL
SELECT '/uni/students/import/confirm' path,'POST' method UNION ALL
SELECT '/uni/teachers/template' path,'GET' method UNION ALL
SELECT '/uni/teachers/export' path,'GET' method UNION ALL
SELECT '/uni/teachers/import/preview' path,'POST' method UNION ALL
SELECT '/uni/teachers/import/confirm' path,'POST' method
) p WHERE NOT EXISTS (SELECT 1 FROM casbin_rule c WHERE c.ptype='p' AND c.v0='2001' AND c.v1=p.path AND c.v2=p.method);
