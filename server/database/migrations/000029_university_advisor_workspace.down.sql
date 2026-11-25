DELETE FROM casbin_rule WHERE ptype='p' AND v0='2001' AND v1 LIKE '/uni/advisor-workspace/%';
ALTER TABLE uni_memberships
    DROP FOREIGN KEY fk_uni_membership_advisor,
    DROP INDEX uk_uni_membership_advisor,
    DROP COLUMN advisor_teacher_id;
