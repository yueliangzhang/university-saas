<template>
  <section class="uni-merged-panel-root">
  <div class="uni-merged-panel">
    <el-card>
      <template #header><b>登录账号</b></template>
      <form class="mb-4 flex flex-wrap gap-2" @submit.prevent="search">
        <el-input v-model="keyword" placeholder="姓名或学号" aria-label="姓名或学号" clearable style="width:240px" @clear="search" />
        <el-select v-model="status" placeholder="全部账号状态" aria-label="账号状态" style="width:160px">
          <el-option label="全部账号状态" value="" />
          <el-option label="未开通" value="0" />
          <el-option label="正常" value="1" />
          <el-option label="停用" value="2" />
        </el-select>
        <el-button type="primary" native-type="submit">查询</el-button>
        <el-button @click="resetFilters">重置筛选</el-button>
      </form>
      <el-table v-loading="loading" :data="rows" border @sort-change="changeSort">
        <el-table-column prop="studentNo" label="学号" sortable="custom" />
        <el-table-column prop="name" label="姓名" sortable="custom" />
        <el-table-column prop="status" label="账号状态" sortable="custom">
          <template #default="{ row }">{{ row.status === 0 ? '未开通' : row.status === 1 ? '正常' : '停用' }}</template>
        </el-table-column>
        <el-table-column prop="lastLoginAt" label="最后登录" sortable="custom" />
        <el-table-column label="操作" width="270"><template #default="{row}">
          <el-button v-if="row.status===0" type="primary" link @click="openPassword(row,false)">开通</el-button>
          <el-button v-else type="primary" link @click="openPassword(row,true)">重置密码</el-button>
          <el-button v-if="row.status!==0" type="warning" link @click="toggle(row)">{{row.status===1?'停用':'启用'}}</el-button>
        </template></el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        :pager-count="5"
        layout="total, sizes, prev, pager, next"
        class="account-pagination"
        @size-change="page = 1"
      />
    </el-card>
    <el-dialog v-model="dialog" :title="reset?'重置学生密码':'开通学生账号'" width="440px">
      <p class="mb-3">{{selected?.name}}（{{selected?.studentNo}}）</p>
      <el-alert title="初始密码：123456，首次登录必须修改" type="warning" :closable="false"/>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">确认</el-button></template>
    </el-dialog>
  </div>

  </section>
</template>
<script setup>
import { ref, watch } from 'vue'
import {ElMessage,ElMessageBox} from 'element-plus'
import request from '@/utils/request'
const rows=ref([]),total=ref(0),page=ref(1),keyword=ref(''),loading=ref(false),saving=ref(false)
const dialog=ref(false),reset=ref(false),selected=ref(null),password=ref('')
const api=(url,method='get',data,params)=>request({url,method,data,params})
const pageSize = ref(20)
const appliedKeyword = ref('')
const status = ref('')
const sort = ref({})
let loadVersion = 0

const load = async () => {
  const version = ++loadVersion
  loading.value = true
  try {
    const result = await api('/uni/student-accounts', 'get', undefined, {
      page: page.value,
      pageSize: pageSize.value,
      keyword: appliedKeyword.value,
      status: status.value || undefined,
      ...sort.value
    })
    if (version !== loadVersion) return
    if (result?.code !== 0) throw new Error(result?.msg || '查询失败')
    rows.value = result.data?.list || []
    total.value = result.data?.total || 0
    const lastPage = Math.max(1, Math.ceil(total.value / pageSize.value))
    if (page.value > lastPage) page.value = lastPage
  } catch (error) {
    if (version === loadVersion) ElMessage.error(error.message || '查询失败')
  } finally {
    if (version === loadVersion) loading.value = false
  }
}

const search = () => {
  const nextKeyword = keyword.value.trim()
  if (page.value === 1 && appliedKeyword.value === nextKeyword) {
    load()
    return
  }
  page.value = 1
  appliedKeyword.value = nextKeyword
}

const resetFilters = () => {
  keyword.value = ''
  const changed = status.value !== ''
  status.value = ''
  if (changed) appliedKeyword.value = ''
  else search()
}

const changeSort = ({ prop, order }) => {
  page.value = 1
  sort.value = order
    ? { sortBy: prop, sortOrder: order === 'ascending' ? 'asc' : 'desc' }
    : {}
}

watch(status, () => { page.value = 1 }, { flush: 'sync' })
watch([page, pageSize, appliedKeyword, status, sort], load, { immediate: true })
const openPassword=(row,isReset)=>{selected.value=row;reset.value=isReset;password.value='';dialog.value=true}
const save=async()=>{saving.value=true;try{const url=reset.value?'/uni/student-accounts/reset-password':'/uni/student-accounts';const r=await api(url,'post',{studentId:selected.value.studentId,password:"123456"});if(r?.code!==0)throw Error(r?.msg||'保存失败');ElMessage.success('操作成功');dialog.value=false;password.value='';await load()}catch(e){ElMessage.error(e.message||'保存失败')}finally{saving.value=false}}
const toggle=async row=>{try{await ElMessageBox.confirm(`确定${row.status===1?'停用':'启用'}该账号？`,'确认');const r=await api('/uni/student-accounts/status','put',{studentId:row.studentId,status:row.status===1?2:1});if(r?.code!==0)throw Error(r?.msg||'操作失败');await load()}catch(e){if(e!=='cancel')ElMessage.error(e.message||'操作失败')}}
</script>

<style scoped>
.uni-merged-panel { min-width:0; }
.uni-merged-panel :deep(.el-card) { border:0; border-radius:0; box-shadow:none; }
.uni-merged-panel :deep(.el-card__body) { padding:18px 20px 22px; }
.uni-merged-panel :deep(.el-card__header) { padding:15px 20px; }
.uni-merged-panel :deep(.el-table) { --el-table-header-bg-color:#f7f9fc; --el-table-border-color:#edf1f6; }
@media(max-width:768px) { .uni-merged-panel :deep(.el-card__body) { padding:12px; } }


/* UNIVERSITY_ADMIN_V17_4_BEGIN */
/* Only applies to this Vue SFC; no global html/body or overflow changes. */
:deep(.el-card) {
  border: 1px solid #e7edf5;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(24, 49, 86, .045);
  background: #fff;
}
:deep(.el-card__header) {
  padding: 18px 22px;
  border-bottom: 1px solid #edf1f7;
  font-weight: 650;
  color: #1d2b43;
}
:deep(.el-card__body) { padding: 22px; }
:deep(.el-button) { border-radius: 9px; font-weight: 550; }
:deep(.el-button--primary) { --el-button-bg-color: #2563eb; --el-button-border-color: #2563eb; --el-button-hover-bg-color: #1d4ed8; --el-button-hover-border-color: #1d4ed8; }
:deep(.el-input__wrapper),
:deep(.el-select__wrapper),
:deep(.el-textarea__inner) {
  border-radius: 9px;
  box-shadow: 0 0 0 1px #e2e8f0 inset;
}
:deep(.el-input__wrapper.is-focus),
:deep(.el-select__wrapper.is-focused) { box-shadow: 0 0 0 1px #2563eb inset; }
:deep(.el-table) {
  --el-table-header-bg-color: #f8fafd;
  --el-table-border-color: #edf1f7;
  --el-table-row-hover-bg-color: #f5f8ff;
  border-radius: 12px;
  font-size: 13px;
}
:deep(.el-table th.el-table__cell) { color: #53647d; font-weight: 650; }
:deep(.el-table td.el-table__cell) { color: #27374f; }
:deep(.el-tag) { border-radius: 7px; font-weight: 550; }
.account-pagination { margin-top: 16px; justify-content: flex-end; flex-wrap: wrap; row-gap: 8px; }
:deep(.el-dialog) { border-radius: 18px; max-width: calc(100vw - 28px); }
:deep(.el-dialog__header) { padding-bottom: 16px; border-bottom: 1px solid #edf1f7; }
:deep(.el-form-item__label) { color: #40516a; font-weight: 550; }
@media (max-width: 768px) {
  :deep(.el-card__body) { padding: 14px; }
  :deep(.el-card__header) { padding: 14px; }
  :deep(.el-dialog) { width: calc(100vw - 24px) !important; margin: 5vh auto; }
}
/* UNIVERSITY_ADMIN_V17_4_END */
</style>
