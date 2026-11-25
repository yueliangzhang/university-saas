ALTER TABLE uni_leave_requests
    ADD INDEX idx_uni_leave_tenant_end_start (tenant_id, end_at, start_at, student_id),
    ADD INDEX idx_uni_leave_tenant_status_end (tenant_id, status, end_at, start_at, student_id),
    ALGORITHM=INPLACE, LOCK=NONE;
