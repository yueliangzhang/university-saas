# 师生门户登录接口

学生和教师登录接口 `POST /uni/student-portal/login` 与 `POST /uni/teacher-portal/login` 接受字符串 `tenantCode`。该值对应 `uni_tenants.code`，去除首尾空白后用于核对账号所属高校。学生提交 `studentNo`，教师提交 `teacherNo`，两者均须提交 `password`。

已有客户端可以继续发送数字 `tenantId`。同时提供非空 `tenantCode` 时以编号为准，即使该编号找不到高校也不会回退到数字 ID。

## 统一网页入口

四类账号统一从 `/#/login` 登录并填写高校编号。高校管理员和请假监督员使用后台用户名，学生和教师分别使用学号和工号。后台身份在签发 JWT 前校验角色、账号状态以及有效高校成员关系；师生沿用各自门户会话。高校统一入口无需验证码，连续失败会触发短时限流。

`POST /uni/portal/login` 接受 `tenantCode`、`account`、`password`和可选的 `role`（`student`、`teacher`、`tenant_admin` 或 `leave_monitor`）。后台登录返回现有 `user` 和 `token`，师生登录返回门户的 `role` 和 `token`。多个身份通过验证时，仅返回 `roles`，用户选择后重新验证凭据。

旧师生登录 API 继续支持已有客户端。旧网页登录路径在未登录时导向统一入口。平台级运维账号从 `/#/platform/login` 登录，沿用平台验证码策略。高校管理员和请假监督员直接调用 `POST /base/login` 时也必须提供有效的 `tenantCode`。
