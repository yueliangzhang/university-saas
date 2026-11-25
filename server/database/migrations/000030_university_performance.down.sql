ALTER TABLE uni_leave_requests
    DROP INDEX idx_uni_leave_tenant_end_start,
    DROP INDEX idx_uni_leave_tenant_status_end,
    ALGORITHM=INPLACE, LOCK=NONE;
