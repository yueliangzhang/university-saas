-- 为平台管理员角色 888 添加高校管理 API 权限
-- 仅添加不存在的策略，避免重复插入

INSERT INTO casbin_rule
    (ptype, v0, v1, v2, v3, v4, v5)
SELECT
    'p', '888', permissions.path, permissions.method, '', '', ''
FROM (
    SELECT '/uni/tenant/create' AS path, 'POST' AS method
    UNION ALL SELECT '/uni/tenant/list', 'POST'
    UNION ALL SELECT '/uni/tenant/detail', 'GET'
    UNION ALL SELECT '/uni/tenant/update', 'PUT'
    UNION ALL SELECT '/uni/tenant/status', 'PUT'
) AS permissions
WHERE NOT EXISTS (
    SELECT 1
    FROM casbin_rule AS existing
    WHERE existing.ptype = 'p'
      AND existing.v0 = '888'
      AND existing.v1 = permissions.path
      AND existing.v2 = permissions.method
);
