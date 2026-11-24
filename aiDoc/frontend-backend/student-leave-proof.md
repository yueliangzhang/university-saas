# 学生请假证明接口

- `POST /uni/student-portal/leaves` 支持原 JSON 请求。上传附件时改用 `multipart/form-data`，字段为 `leaveType`、`reason`、`startAt`、`endAt` 和可重复的选填文件字段 `proof`。最多 5 个 PDF、JPG 或 PNG，每个不超过 5 MiB。
- 两种请求都校验请假类型、去除首尾空白后的原因（必填，最多 2000 字节）及时间。时间格式为 `YYYY-MM-DD HH:mm:ss`。开始时间须晚于数据库当前时间，结束时间须晚于开始时间。
- 学生端与教师端请假列表返回布尔字段 `hasProof` 和整数字段 `proofCount`。
- 学生和教师使用各自的 Bearer 会话访问 `GET /uni/student-portal/leaves/:id/proof?index=2` 和 `GET /uni/teacher-portal/leaves/:id/proof?index=2`。序号从 1 开始，默认为 1。附件不存在或超出会话权限时返回 404。成功响应使用 `Content-Disposition` 携带文件名。
