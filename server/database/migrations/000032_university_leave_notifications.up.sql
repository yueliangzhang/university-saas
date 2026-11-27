CREATE TABLE uni_leave_notifications (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    tenant_id BIGINT UNSIGNED NOT NULL,
    leave_id BIGINT UNSIGNED NOT NULL,
    event VARCHAR(16) NOT NULL,
    recipient_kind VARCHAR(8) NOT NULL,
    recipient_id BIGINT UNSIGNED NOT NULL,
    read_at DATETIME(3) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_uni_notification_event (tenant_id, leave_id, event),
    KEY idx_uni_notification_inbox (tenant_id, recipient_kind, recipient_id, id),
    KEY idx_uni_notification_unread (tenant_id, recipient_kind, recipient_id, read_at),
    CONSTRAINT fk_uni_notification_leave FOREIGN KEY (leave_id)
        REFERENCES uni_leave_requests (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', '2001', '/uni/advisor-workspace/notifications', 'GET'),
    ('p', '2001', '/uni/advisor-workspace/notifications/read', 'POST'),
    ('p', '2002', '/uni/advisor-workspace/notifications', 'GET'),
    ('p', '2002', '/uni/advisor-workspace/notifications/read', 'POST');
