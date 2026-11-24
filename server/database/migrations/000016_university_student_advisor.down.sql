-- Rollback V17.7.4 student advisor assignment

ALTER TABLE uni_students
    DROP FOREIGN KEY fk_uni_students_advisor,
    DROP INDEX idx_uni_students_tenant_advisor,
    DROP COLUMN advisor_id;

ALTER TABLE uni_teachers
    DROP INDEX uk_uni_teachers_tenant_id;
