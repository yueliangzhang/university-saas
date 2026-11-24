-- Validate all three tables before changing any of their indexes.
CREATE TEMPORARY TABLE uni_org_name_upgrade_check (
    duplicate_groups BIGINT NOT NULL,
    CONSTRAINT chk_uni_org_names_unique_before_upgrade CHECK (duplicate_groups = 0)
);

INSERT INTO uni_org_name_upgrade_check (duplicate_groups)
SELECT COUNT(*)
FROM (
    SELECT tenant_id, name FROM uni_colleges GROUP BY tenant_id, name HAVING COUNT(*) > 1
    UNION ALL
    SELECT tenant_id, name FROM uni_majors GROUP BY tenant_id, name HAVING COUNT(*) > 1
    UNION ALL
    SELECT tenant_id, name FROM uni_classes GROUP BY tenant_id, name HAVING COUNT(*) > 1
) duplicates;

DROP TEMPORARY TABLE uni_org_name_upgrade_check;

ALTER TABLE uni_colleges ADD UNIQUE KEY uk_uni_colleges_tenant_name (tenant_id, name);
ALTER TABLE uni_majors ADD UNIQUE KEY uk_uni_majors_tenant_name (tenant_id, name);
ALTER TABLE uni_classes ADD UNIQUE KEY uk_uni_classes_tenant_name (tenant_id, name);
