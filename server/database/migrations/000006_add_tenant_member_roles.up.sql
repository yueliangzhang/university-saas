-- V6: 高校成员角色与高校管理员系统角色

ALTER TABLE uni_memberships
ADD COLUMN role VARCHAR(32) NOT NULL DEFAULT 'member'
COMMENT 'tenant_admin, teacher, student, member'
AFTER user_id;

-- 如果角色 ID 2001 已存在，直接报错，避免误用其他角色
INSERT INTO sys_authorities (
    authority_id,
    authority_name,
    parent_id,
    default_router,
    created_at,
    updated_at
)
VALUES (
    2001,
    '高校管理员',
    0,
    'dashboard',
    NOW(3),
    NOW(3)
);
