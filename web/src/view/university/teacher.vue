<template>
  <!-- UNI_TEACHER_MERGE_V17_6_8 -->
  <div class="uni-merged-teacher">
    <header class="uni-merged-heading"><div><h2>教师管理</h2><p>统一管理教师档案、登录账号与班级辅导员</p></div></header>
    <el-tabs v-model="uniTeacherTab" class="uni-merged-tabs">
      <el-tab-pane label="教师档案" name="profiles" />
      <el-tab-pane label="登录账号" name="accounts" />
      <el-tab-pane label="辅导员分配" name="advisors" />
    </el-tabs>
    <div v-show="uniTeacherTab==='profiles'" class="uni-merged-profile">

  <div class="uni-teacher-v1767">
    <header class="teacher-heading">
      <div>
        <h1>教师管理</h1>
        <p>管理教师档案、学院归属与在职状态</p>
      </div>
      <div style="display:flex;gap:8px;flex-wrap:wrap"><el-button type="primary" @click="openCreate">新增教师</el-button><ExcelImport kind="teachers" :filters="{keyword,collegeId:filterCollege||undefined}" @imported="search" /><el-button type="danger" plain :loading="deleting" :disabled="loading || !selectedRows.length" @click="removeSelected">批量删除</el-button></div>
    </header>

    <section class="teacher-panel">
      <div class="teacher-panel-top">
        <div class="teacher-panel-title">教师档案 <span>共 {{ total }} 名教师</span></div>
      </div>
      <el-form class="teacher-filters" inline @submit.prevent="search">
        <el-form-item label="姓名 / 工号">
          <el-input v-model="keyword" clearable placeholder="搜索教师姓名或工号" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="所属学院">
          <el-select v-model="filterCollege" clearable placeholder="全部学院" style="width:190px" @change="search">
            <el-option v-for="c in colleges" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item class="teacher-filter-actions">
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="teacher-table-scroll">
        <el-table v-loading="loading" :data="rows" style="width:100%" empty-text="暂无教师档案" @selection-change="selectedRows = $event" @sort-change="changeSort">
          <el-table-column type="selection" width="48" />
          <el-table-column prop="teacherNo" label="教师工号" min-width="125" sortable="custom" />
          <el-table-column prop="name" label="姓名" min-width="110" sortable="custom">
            <template #default="{ row }"><span class="teacher-name">{{ row.name }}</span></template>
          </el-table-column>
          <el-table-column label="所属学院" min-width="150" sortable="custom" prop="collegeName">
            <template #default="{ row }">{{ collegeName(row.collegeId) }}</template>
          </el-table-column>
          <el-table-column prop="title" label="职称" min-width="100" sortable="custom" />
          <el-table-column prop="phone" label="电话" min-width="130" sortable="custom" />
          <el-table-column prop="email" label="邮箱" min-width="160" show-overflow-tooltip sortable="custom" />
          <el-table-column label="状态" width="100" sortable="custom" prop="status">
            <template #default="{ row }"><el-tag size="small" :type="row.status===1?'success':'info'" effect="light">{{ row.status===1?'在职':'停用' }}</el-tag></template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
              <el-button link type="danger" @click="remove(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <div class="teacher-pager">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total"
          :page-sizes="[10,20,50,100]" layout="total, sizes, prev, pager, next"
          @current-change="load" @size-change="search" />
      </div>
    </section>
    <el-dialog v-model="dialog" :title="editing?'编辑教师':'新增教师'" width="min(560px,95vw)" append-to-body>
      <el-form label-width="100px" :model="form">
        <el-form-item label="教师工号" required><el-input v-model="form.teacherNo" maxlength="64" /></el-form-item>
        <el-form-item label="姓名" required><el-input v-model="form.name" maxlength="191" /></el-form-item>
        <el-form-item label="所属学院" required><el-select v-model="form.collegeId" filterable placeholder="请选择学院" style="width:100%"><el-option v-for="c in activeColleges" :key="c.id" :label="c.name" :value="c.id" /></el-select></el-form-item>
        <el-form-item label="职称"><el-input v-model="form.title" maxlength="100" placeholder="如：讲师、副教授" /></el-form-item>
        <el-form-item label="性别"><el-select v-model="form.gender" style="width:100%"><el-option label="未填写" :value="0" /><el-option label="男" :value="1" /><el-option label="女" :value="2" /></el-select></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" maxlength="32" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" maxlength="191" /></el-form-item>
        <el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio :value="1">在职</el-radio><el-radio :value="2">停用</el-radio></el-radio-group></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </div>

    </div>
    <TeacherAccountPanel v-if="uniTeacherTab!=='profiles'" :mode="uniTeacherTab" />
  </div>
</template>
<script setup>
import ExcelImport from './ExcelImport.vue' 
import TeacherAccountPanel from './teacherAccountPanel.vue'
import { ref as uniMergeRef } from 'vue'
const uniTeacherTab = uniMergeRef('profiles')

import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import { useRecordTable } from './useRecordTable'
const rows=ref([]),colleges=ref([]),keyword=ref(''),filterCollege=ref(null),page=ref(1),pageSize=ref(20),total=ref(0)
const loading=ref(false),saving=ref(false),dialog=ref(false),editing=ref(false),form=ref({})
const activeColleges=computed(()=>colleges.value.filter(c=>c.status===1))
const collegeName=id=>colleges.value.find(c=>c.id===id)?.name||`学院 #${id}`
const api=(method='get',data,params)=>request({url:'/uni/teachers',method,data,params})
let loadVersion = 0
const load = async () => {
  const version = ++loadVersion
  loading.value = true
  try {
    const result = await api('get', undefined, {
      page: page.value, pageSize: pageSize.value, keyword: keyword.value,
      collegeId: filterCollege.value || undefined,
      ...sort.value
    })
    if (version !== loadVersion) return
    if (result?.code !== 0) throw new Error(result?.msg || '加载失败')
    rows.value = result.data?.list || []
    total.value = result.data?.total || 0
    const lastPage = Math.max(1, Math.ceil(total.value / pageSize.value))
    if (page.value > lastPage) page.value = lastPage
  } catch (error) {
    if (version === loadVersion) ElMessage.error(error.message || '加载失败')
  } finally {
    if (version === loadVersion) loading.value = false
  }
}
const { selectedRows, deleting, removeSelected, sort, setSort } = useRecordTable(
  ids => api('delete', undefined, { ids }),
  load
)
const changeSort = event => {
  setSort(event)
  search()
}
const loadColleges=async()=>{try{const r=await request({url:'/uni/org/colleges',method:'get'});if(r?.code!==0)throw Error(r?.msg||'学院加载失败');colleges.value=r.data?.list||[]}catch(e){ElMessage.error(e.message||'学院加载失败')}}
const search=()=>{page.value=1;load()}
const reset=()=>{keyword.value='';filterCollege.value=null;search()}
const openCreate=()=>{editing.value=false;form.value={teacherNo:'',name:'',gender:0,phone:'',email:'',collegeId:null,title:'',status:1};dialog.value=true}
const openEdit=row=>{editing.value=true;form.value={...row};dialog.value=true}
const save=async()=>{const f=form.value;if(!f.teacherNo?.trim()||!f.name?.trim()||!f.collegeId){ElMessage.warning('请填写工号、姓名和学院');return}saving.value=true;try{const r=await api(editing.value?'put':'post',f);if(r?.code!==0)throw Error(r?.msg||'保存失败');ElMessage.success('保存成功');dialog.value=false;await load()}catch(e){ElMessage.error(e.message||'保存失败')}finally{saving.value=false}}
const remove=async row=>{try{await ElMessageBox.confirm(`确定删除教师「${row.name}」？`,'删除确认',{type:'warning'});const r=await api('delete',undefined,{id:row.id});if(r?.code!==0)throw Error(r?.msg||'删除失败');ElMessage.success('已删除');await load()}catch(e){if(e!=='cancel'&&e!=='close')ElMessage.error(e.message||'删除失败')}}
onMounted(async()=>{await loadColleges();await load()})
</script>
<style scoped>
.uni-teacher-v1767 { width:100%; min-width:0; max-width:none; padding:20px 24px 28px; color:#243247; box-sizing:border-box; }
.teacher-heading { display:flex; align-items:center; justify-content:space-between; gap:16px; margin-bottom:18px; }
.teacher-heading h1 { margin:0; font-size:22px; font-weight:650; line-height:1.4; }
.teacher-heading p { margin:5px 0 0; color:#738198; font-size:13px; }
.teacher-panel { min-width:0; padding:20px; background:#fff; border:1px solid #e5eaf2; border-radius:14px; box-shadow:0 3px 16px rgba(30,52,87,.035); }
.teacher-panel-top { display:flex; justify-content:space-between; align-items:center; padding-bottom:16px; }
.teacher-panel-title { font-size:15px; font-weight:650; }
.teacher-panel-title span { margin-left:12px; color:#8a96a9; font-size:12px; font-weight:400; }
.teacher-filters { display:flex; flex-wrap:wrap; align-items:center; gap:4px 12px; margin-bottom:12px; }
.teacher-filters :deep(.el-form-item) { margin-right:0; margin-bottom:8px; }
.teacher-filters :deep(.el-input) { width:220px; }
.teacher-table-scroll { width:100%; overflow-x:auto; }
.teacher-table-scroll :deep(.el-table) { min-width:850px; --el-table-header-bg-color:#f7f9fc; --el-table-border-color:#edf0f5; --el-table-row-hover-bg-color:#f7faff; }
.teacher-table-scroll :deep(.el-table th.el-table__cell) { height:44px; color:#65748b; font-weight:600; }
.teacher-table-scroll :deep(.el-table td.el-table__cell) { height:48px; }
.teacher-name { font-weight:600; color:#27364d; }
.teacher-pager { display:flex; justify-content:flex-end; padding-top:16px; overflow-x:auto; }
.uni-teacher-v1767 :deep(.el-button--primary) { --el-button-bg-color:#2563eb; --el-button-border-color:#2563eb; }
.uni-teacher-v1767 :deep(.el-input__wrapper), .uni-teacher-v1767 :deep(.el-select__wrapper) { border-radius:8px; }
@media (max-width:768px) {
  .uni-teacher-v1767 { padding:12px; }
  .teacher-heading { align-items:flex-start; }
  .teacher-heading h1 { font-size:19px; }
  .teacher-panel { padding:14px; }
  .teacher-filters { display:block; }
  .teacher-filters :deep(.el-form-item) { display:flex; margin-bottom:12px; }
  .teacher-filters :deep(.el-form-item__content) { flex:1; min-width:0; }
  .teacher-filters :deep(.el-input), .teacher-filters :deep(.el-select) { width:100% !important; }
  .teacher-pager { justify-content:flex-start; }
}
</style>

<style scoped>
.uni-merged-teacher{padding:18px 20px 26px;width:100%;min-width:0;box-sizing:border-box}.uni-merged-heading{display:flex;justify-content:space-between;align-items:center;margin-bottom:10px}.uni-merged-heading h2{margin:0;font-size:21px;font-weight:650;color:#1e293b}.uni-merged-heading p{margin:5px 0 0;color:#64748b;font-size:12px}.uni-merged-tabs{margin-bottom:14px}.uni-merged-profile{min-width:0}.uni-merged-profile :deep(.hero){display:none!important}.uni-merged-profile :deep(.teacher-page){padding:0!important;display:block!important;max-width:none!important;width:100%!important}.uni-merged-profile :deep(.content){width:100%!important;max-width:none!important}.uni-merged-profile :deep(.content .el-card__body){padding:20px!important}@media(max-width:640px){.uni-merged-teacher{padding:12px}.uni-merged-profile :deep(.content .el-card__body){padding:12px!important}}
</style>
