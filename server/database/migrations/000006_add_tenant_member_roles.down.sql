-- V6 回滚
-- 仅在没有用户使用角色 2001、且没有需要保留的成员角色数据时执行

DELETE FROM sys_authorities
WHERE authority_id = 2001
  AND authority_name = '高校管理员';

ALTER TABLE uni_memberships
DROP COLUMN role;
