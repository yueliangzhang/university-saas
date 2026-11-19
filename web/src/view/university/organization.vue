<template>
  <!-- UNI_ADMIN_V17_6_2 -->
  <div class="uni-admin-v1762-page">
    <header class="uni-admin-v1762-heading">
      <div class="uni-admin-v1762-heading-copy">
        <span class="uni-admin-v1762-eyebrow">UNIVERSITY MANAGEMENT</span>
        <h1>高校组织架构</h1>
        <p>学院、专业与班级的统一管理</p>
      </div>
      <span class="uni-admin-v1762-heading-mark" aria-hidden="true">校务管理</span>
    </header>
    <div class="uni-admin-v1762-content">

  <div class="p-6">
    <el-card>
      <template #header>高校组织架构</template>
      <el-tabs v-model="tab" @tab-change="load">
        <el-tab-pane label="学院" name="colleges" />
        <el-tab-pane label="专业" name="majors" />
        <el-tab-pane label="班级" name="classes" />
      </el-tabs>
      <el-alert v-if="error" :title="error" type="error" :closable="false" class="mb-3" />
      <el-button type="primary" @click="openCreate">新增{{ labels[tab] }}</el-button>
      <el-table v-loading="loading" :data="rows" border class="mt-4">
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column prop="code" label="编码" />
        <el-table-column prop="name" label="名称" />
        <el-table-column v-if="tab==='majors'" prop="collegeId" label="学院 ID" />
        <el-table-column v-if="tab==='classes'" prop="majorId" label="专业 ID" />
        <el-table-column v-if="tab==='classes'" prop="enrollmentYear" label="入学年份" />
        <el-table-column label="状态" width="100"><template #default="{row}">{{ row.status===1?'启用':'停用' }}</template></el-table-column>
        <el-table-column label="操作" width="170"><template #default="{row}">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template></el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="dialog" :title="(editing?'编辑':'新增')+labels[tab]" width="520px">
      <el-form label-width="100px">
        <el-form-item label="编码"><el-input v-model="form.code" maxlength="64" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" maxlength="191" /></el-form-item>
        <el-form-item v-if="tab==='majors'" label="所属学院">
          <el-select v-model="form.collegeId" placeholder="选择学院" style="width:100%">
            <el-option v-for="x in colleges" :key="x.id" :label="`${x.name} (${x.code})`" :value="x.id" :disabled="x.status!==1" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="tab==='classes'" label="所属专业">
          <el-select v-model="form.majorId" placeholder="选择专业" style="width:100%">
            <el-option v-for="x in majors" :key="x.id" :label="`${x.name} (${x.code})`" :value="x.id" :disabled="x.status!==1" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="tab==='classes'" label="入学年份"><el-input-number v-model="form.enrollmentYear" :min="1900" :max="2200" /></el-form-item>
        <el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio :value="1">启用</el-radio><el-radio :value="2">停用</el-radio></el-radio-group></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </div>

    </div>
  </div>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
const labels={colleges:'学院',majors:'专业',classes:'班级'}
const tab=ref('colleges'),rows=ref([]),colleges=ref([]),majors=ref([])
const loading=ref(false),saving=ref(false),error=ref(''),dialog=ref(false),editing=ref(false)
const form=ref({})
const api=(kind,method='get',data,params)=>request({url:`/uni/org/${kind}`,method,data,params})
const fetchList=async kind=>{const r=await api(kind);if(r?.code!==0)throw new Error(r?.msg||'加载失败');return r.data?.list||[]}
const load=async()=>{loading.value=true;error.value='';try{rows.value=await fetchList(tab.value)}catch(e){error.value=e.message}finally{loading.value=false}}
const openCreate=async()=>{editing.value=false;form.value={code:'',name:'',status:1,collegeId:null,majorId:null,enrollmentYear:new Date().getFullYear()};await openDialog()}
const openEdit=async row=>{editing.value=true;form.value={...row};await openDialog()}
const openDialog=async()=>{try{if(tab.value==='majors')colleges.value=await fetchList('colleges');if(tab.value==='classes')majors.value=await fetchList('majors');dialog.value=true}catch(e){ElMessage.error(e.message)}}
const save=async()=>{if(!form.value.code?.trim()||!form.value.name?.trim()){ElMessage.warning('请填写编码和名称');return}
 saving.value=true;try{const r=await api(tab.value,editing.value?'put':'post',form.value);if(r?.code!==0)throw new Error(r?.msg||'保存失败');ElMessage.success('保存成功');dialog.value=false;await load()}catch(e){ElMessage.error(e.message)}finally{saving.value=false}}
const remove=async row=>{try{await ElMessageBox.confirm(`确定删除「${row.name}」？关联的下级记录必须先删除。`,'确认删除',{type:'warning'});const r=await api(tab.value,'delete',undefined,{id:row.id});if(r?.code!==0)throw new Error(r?.msg||'删除失败');ElMessage.success('删除成功');await load()}catch(e){if(e!=='cancel')ElMessage.error(e.message||'操作失败')}}
onMounted(load)
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
