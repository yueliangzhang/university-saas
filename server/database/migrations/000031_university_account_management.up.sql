START TRANSACTION;
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', '888', '/uni/admin/delete', 'DELETE'),
    ('p', '2002', '/uni/advisor-workspace/status', 'GET'),
    ('p', '2002', '/uni/advisor-workspace/me', 'GET'),
    ('p', '2002', '/uni/advisor-workspace/classes', 'GET'),
    ('p', '2002', '/uni/advisor-workspace/students', 'GET'),
    ('p', '2002', '/uni/advisor-workspace/leaves', 'GET'),
    ('p', '2002', '/uni/advisor-workspace/leaves/:id/proof', 'GET'),
    ('p', '2002', '/uni/advisor-workspace/leaves/review', 'POST');
COMMIT;
