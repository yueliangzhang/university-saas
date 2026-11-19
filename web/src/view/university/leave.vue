<template>
  <!-- UNI_ADMIN_V17_6_2 -->
  <div class="uni-admin-v1762-page">
    <header class="uni-admin-v1762-heading">
      <div class="uni-admin-v1762-heading-copy">
        <span class="uni-admin-v1762-eyebrow">UNIVERSITY MANAGEMENT</span>
        <h1>学生请假管理</h1>
        <p>请假申请、审批与记录查询</p>
      </div>
      <span class="uni-admin-v1762-heading-mark" aria-hidden="true">校务管理</span>
    </header>
    <div class="uni-admin-v1762-content">

  <div class="p-6">
    <el-card>
      <template #header>学生请假管理</template>
      <el-form inline>
        <el-form-item label="姓名/学号"><el-input v-model="keyword" clearable placeholder="搜索学生" @keyup.enter="search" /></el-form-item>
        <el-form-item label="状态"><el-select v-model="status" clearable placeholder="全部" style="width:140px"><el-option v-for="(name,key) in statusNames" :key="key" :label="name" :value="key" /></el-select></el-form-item>
        <el-form-item><el-button type="primary" @click="search">查询</el-button><el-button @click="reset">重置</el-button><el-button type="success" @click="openCreate">新增请假</el-button></el-form-item>
      </el-form>
      <el-table v-loading="loading" :data="rows" border>
        <el-table-column label="学生" min-width="150"><template #default="{row}">{{ studentName(row.studentId) }}</template></el-table-column>
        <el-table-column label="类型" width="95"><template #default="{row}">{{ typeNames[row.leaveType]||row.leaveType }}</template></el-table-column>
        <el-table-column prop="reason" label="请假原因" min-width="200" show-overflow-tooltip />
        <el-table-column prop="startAt" label="开始时间" min-width="175"><template #default="{row}">{{ fmt(row.startAt) }}</template></el-table-column>
        <el-table-column prop="endAt" label="结束时间" min-width="175"><template #default="{row}">{{ fmt(row.endAt) }}</template></el-table-column>
        <el-table-column label="状态" width="95"><template #default="{row}">{{ statusNames[row.status]||row.status }}</template></el-table-column>
        <el-table-column prop="reviewNote" label="审批备注" min-width="130" show-overflow-tooltip />
        <el-table-column label="操作" width="235" fixed="right"><template #default="{row}"><template v-if="row.status==='pending'"><el-button link type="primary" @click="openEdit(row)">编辑</el-button><el-button link type="success" @click="review(row,'approved')">通过</el-button><el-button link type="danger" @click="review(row,'rejected')">驳回</el-button><el-button link @click="cancel(row)">撤销</el-button></template><span v-else>—</span></template></el-table-column>
      </el-table>
      <div class="mt-4 flex justify-end"><el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10,20,50,100]" layout="total, sizes, prev, pager, next" @current-change="load" @size-change="search" /></div>
    </el-card>
    <el-dialog v-model="dialog" :title="editing?'编辑请假':'新增请假'" width="600px">
      <el-form :model="form" label-width="95px">
        <el-form-item label="学生" required><el-select v-model="form.studentId" filterable placeholder="选择在读学生" style="width:100%"><el-option v-for="s in activeStudents" :key="s.id" :label="`${s.name} (${s.studentNo})`" :value="s.id" /></el-select></el-form-item>
        <el-form-item label="请假类型" required><el-select v-model="form.leaveType" style="width:100%"><el-option v-for="(name,key) in typeNames" :key="key" :label="name" :value="key" /></el-select></el-form-item>
        <el-form-item label="开始时间" required><el-date-picker v-model="form.startAt" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width:100%" /></el-form-item>
        <el-form-item label="结束时间" required><el-date-picker v-model="form.endAt" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width:100%" /></el-form-item>
        <el-form-item label="请假原因" required><el-input v-model="form.reason" type="textarea" :rows="3" maxlength="2000" show-word-limit /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </div>

    </div>
  </div>
</template>
<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
const typeNames={sick:'病假',personal:'事假',other:'其他'}
const statusNames={pending:'待审批',approved:'已通过',rejected:'已驳回',cancelled:'已撤销'}
const rows=ref([]),students=ref([]),keyword=ref(''),status=ref(''),page=ref(1),pageSize=ref(20),total=ref(0)
const loading=ref(false),saving=ref(false),dialog=ref(false),editing=ref(false),form=ref({})
const activeStudents=computed(()=>students.value.filter(x=>x.status===1))
const studentName=id=>{const s=students.value.find(x=>x.id===id);return s?`${s.name} (${s.studentNo})`:`学生 #${id}`}
const fmt=v=>v?String(v).replace('T',' ').slice(0,19):''
const api=(method='get',data,params,url='/uni/leaves')=>request({url,method,data,params})
const load=async()=>{loading.value=true;try{const r=await api('get',undefined,{page:page.value,pageSize:pageSize.value,keyword:keyword.value,status:status.value||undefined});if(r?.code!==0)throw Error(r?.msg||'加载失败');rows.value=r.data?.list||[];total.value=r.data?.total||0}catch(e){ElMessage.error(e.message||'加载失败')}finally{loading.value=false}}
const loadStudents=async()=>{try{let p=1,all=[];while(p<=100){const r=await request({url:'/uni/students',method:'get',params:{page:p,pageSize:100}});if(r?.code!==0)throw Error(r?.msg||'学生列表加载失败');const chunk=r.data?.list||[];all.push(...chunk);if(all.length>=(r.data?.total||0)||chunk.length===0)break;p++}students.value=all}catch(e){ElMessage.error(e.message||'学生列表加载失败')}}
const search=()=>{page.value=1;load()}
const reset=()=>{keyword.value='';status.value='';search()}
const openCreate=()=>{editing.value=false;form.value={studentId:null,leaveType:'personal',reason:'',startAt:'',endAt:''};dialog.value=true}
const openEdit=row=>{editing.value=true;form.value={id:row.id,studentId:row.studentId,leaveType:row.leaveType,reason:row.reason,startAt:fmt(row.startAt),endAt:fmt(row.endAt)};dialog.value=true}
const save=async()=>{const f=form.value;if(!f.studentId||!f.reason?.trim()||!f.startAt||!f.endAt||f.endAt<=f.startAt){ElMessage.warning('请填写完整信息，结束时间必须晚于开始时间');return}saving.value=true;try{const r=await api(editing.value?'put':'post',f);if(r?.code!==0)throw Error(r?.msg||'保存失败');ElMessage.success('保存成功');dialog.value=false;await load()}catch(e){ElMessage.error(e.message||'保存失败')}finally{saving.value=false}}
const review=async(row,action)=>{try{let note='';if(action==='rejected'){const result=await ElMessageBox.prompt('请输入驳回原因','驳回请假',{inputValidator:v=>!!v?.trim()&&v.length<=1000||'请填写1000字以内的原因'});note=result.value}else{await ElMessageBox.confirm('确认通过该请假申请？','审批确认')}const r=await api('post',{id:row.id,action,note},undefined,'/uni/leaves/review');if(r?.code!==0)throw Error(r?.msg||'审批失败');ElMessage.success('审批完成');await load()}catch(e){if(e!=='cancel'&&e!=='close')ElMessage.error(e.message||'审批失败')}}
const cancel=async row=>{try{await ElMessageBox.confirm('确认撤销该待审批申请？','撤销确认');const r=await api('post',{id:row.id},undefined,'/uni/leaves/cancel');if(r?.code!==0)throw Error(r?.msg||'撤销失败');ElMessage.success('已撤销');await load()}catch(e){if(e!=='cancel'&&e!=='close')ElMessage.error(e.message||'撤销失败')}}
onMounted(async()=>{await loadStudents();await load()})
</script>

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
