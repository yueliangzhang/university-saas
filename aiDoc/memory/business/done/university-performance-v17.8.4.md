# V17.8.4 性能优化

对请假时段查询增加联合索引，合并部分 Dashboard 统计查询，加入按高校和角色隔离的可选 Redis 缓存。补充查询超时、连接池闲置回收、流式 Excel 导出、统计图表异步加载，以及 Nginx 缓存和 k6 验收配置。部署与测试口径记录在 `deploy/performance/README.md`。
