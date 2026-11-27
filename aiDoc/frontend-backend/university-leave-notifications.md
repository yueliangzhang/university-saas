# 请假站内通知

提交和撤销请假向当时负责该学生的辅导员发通知。通过或驳回向申请学生发通知，批量审批按申请单分别记录。申请状态和通知在同一个数据库事务写入。同校同一申请的同类事件只有一条通知。没有辅导员的学生提交或撤销时跳过辅导员通知。

## 接口

学生门户、教师门户和辅导员工作台分别复用已有会话，接口前缀为 `/uni/student-portal`、`/uni/teacher-portal` 和 `/uni/advisor-workspace`。

- `GET /notifications?page=1` 返回 `{ list, total, unread, page, pageSize }`，按最新通知优先，每页 15 条。条目为 `{ id, leaveId, event, readAt, createdAt }`。
- `GET /notifications?unreadOnly=true` 仅返回 `{ unread }`。页面可见时每分钟刷新一次，重新回到页面也会刷新。
- `POST /notifications/read` 接受 `{ id }` 标记单条或 `{ all: true }` 标记全部。重复标记成功。
- 点击通知通过现有 `GET /leaves?id=...` 打开申请详情。列表与详情均检查当前数据归属。请假原因、审批备注和证明文件留在详情接口。

辅导员的通知以教师档案为收件人，同一教师的门户和关联工作台共用已读状态。请假监管页没有通知入口。学生更换辅导员后，旧辅导员看不到该学生的通知，新辅导员可通过待审批列表处理已有申请。

## 发布

先应用 `000032_university_leave_notifications`，再发布前后端并重启后端加载权限。通知从发布后发生的事件开始记录。
