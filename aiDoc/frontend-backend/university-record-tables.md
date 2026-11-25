# 高校档案列表排序与批量删除

学生和教师列表支持 `sortBy` 与 `sortOrder=asc|desc`。排序在分页之前执行，相同值按记录 ID 降序排列。缺少排序参数或字段不受支持时，默认按 ID 降序排列。

- `GET /uni/students` 支持 `studentNo`、`name`、`gender`、`className`、`advisorName`、`enrollmentYear`、`phone`、`status`。
- `GET /uni/teachers` 支持 `teacherNo`、`name`、`collegeName`、`title`、`phone`、`email`、`status`。
- 学院、专业、班级列表在前端对筛选结果排序，再切分页。

现有 `DELETE /uni/students`、`DELETE /uni/teachers`、`DELETE /uni/org/classes` 接受 `ids` 查询参数，例如 `?ids=1,2,3`。每批最多 100 个互异的正整数 ID。单条删除继续传 `id`。批量删除在当前高校内执行，任一记录不存在或有关联限制时整批回滚。
