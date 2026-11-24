ALTER TABLE uni_student_accounts DROP COLUMN must_change_password, DROP COLUMN failed_attempts, DROP COLUMN locked_until;
ALTER TABLE uni_teacher_accounts DROP COLUMN must_change_password, DROP COLUMN failed_attempts, DROP COLUMN locked_until;
