DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 IN ('2001', '2002') AND (
    (v1 = '/uni/advisor-workspace/notifications' AND v2 = 'GET') OR
    (v1 = '/uni/advisor-workspace/notifications/read' AND v2 = 'POST')
);
DROP TABLE IF EXISTS uni_leave_notifications;
