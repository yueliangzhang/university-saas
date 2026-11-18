CREATE TABLE uni_tenants (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code VARCHAR(64) NOT NULL COMMENT '高校唯一编码',
    name VARCHAR(200) NOT NULL COMMENT '高校名称',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1正常 0停用',
    plan VARCHAR(32) NOT NULL DEFAULT 'basic',
    expires_at DATETIME(3) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
        ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_uni_tenants_code (code),
    KEY idx_uni_tenants_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_unicode_ci;
