ALTER TABLE uni_memberships
ADD CONSTRAINT fk_uni_memberships_user
FOREIGN KEY (user_id)
REFERENCES sys_users(id)
ON DELETE RESTRICT;
