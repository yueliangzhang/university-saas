# V17.8.2 高校管理端

学生档案列表和 Excel 导出支持 `status`（1 在读、2 停用）与 `enrollmentYear`，可与班级、辅导员及关键词组合。教师档案列表补充 `studentCount` 和 `classCount`，按本校当前学生的辅导员分配统计。

学生及教师导出可传逗号分隔的 `fields`，值为导出原有表头中从零开始且互不重复的列序号。省略时仍返回原有全部列。导入结果继续使用原有预览接口，页面支持按行号、编号和错误原因查找。

`POST /uni/student-accounts/reset-password` 接受原有 `studentId` 或批量 `studentIds`。教师对应接口接受原有 `teacherId` 或批量 `teacherIds`。一批最多 100 个不同编号，同一请求仅传一种编号字段。重置和会话撤销一并在当前高校的事务内完成。

本版本无需数据库迁移。发布继续使用已有安装与回滚脚本，安装参数传入本版本提交。
