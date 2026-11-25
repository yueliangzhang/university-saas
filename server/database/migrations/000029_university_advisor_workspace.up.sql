ALTER TABLE uni_memberships
    ADD COLUMN advisor_teacher_id BIGINT UNSIGNED NULL,
    ADD UNIQUE KEY uk_uni_membership_advisor (tenant_id, advisor_teacher_id),
    ADD CONSTRAINT fk_uni_membership_advisor FOREIGN KEY (tenant_id, advisor_teacher_id)
        REFERENCES uni_teachers (tenant_id, id) ON DELETE RESTRICT;

INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES
    ('p', '2001', '/uni/advisor-workspace/status', 'GET'),
    ('p', '2001', '/uni/advisor-workspace/me', 'GET'),
    ('p', '2001', '/uni/advisor-workspace/classes', 'GET'),
    ('p', '2001', '/uni/advisor-workspace/students', 'GET'),
    ('p', '2001', '/uni/advisor-workspace/leaves', 'GET'),
    ('p', '2001', '/uni/advisor-workspace/leaves/:id/proof', 'GET'),
    ('p', '2001', '/uni/advisor-workspace/leaves/review', 'POST');
