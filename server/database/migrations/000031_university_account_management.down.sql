START TRANSACTION;
DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = '888' AND v1 = '/uni/admin/delete' AND v2 = 'DELETE';
DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = '2002' AND (
    (v2 = 'GET' AND v1 IN (
        '/uni/advisor-workspace/status',
        '/uni/advisor-workspace/me',
        '/uni/advisor-workspace/classes',
        '/uni/advisor-workspace/students',
        '/uni/advisor-workspace/leaves',
        '/uni/advisor-workspace/leaves/:id/proof'
    )) OR (v2 = 'POST' AND v1 = '/uni/advisor-workspace/leaves/review')
);
COMMIT;
