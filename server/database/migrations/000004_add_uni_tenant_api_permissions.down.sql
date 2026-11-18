-- 回滚高校管理 API 权限
DELETE FROM casbin_rule
WHERE ptype = 'p'
  AND v0 = '888'
  AND (
      (v1 = '/uni/tenant/create' AND v2 = 'POST')
      OR (v1 = '/uni/tenant/list' AND v2 = 'POST')
      OR (v1 = '/uni/tenant/detail' AND v2 = 'GET')
      OR (v1 = '/uni/tenant/update' AND v2 = 'PUT')
      OR (v1 = '/uni/tenant/status' AND v2 = 'PUT')
  );
