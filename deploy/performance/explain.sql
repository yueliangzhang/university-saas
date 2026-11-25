-- Run against a staging copy. Adjust the tenant and dates to match the load run.
SET @tenant_id = 1;
SET @period_start = '2026-11-01 00:00:00';
SET @period_end = '2026-12-01 00:00:00';

EXPLAIN FORMAT=JSON
SELECT student_id, start_at, end_at
FROM uni_leave_requests
WHERE tenant_id=@tenant_id AND status='approved'
  AND start_at<@period_end AND end_at>@period_start;

EXPLAIN FORMAT=JSON
SELECT id, student_id, start_at, end_at
FROM uni_leave_requests
WHERE tenant_id=@tenant_id AND start_at<@period_end AND end_at>@period_start;

SELECT COUNT_STAR, ROUND(AVG_TIMER_WAIT / 1000000000, 2) AS avg_ms,
       LEFT(DIGEST_TEXT, 180) AS statement_signature
FROM performance_schema.events_statements_summary_by_digest
WHERE SCHEMA_NAME=DATABASE() AND DIGEST_TEXT LIKE '%uni_leave_requests%'
ORDER BY SUM_TIMER_WAIT DESC
LIMIT 10;
