# 高校管理员切换辅导员工作台

平台级运维账号在「高校账号管理」编辑高校管理员，填写该校教师工号。关联保存到 `uni_memberships.advisor_teacher_id`。清空工号解除关联，省略字段保留原关联。同一教师档案只能关联一个高校管理员账号。

高校管理员登录后通过头像菜单切换到 `/#/advisor`，在辅导员工作台可切回后台。切换沿用管理员会话，每次请求重新验证成员身份和启用的教师档案。辅导员工作台沿用教师端的数据查询、证明下载及审批规则，仅处理该教师负责的学生。

`GET /uni/advisor-workspace/status` 返回 `data.available`。同路径下的 `me`、`classes`、`students`、`leaves`、`leaves/:id/proof` 和 `POST leaves/review` 均要求有效关联。上线前应用数据库迁移 `000029_university_advisor_workspace`，并在部署后重新加载权限。
