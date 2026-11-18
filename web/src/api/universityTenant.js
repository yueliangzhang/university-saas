import service from '@/utils/request'

// 创建高校
export const createTenant = (data) => {
  return service({
    url: '/uni/tenant/create',
    method: 'post',
    data
  })
}

// 分页查询高校
export const getTenantList = (data) => {
  return service({
    url: '/uni/tenant/list',
    method: 'post',
    data
  })
}

// 查询高校详情
export const getTenantDetail = (id) => {
  return service({
    url: '/uni/tenant/detail',
    method: 'get',
    params: { id }
  })
}

// 修改高校
export const updateTenant = (data) => {
  return service({
    url: '/uni/tenant/update',
    method: 'put',
    data
  })
}

// 启用或停用高校
export const setTenantStatus = (data) => {
  return service({
    url: '/uni/tenant/status',
    method: 'put',
    data
  })
}
