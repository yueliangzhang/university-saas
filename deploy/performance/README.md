# V17.8.4 性能验证

## MySQL

在测试库应用迁移 `000030_university_performance`，然后运行同目录的 `explain.sql`。将日期和租户改为测试数据，记录执行计划中的索引和扫描行数，并检查 Performance Schema 的查询摘要。生产库如沿用逐个执行 SQL 的流程，只需应用本版本的 `000030`。索引支持减量查询，应用回滚与索引回滚可以分别执行。

本地 MySQL 8.0 用 10000 条合成记录验证了升级和回滚。样例已批准请假窗口查询的 EXPLAIN 估算扫描量从 5000 行降至 210 行，并使用了新覆盖索引。实际效果随数据分布和完整查询计划变化。

## Redis 和应用

Dashboard 仅在应用已配置 Redis 客户端时缓存 30 秒。缓存按高校、角色、周期和组织筛选隔离，仍经过现有鉴权。缓存读写失败时走数据库，报表和登录会话继续使用现有的数据来源。启用前在测试环境检查 Redis 连接及本环境专用 Redis 库，避免与其他环境共用同一库。

Dashboard 查询预算为 8 秒，报表查询预算为 12 秒。MySQL 连接池闲置连接最多保留 5 分钟。师生档案导出按行读取和写入工作簿，现有导入继续逐行解析、批量入库。

## Nginx

在测试环境把 `deploy/nginx/university-performance.conf` 合并到服务前端的 HTTPS server 块。已有同名 location 或 gzip 设置时合并对应配置。执行 `nginx -t` 后再重载。验证两个响应头：

- `/index.html` 应返回 `Cache-Control: no-cache`。
- 内容哈希资源应返回一年缓存和 `immutable`。带 `Accept-Encoding: gzip` 请求超过 1 KB 的文本资源时应返回 gzip。

## k6

预先创建测试高校和账号，准备短期测试会话，并指定已授权的测试服务。示例假设 `GVA_ADMIN_TOKEN` 已通过环境变量注入。可选 `GVA_STUDENT_TOKEN` 会加入学生列表请求。

```sh
GVA_BASE_URL=http://localhost:8080 GVA_VUS=25 k6 run deploy/performance/university.k6.js
```

先跑 25 VU 基线，再逐步提高并发，容量允许时用 `GVA_VUS=1000` 验收。脚本检查查询 P95 ≤ 200ms、Dashboard P95 ≤ 500ms 和成功率 ≥ 99.9%。使用两个测试高校账号验证组织与 Dashboard 数据隔离，使用移动端浏览器记录首屏 LCP，目标 ≤ 2.5 秒。以上为验收标准，生产环境的实测结果需另外记录。
