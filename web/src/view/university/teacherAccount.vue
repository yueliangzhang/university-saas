<template>
  <!-- UNI_ADMIN_V17_6_2 -->
  <div class="uni-admin-v1762-page">
    <header class="uni-admin-v1762-heading">
      <div class="uni-admin-v1762-heading-copy">
        <span class="uni-admin-v1762-eyebrow">UNIVERSITY MANAGEMENT</span>
        <h1>教师账号与辅导员</h1>
        <p>教师账号及班级辅导员分配</p>
      </div>
      <span class="uni-admin-v1762-heading-mark" aria-hidden="true">校务管理</span>
    </header>
    <div class="uni-admin-v1762-content">

 <div class="teacher-account"><el-card shadow="never"><template #header><b>教师账号与辅导员分配</b></template><el-alert title="教师使用独立账号登录；每个班级目前只允许指定一名负责审批的辅导员。" type="info" :closable="false" style="margin-bottom:16px"/>
 <h3>教师账号</h3><el-table v-loading="loading" :data="teachers" style="width:100%"><el-table-column prop="teacherNo" label="工号"/><el-table-column prop="name" label="姓名"/><el-table-column label="账号状态"><template #default="{row}"><el-tag v-if="account(row.id)" :type="account(row.id).status===1?'success':'info'">{{account(row.id).status===1?'正常':'停用'}}</el-tag><el-tag v-else type="warning">未开通</el-tag></template></el-table-column><el-table-column label="操作" min-width="250"><template #default="{row}"><el-button v-if="!account(row.id)" type="primary" link @click="passwordDialog(row,false)">开通账号</el-button><template v-else><el-button type="primary" link @click="passwordDialog(row,true)">重置密码</el-button><el-button type="warning" link @click="toggle(row)">{{account(row.id).status===1?'停用':'启用'}}</el-button></template></template></el-table-column></el-table>
 <el-divider/><h3>班级辅导员</h3><el-table :data="classes" style="width:100%"><el-table-column prop="name" label="班级"/><el-table-column prop="code" label="班级代码"/><el-table-column label="负责教师" min-width="250"><template #default="{row}"><el-select :model-value="advisor(row.id)" clearable placeholder="未分配" style="width:230px" @change="v=>setAdvisor(row.id,v)"><el-option v-for="t in teachers.filter(x=>x.status===1)" :key="t.id" :label="`${t.name}（${t.teacherNo}）`" :value="t.id"/></el-select></template></el-table-column></el-table>
 </el-card><el-dialog v-model="dialog" :title="reset?'重置教师密码':'开通教师账号'" width="min(450px,95vw)"><p>{{selected?.name}}（{{selected?.teacherNo}}）</p><el-alert title="初始密码：123456，首次登录必须修改" type="warning" :closable="false"/><template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="savePassword">确认</el-button></template></el-dialog></div>

    </div>
  </div>
</template>
<script setup>
import {ref,onMounted} from 'vue'
import {ElMessage,ElMessageBox} from 'element-plus'
import request from '@/utils/request'
const teachers=ref([]),accounts=ref([]),classes=ref([]),advisors=ref([]),loading=ref(false),saving=ref(false),dialog=ref(false),selected=ref(null),reset=ref(false),password=ref('')
const api=async(url,method='get',data,params)=>{const r=await request({url,method,data,params});if(r?.code!==0)throw Error(r?.msg||'操作失败');return r.data}
const load=async()=>{loading.value=true;try{const [t,a,c,v]=await Promise.all([api('/uni/teachers','get',undefined,{page:1,pageSize:100}),api('/uni/teacher-accounts'),api('/uni/org/classes','get',undefined,{page:1,pageSize:100}),api('/uni/class-advisors')]);teachers.value=t?.list||[];accounts.value=a?.list||[];classes.value=c?.list||[];advisors.value=v?.list||[]}catch(e){ElMessage.error(e.message||'加载失败')}finally{loading.value=false}}
const account=id=>accounts.value.find(a=>a.teacherId===id)
const advisor=id=>advisors.value.find(a=>a.classId===id)?.teacherId||''
const passwordDialog=(row,isReset)=>{selected.value=row;reset.value=isReset;password.value='';dialog.value=true}
const savePassword=async()=>{saving.value=true;try{await api(reset.value?'/uni/teacher-accounts/reset-password':'/uni/teacher-accounts','post',{teacherId:selected.value.id,password:"123456"});ElMessage.success('操作成功');dialog.value=false;password.value='';await load()}catch(e){ElMessage.error(e.message||'操作失败')}finally{saving.value=false}}
const toggle=async row=>{const a=account(row.id);try{await ElMessageBox.confirm(`确认${a.status===1?'停用':'启用'}该账号？`,'确认');await api('/uni/teacher-accounts/status','put',{teacherId:row.id,status:a.status===1?2:1});ElMessage.success('已更新');await load()}catch(e){if(e!=='cancel'&&e!=='close')ElMessage.error(e.message||'操作失败')}}
const setAdvisor=async(classId,teacherId)=>{try{await api('/uni/class-advisors','put',{classId,teacherId:teacherId||0});ElMessage.success('分配已更新');await load()}catch(e){ElMessage.error(e.message||'分配失败');await load()}}
onMounted(load)
</script>
<style scoped>.teacher-account{padding:20px}.teacher-account h3{margin:18px 0}.teacher-account :deep(.el-card){border-radius:12px}@media(max-width:600px){.teacher-account{padding:10px}}</style>

<style scoped>

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
