-- V17.7.4: Independent student advisor assignment

-- 1. Add tenant-scoped unique index for teachers
ALTER TABLE uni_teachers
    ADD UNIQUE KEY uk_uni_teachers_tenant_id
        (tenant_id, id);

-- 2. Add student advisor column and foreign key
ALTER TABLE uni_students
    ADD COLUMN advisor_id BIGINT UNSIGNED
        NULL DEFAULT NULL COMMENT '辅导员ID'
        AFTER class_id,
    ADD INDEX idx_uni_students_tenant_advisor
        (tenant_id, advisor_id),
    ADD CONSTRAINT fk_uni_students_advisor
        FOREIGN KEY (tenant_id, advisor_id)
        REFERENCES uni_teachers (tenant_id, id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT;

-- 3. Backfill students with exactly one class advisor
UPDATE uni_students AS s
JOIN (
    SELECT
        tenant_id,
        class_id,
        MIN(teacher_id) AS teacher_id
    FROM uni_class_advisors
    GROUP BY tenant_id, class_id
    HAVING COUNT(DISTINCT teacher_id) = 1
) AS a
    ON a.tenant_id = s.tenant_id
    AND a.class_id = s.class_id
JOIN uni_teachers AS t
    ON t.tenant_id = s.tenant_id
    AND t.id = a.teacher_id
    AND t.status = 1
SET s.advisor_id = t.id
WHERE s.advisor_id IS NULL;
