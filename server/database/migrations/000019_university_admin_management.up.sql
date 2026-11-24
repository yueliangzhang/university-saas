ALTER TABLE uni_memberships ADD COLUMN session_cutoff BIGINT NOT NULL DEFAULT 0;
START TRANSACTION;
INSERT INTO sys_base_menus (created_at,updated_at,menu_level,parent_id,path,name,hidden,component,sort,active_name,keep_alive,default_menu,title,icon,close_tab) VALUES (NOW(3),NOW(3),0,0,'universityManagement','universityManagement',0,'view/university/management.vue',20,'',0,0,'高校管理','school',0);
SET @parent=LAST_INSERT_ID();
UPDATE sys_base_menus SET parent_id=@parent,menu_level=1,title='高校租户管理',sort=1 WHERE name='universityTenant' AND deleted_at IS NULL;
INSERT INTO sys_base_menus (created_at,updated_at,menu_level,parent_id,path,name,hidden,component,sort,active_name,keep_alive,default_menu,title,icon,close_tab) VALUES (NOW(3),NOW(3),1,@parent,'universityAdmin','universityAdmin',0,'view/university/universityAdmin.vue',2,'',0,0,'高校管理员管理','user',0);
INSERT INTO sys_authority_menus (sys_base_menu_id,sys_authority_authority_id) SELECT id,888 FROM sys_base_menus WHERE name IN ('universityManagement','universityAdmin') AND deleted_at IS NULL;
INSERT INTO casbin_rule (ptype,v0,v1,v2) VALUES ('p','888','/uni/admin/edit','PUT'),('p','888','/uni/admin/password','PUT');
COMMIT;
