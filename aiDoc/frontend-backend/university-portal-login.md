# 师生门户登录接口

学生和教师登录接口 `POST /uni/student-portal/login` 与 `POST /uni/teacher-portal/login` 接受字符串 `tenantCode`。该值对应 `uni_tenants.code`，去除首尾空白后用于核对账号所属高校。学生提交 `studentNo`，教师提交 `teacherNo`，两者均须提交 `password`。

已有客户端可以继续发送数字 `tenantId`。同时提供非空 `tenantCode` 时以编号为准，即使该编号找不到高校也不会回退到数字 ID。
