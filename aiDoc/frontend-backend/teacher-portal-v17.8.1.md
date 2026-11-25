# V17.8.1 教师端工作台

教师门户 `/uni/teacher-portal` 与关联管理员的辅导员工作台 `/uni/advisor-workspace` 使用同一套班级、学生和请假接口。查询和审批始终由当前会话确定租户和教师身份，并核对学生当前的辅导员分配。

- `GET /classes` 返回每班负责学生数 `studentCount` 和当前已批准并处于请假时段内的去重学生数 `onLeave`。页面用两者计算请假率。
- `GET /students` 支持姓名或学号 `keyword` 与班级 `classId`，返回入学年份供学生详情展示。
- `GET /leaves` 支持姓名或学号 `keyword`、班级 `classId`、审批状态 `status`、生效时段重叠的日期范围 `from` 与 `to`。两端日期格式为 `YYYY-MM-DD`，须成对填写。
- `GET /leaves?includeSummary=true` 同时返回全负责范围的待审批总数及近七天每日已批准请假去重人数，与当前列表筛选独立。
- `GET /leaves?format=xlsx` 导出当前筛选结果，每次最多 10000 条。
- `POST /leaves/review` 可传单条 `id` 或互不重复的 `ids`（最多 50 个）。每批使用同一审批动作和备注。驳回必须填写原因。只要一条记录不存在、已处理或超出负责范围，整批回滚并返回 409。

本版本无需数据库迁移，沿用仓库中 `deploy/releases/v17.8.0` 的安装和回滚脚本，安装参数传入 V17.8.1 的发布提交。
