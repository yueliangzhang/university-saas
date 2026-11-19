<template>
  <!-- UNI_ADMIN_V17_6_2 -->
  <div class="uni-admin-v1762-page">
    <header class="uni-admin-v1762-heading">
      <div class="uni-admin-v1762-heading-copy">
        <span class="uni-admin-v1762-eyebrow">UNIVERSITY MANAGEMENT</span>
        <h1>学生管理</h1>
        <p>学生档案、学籍信息与状态管理</p>
      </div>
      <span class="uni-admin-v1762-heading-mark" aria-hidden="true">校务管理</span>
    </header>
    <div class="uni-admin-v1762-content">

  <div class="p-6">
    <el-card>
      <template #header>学生管理</template>
      <el-form inline>
        <el-form-item label="姓名/学号"><el-input v-model="keyword" clearable placeholder="搜索姓名或学号" @keyup.enter="search" /></el-form-item>
        <el-form-item label="班级"><el-select v-model="filterClass" clearable placeholder="全部班级" style="width:180px"><el-option v-for="x in classes" :key="x.id" :label="x.name" :value="x.id" /></el-select></el-form-item>
        <el-form-item><el-button type="primary" @click="search">查询</el-button><el-button @click="reset">重置</el-button><el-button type="success" @click="openCreate">新增学生</el-button></el-form-item>
      </el-form>
      <el-table v-loading="loading" :data="rows" border>
        <el-table-column prop="studentNo" label="学号" min-width="135" />
        <el-table-column prop="name" label="姓名" min-width="110" />
        <el-table-column label="性别" width="80"><template #default="{row}">{{ ['未填写','男','女'][row.gender] || '未知' }}</template></el-table-column>
        <el-table-column label="班级" min-width="150"><template #default="{row}">{{ className(row.classId) }}</template></el-table-column>
        <el-table-column prop="enrollmentYear" label="入学年份" width="110" />
        <el-table-column prop="phone" label="电话" min-width="120" />
        <el-table-column label="状态" width="85"><template #default="{row}">{{ row.status===1?'在读':'停用' }}</template></el-table-column>
        <el-table-column label="操作" width="150"><template #default="{row}"><el-button link type="primary" @click="openEdit(row)">编辑</el-button><el-button link type="danger" @click="remove(row)">删除</el-button></template></el-table-column>
      </el-table>
      <div class="mt-4 flex justify-end"><el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10,20,50,100]" layout="total, sizes, prev, pager, next" @current-change="load" @size-change="search" /></div>
    </el-card>
    <el-dialog v-model="dialog" :title="editing?'编辑学生':'新增学生'" width="560px">
      <el-form :model="form" label-width="95px">
        <el-form-item label="学号" required><el-input v-model="form.studentNo" maxlength="64" /></el-form-item>
        <el-form-item label="姓名" required><el-input v-model="form.name" maxlength="191" /></el-form-item>
        <el-form-item label="性别"><el-select v-model="form.gender" style="width:100%"><el-option label="未填写" :value="0" /><el-option label="男" :value="1" /><el-option label="女" :value="2" /></el-select></el-form-item>
        <el-form-item label="班级" required><el-select v-model="form.classId" filterable placeholder="选择班级" style="width:100%"><el-option v-for="x in activeClasses" :key="x.id" :label="x.name" :value="x.id" /></el-select></el-form-item>
        <el-form-item label="入学年份" required><el-input-number v-model="form.enrollmentYear" :min="1900" :max="2200" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" maxlength="32" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" maxlength="191" /></el-form-item>
        <el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio :value="1">在读</el-radio><el-radio :value="2">停用</el-radio></el-radio-group></el-form-item>
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
const rows=ref([]),classes=ref([]),majors=ref([]),colleges=ref([])
const keyword=ref(''),filterClass=ref(null),page=ref(1),pageSize=ref(20),total=ref(0)
const loading=ref(false),saving=ref(false),dialog=ref(false),editing=ref(false),form=ref({})
const activeClasses=computed(()=>classes.value.filter(x=>x.status===1 && majors.value.some(m=>m.id===x.majorId && m.status===1 && colleges.value.some(c=>c.id===m.collegeId && c.status===1))))
const className=id=>classes.value.find(x=>x.id===id)?.name||`班级 #${id}`
const api=(method='get',data,params)=>request({url:'/uni/students',method,data,params})
const load=async()=>{loading.value=true;try{const r=await api('get',undefined,{page:page.value,pageSize:pageSize.value,keyword:keyword.value,classId:filterClass.value||undefined});if(r?.code!==0)throw Error(r?.msg||'加载失败');rows.value=r.data?.list||[];total.value=r.data?.total||0}catch(e){ElMessage.error(e.message)}finally{loading.value=false}}
const loadOptions=async()=>{try{for(const [kind,target] of [['colleges',colleges],['majors',majors],['classes',classes]]){const r=await request({url:`/uni/org/${kind}`,method:'get'});if(r?.code!==0)throw Error(r?.msg||'组织架构加载失败');target.value=r.data?.list||[]}}catch(e){ElMessage.error(e.message)}}
const search=()=>{page.value=1;load()}
const reset=()=>{keyword.value='';filterClass.value=null;search()}
const openCreate=()=>{editing.value=false;form.value={studentNo:'',name:'',gender:0,classId:null,enrollmentYear:new Date().getFullYear(),phone:'',email:'',status:1};dialog.value=true}
const openEdit=row=>{editing.value=true;form.value={...row};dialog.value=true}
const save=async()=>{if(!form.value.studentNo?.trim()||!form.value.name?.trim()||!form.value.classId){ElMessage.warning('请填写学号、姓名并选择班级');return}saving.value=true;try{const r=await api(editing.value?'put':'post',form.value);if(r?.code!==0)throw Error(r?.msg||'保存失败');ElMessage.success('保存成功');dialog.value=false;await load()}catch(e){ElMessage.error(e.message)}finally{saving.value=false}}
const remove=async row=>{try{await ElMessageBox.confirm(`确定删除学生「${row.name}」？`,'删除确认',{type:'warning'});const r=await api('delete',undefined,{id:row.id});if(r?.code!==0)throw Error(r?.msg||'删除失败');ElMessage.success('删除成功');await load()}catch(e){if(e!=='cancel')ElMessage.error(e.message||'删除失败')}}
onMounted(async()=>{await loadOptions();await load()})
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
