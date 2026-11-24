START TRANSACTION;
INSERT INTO sys_base_menus(created_at,updated_at,parent_id,path,name,component,sort,hidden,title,icon) VALUES(NOW(3),NOW(3),0,'universityLeaveDashboard','universityLeaveDashboard','view/university/leaveDashboard.vue',20,0,'请假数据看板','calendar');
INSERT INTO sys_authority_menus(sys_base_menu_id,sys_authority_authority_id) SELECT LAST_INSERT_ID(),2002;
UPDATE sys_authorities SET default_router='universityLeaveDashboard' WHERE authority_id=2002 AND default_router='universityLeaveMonitor';
INSERT INTO casbin_rule(ptype,v0,v1,v2) VALUES ('p','2001','/uni/monitor/analytics','GET'),('p','2002','/uni/monitor/analytics','GET');
COMMIT;
