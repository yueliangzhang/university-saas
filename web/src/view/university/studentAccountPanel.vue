<template>
  <section class="uni-merged-panel-root">
  <div class="uni-merged-panel">
    <el-card>
      <template #header><b>登录账号</b></template>
      <div class="mb-4 flex gap-2"><el-input v-model="keyword" placeholder="姓名或学号" clearable style="width:240px" @keyup.enter="load"/><el-button type="primary" @click="load">查询</el-button></div>
      <el-table v-loading="loading" :data="rows" border>
        <el-table-column prop="studentNo" label="学号"/><el-table-column prop="name" label="姓名"/>
        <el-table-column label="账号状态"><template #default="{row}">{{row.status===0?'未开通':row.status===1?'正常':'停用'}}</template></el-table-column>
        <el-table-column prop="lastLoginAt" label="最后登录"/>
        <el-table-column label="操作" width="270"><template #default="{row}">
          <el-button v-if="row.status===0" type="primary" link @click="openPassword(row,false)">开通</el-button>
          <el-button v-else type="primary" link @click="openPassword(row,true)">重置密码</el-button>
          <el-button v-if="row.status!==0" type="warning" link @click="toggle(row)">{{row.status===1?'停用':'启用'}}</el-button>
        </template></el-table-column>
      </el-table>
      <el-pagination class="mt-4" background layout="total,prev,pager,next" :total="total" :page-size="20" v-model:current-page="page" @current-change="load"/>
    </el-card>
    <el-dialog v-model="dialog" :title="reset?'重置学生密码':'开通学生账号'" width="440px">
      <p class="mb-3">{{selected?.name}}（{{selected?.studentNo}}）</p>
      <el-input v-model="password" type="password" show-password placeholder="设置至少12位的初始密码" autocomplete="new-password"/>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">确认</el-button></template>
    </el-dialog>
  </div>

  </section>
</template>
<script setup>
import {ref,onMounted} from 'vue'
import {ElMessage,ElMessageBox} from 'element-plus'
import request from '@/utils/request'
const rows=ref([]),total=ref(0),page=ref(1),keyword=ref(''),loading=ref(false),saving=ref(false)
const dialog=ref(false),reset=ref(false),selected=ref(null),password=ref('')
const api=(url,method='get',data,params)=>request({url,method,data,params})
const load=async()=>{loading.value=true;try{const r=await api('/uni/student-accounts','get',undefined,{page:page.value,pageSize:20,keyword:keyword.value});if(r?.code!==0)throw Error(r?.msg||'查询失败');rows.value=r.data?.list||[];total.value=r.data?.total||0}catch(e){ElMessage.error(e.message||'查询失败')}finally{loading.value=false}}
const openPassword=(row,isReset)=>{selected.value=row;reset.value=isReset;password.value='';dialog.value=true}
const save=async()=>{if(password.value.length<12){ElMessage.warning('密码至少12位');return}saving.value=true;try{const url=reset.value?'/uni/student-accounts/reset-password':'/uni/student-accounts';const r=await api(url,'post',{studentId:selected.value.studentId,password:password.value});if(r?.code!==0)throw Error(r?.msg||'保存失败');ElMessage.success('操作成功');dialog.value=false;password.value='';await load()}catch(e){ElMessage.error(e.message||'保存失败')}finally{saving.value=false}}
const toggle=async row=>{try{await ElMessageBox.confirm(`确定${row.status===1?'停用':'启用'}该账号？`,'确认');const r=await api('/uni/student-accounts/status','put',{studentId:row.studentId,status:row.status===1?2:1});if(r?.code!==0)throw Error(r?.msg||'操作失败');await load()}catch(e){if(e!=='cancel')ElMessage.error(e.message||'操作失败')}}
onMounted(load)
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
:deep(.el-pagination) { margin-top: 16px; }
:deep(.el-dialog) { border-radius: 18px; max-width: calc(100vw - 28px); }
:deep(.el-dialog__header) { padding-bottom: 16px; border-bottom: 1px solid #edf1f7; }
:deep(.el-form-item__label) { color: #40516a; font-weight: 550; }
@media (max-width: 768px) {
  :deep(.el-card__body) { padding: 14px; }
  :deep(.el-card__header) { padding: 14px; }
  :deep(.el-dialog) { width: calc(100vw - 24px) !important; margin: 5vh auto; }
  :deep(.el-pagination) { flex-wrap: wrap; justify-content: center; row-gap: 8px; }
}
/* UNIVERSITY_ADMIN_V17_4_END */
</style>
