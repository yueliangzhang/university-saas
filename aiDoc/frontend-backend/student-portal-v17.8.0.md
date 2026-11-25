# V17.8.0 学生端

统一登录页通过 `GET /uni/portal/school?code=...` 查询已启用高校的展示名称，只返回 `data.name`。找不到高校时该字段为空。认证和租户归属校验继续由登录接口完成。

`GET /uni/student-portal/me` 增加 `schoolName`、`collegeName`、`majorName`、`className`。学生仍通过同一门户会话访问，并保留首次强制改密流程。

`GET /uni/student-portal/leaves` 的 `from` 和 `to` 均以 `YYYY-MM-DD` 传入，须成对出现。结果包含与该日期窗口有重叠的申请，结束日期覆盖当天，边界按存储的请假时间计算。状态和日期在总数与分页之前筛选。`includeSummary=true` 增加当前学生全量的累计、待审批及已批准统计，独立于当前筛选。详情使用列表返回的申请、审批和附件信息。

本版本无需数据库迁移。生产服务器的版本化脚本位于 `deploy/releases/v17.8.0`。在服务器检出并确认发布提交后，可用 `bash deploy/releases/v17.8.0/install.sh <提交或标签>` 构建和切换，两项健康检查或前端内容校验失败时会切回上一版。用 `bash deploy/releases/v17.8.0/rollback.sh` 可回到安装时记录的版本，也可显式传入发布目录。
