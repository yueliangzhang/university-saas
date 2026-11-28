<template>
  <!-- UNI_STUDENT_V17_6_9 -->
  <main class="uni-student-page">
    <div class="uni-student-heading">
      <div>
        <h1>学生管理</h1>
        <p>统一管理学生档案、请假审批与登录账号</p>
      </div>
      <div v-if="activeTab === 'profile'" style="display:flex;gap:8px;flex-wrap:wrap"><el-button v-if="activeTab === 'profile'" type="primary" @click="openCreate">+ 新增学生</el-button><RecordColumns v-model="visibleColumns" :columns="optionalColumns" /><ExcelImport kind="students" :filters="{keyword, classId:filterClass||undefined, advisorId:filterAdvisor??undefined, status:filterStatus||undefined, enrollmentYear:filterYear||undefined}" @imported="search" /><el-button type="danger" plain :loading="deleting" :disabled="loading || !selectedRows.length" @click="removeSelected">批量删除</el-button></div>
    </div>

    <nav class="uni-student-tabs" aria-label="学生管理功能">
      <button type="button" :class="{ active: activeTab === 'profile' }" @click="activeTab = 'profile'">学生档案</button>
      <button type="button" :class="{ active: activeTab === 'leave' }" @click="activeTab = 'leave'">请假审批</button>
      <button type="button" :class="{ active: activeTab === 'account' }" @click="activeTab = 'account'">登录账号</button>
    </nav>
    <section v-show="activeTab === 'profile'" class="uni-student-panel">
      <div class="uni-student-panel-title">
        <div class="uni-student-panel-name">学生档案 <span class="uni-student-count">共 {{ total }} 人</span></div>
      </div>
      <el-form class="uni-student-filter" inline @submit.prevent="search">
        <el-form-item label="姓名 / 学号">
          <el-input v-model="keyword" clearable placeholder="搜索姓名或学号" style="width:220px" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="所属班级">
          <el-select v-model="filterClass" clearable filterable placeholder="全部班级" style="width:200px" @change="search">
            <el-option v-for="x in classes" :key="x.id" :label="x.name" :value="x.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="辅导员"><el-select v-model="filterAdvisor" clearable filterable placeholder="全部辅导员" style="width:190px" @visible-change="visible => { if (visible) loadTeachers() }" @change="search"><el-option label="未分配" :value="0"/><el-option v-for="t in teachers" :key="t.id" :label="`${t.name}（${t.teacherNo}）`" :value="t.id"/></el-select></el-form-item>
        <el-form-item label="学籍状态">
          <el-select v-model="filterStatus" clearable placeholder="全部状态" style="width:130px" @change="search">
            <el-option label="在读" value="1" /><el-option label="停用" value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="入学年份">
          <el-input-number v-model="filterYear" :min="1900" :max="2100" :controls="false" placeholder="全部年份" @change="search" />
        </el-form-item>
        <el-form-item class="uni-student-filter-actions">
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="uni-student-table-scroll">
        <el-table :max-height="tableHeight" v-loading="loading" :data="rows" stripe style="width:100%" empty-text="暂无符合条件的学生" @selection-change="selectedRows = $event" @sort-change="changeSort">
          <el-table-column type="selection" width="48" />
          <el-table-column prop="studentNo" label="学号" min-width="130" sortable="custom">
            <template #default="{row}"><span class="uni-student-number">{{ row.studentNo }}</span></template>
          </el-table-column>
          <el-table-column prop="name" label="姓名" min-width="150" sortable="custom">
            <template #default="{row}"><span class="uni-student-name">{{ row.name }}</span></template>
          </el-table-column>
          <el-table-column v-if="visibleColumns.includes('gender')" label="性别" width="85" sortable="custom" prop="gender">
            <template #default="{row}">{{ ['未填写','男','女'][row.gender] || '未知' }}</template>
          </el-table-column>
          <el-table-column v-if="visibleColumns.includes('className')" label="所属班级" min-width="180" show-overflow-tooltip sortable="custom" prop="className">
            <template #default="{row}">{{ className(row.classId) }}</template>
          </el-table-column>
          <el-table-column v-if="visibleColumns.includes('advisorName')" label="所属辅导员" min-width="135" sortable="custom" prop="advisorName"><template #default="{row}">{{row.advisorName || '未分配'}}</template></el-table-column>
          <el-table-column v-if="visibleColumns.includes('enrollmentYear')" prop="enrollmentYear" label="入学年份" width="110" sortable="custom" />
          <el-table-column v-if="visibleColumns.includes('phone')" prop="phone" label="联系电话" min-width="140" show-overflow-tooltip sortable="custom">
            <template #default="{row}">{{ row.phone || '—' }}</template>
          </el-table-column>
          <el-table-column v-if="visibleColumns.includes('status')" label="学籍状态" width="110" sortable="custom" prop="status">
            <template #default="{row}"><el-tag :type="row.status === 1 ? 'success' : 'info'" effect="light" size="small">{{ row.status === 1 ? '在读' : '停用' }}</el-tag></template>
          </el-table-column>
          <el-table-column label="操作" width="176" fixed="right" align="right">
            <template #default="{row}">
              <el-button link type="primary" @click="detail = row">详情</el-button>
              <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
              <el-button link type="danger" @click="remove(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <div class="uni-student-pagination">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10,20,50,100]" layout="total, sizes, prev, pager, next" @current-change="load" @size-change="search" />
      </div>
    </section>

    <section v-if="activeTab === 'leave'" class="uni-student-panel uni-student-subpanel"><StudentLeavePanel /></section>
    <section v-if="activeTab === 'account'" class="uni-student-panel uni-student-subpanel"><StudentAccountPanel /></section>

    <el-drawer :model-value="!!detail" title="学生详情" size="min(480px, 100vw)" append-to-body @update:model-value="value => { if (!value) detail = null }">
      <el-descriptions v-if="detail" :column="1" border>
        <el-descriptions-item label="学号">{{ detail.studentNo }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ detail.name }}</el-descriptions-item>
        <el-descriptions-item label="所属班级">{{ className(detail.classId) }}</el-descriptions-item>
        <el-descriptions-item label="所属辅导员">{{ detail.advisorName || '未分配' }}</el-descriptions-item>
        <el-descriptions-item label="入学年份">{{ detail.enrollmentYear || '—' }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ detail.phone || '—' }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ detail.email || '—' }}</el-descriptions-item>
        <el-descriptions-item label="学籍状态">{{ detail.status === 1 ? '在读' : '停用' }}</el-descriptions-item>
      </el-descriptions>
    </el-drawer>
    <el-dialog v-model="dialog" :title="editing ? '编辑学生' : '新增学生'" width="560px" append-to-body class="uni-student-dialog">
      <el-form :model="form" label-width="95px">
        <el-form-item label="学号" required><el-input v-model="form.studentNo" maxlength="64" /></el-form-item>
        <el-form-item label="姓名" required><el-input v-model="form.name" maxlength="191" /></el-form-item>
        <el-form-item label="性别"><el-select v-model="form.gender" style="width:100%"><el-option label="未填写" :value="0" /><el-option label="男" :value="1" /><el-option label="女" :value="2" /></el-select></el-form-item>
        <el-form-item label="班级" required><el-select v-model="form.classId" filterable placeholder="选择班级" style="width:100%"><el-option v-for="x in activeClasses" :key="x.id" :label="x.name" :value="x.id" /></el-select></el-form-item>
        <el-form-item label="辅导员"><el-select v-model="advisorChoice" filterable style="width:100%"><el-option v-if="!editing" label="按班级自动分配" value="auto"/><el-option label="暂不分配" value="none"/><el-option v-for="t in teachers" :key="t.id" :label="`${t.name}（${t.teacherNo}）${t.status!==1?' · 已停用':''}`" :value="t.id" :disabled="t.status!==1"/></el-select></el-form-item>
        <el-form-item label="入学年份" required><el-input-number v-model="form.enrollmentYear" :min="1900" :max="2200" /></el-form-item>
        <el-form-item label="电话"><el-input v-model="form.phone" maxlength="32" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" maxlength="191" /></el-form-item>
        <el-form-item label="状态"><el-radio-group v-model="form.status"><el-radio :value="1">在读</el-radio><el-radio :value="2">停用</el-radio></el-radio-group></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
  </main>
</template>
<script setup>
import { useSessionStorage, StorageSerializers, useWindowSize } from '@vueuse/core'
import RecordColumns from './RecordColumns.vue'
import ExcelImport from './ExcelImport.vue' 
import { ref, computed, onMounted } from 'vue'
import StudentLeavePanel from './studentLeavePanel.vue'
import StudentAccountPanel from './studentAccountPanel.vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import { useRecordTable } from './useRecordTable'
const activeTab=ref('profile')
const teachers=ref([]),filterAdvisor=useSessionStorage('university-student-advisor', null, { serializer: StorageSerializers.object }),advisorChoice=ref('auto')
let teachersRequest
const loadTeachers = () => {
  if (teachersRequest) return teachersRequest
  teachersRequest = (async () => {
    try {
      const all = []
      for (let page = 1; ; page++) {
        const result = await request({
          url: '/uni/teachers', method: 'get', params: { page, pageSize: 100 }, donNotShowLoading: true
        })
        if (result?.code !== 0) throw new Error(result?.msg || '教师加载失败')
        const chunk = result.data?.list || []
        all.push(...chunk)
        if (!chunk.length || all.length >= result.data.total) break
      }
      teachers.value = all
    } catch (error) {
      teachersRequest = null
      ElMessage.error(error.message || '教师加载失败')
    }
  })()
  return teachersRequest
}
const rows=ref([]),classes=ref([]),majors=ref([]),colleges=ref([])
const detail = ref(null)
const { height: viewportHeight } = useWindowSize()
const tableHeight = computed(() => Math.min(720, Math.max(260, viewportHeight.value * 0.65)))
const optionalColumns = { gender: '性别', className: '所属班级', advisorName: '所属辅导员', enrollmentYear: '入学年份', phone: '联系电话', status: '学籍状态' }
const visibleColumns = useSessionStorage('university-student-columns', Object.keys(optionalColumns))
const keyword=useSessionStorage('university-student-keyword', ''),filterClass=useSessionStorage('university-student-class', null, { serializer: StorageSerializers.object }),page=ref(1),pageSize=ref(20),total=ref(0)
const loading=ref(false),saving=ref(false),dialog=ref(false),editing=ref(false),form=ref({})
const filterStatus = useSessionStorage('university-student-status', '')
const filterYear = useSessionStorage('university-student-year', null, { serializer: StorageSerializers.object })
const activeClasses=computed(()=>classes.value.filter(x=>x.status===1 && majors.value.some(m=>m.id===x.majorId && m.status===1 && colleges.value.some(c=>c.id===m.collegeId && c.status===1))))
const className=id=>classes.value.find(x=>x.id===id)?.name||`班级 #${id}`
const api=(method='get',data,params)=>request({url:'/uni/students',method,data,params})
let loadVersion = 0
const load = async () => {
  const version = ++loadVersion
  loading.value = true
  try {
    const result = await api('get', undefined, {
      page: page.value, pageSize: pageSize.value, keyword: keyword.value,
      classId: filterClass.value || undefined, advisorId: filterAdvisor.value ?? undefined, status: filterStatus.value || undefined, enrollmentYear: filterYear.value || undefined,
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
const loadOptions = () => Promise.all(
  [['colleges', colleges], ['majors', majors], ['classes', classes]].map(async ([kind, target]) => {
    try {
      const result = await request({ url: `/uni/org/${kind}`, method: 'get', donNotShowLoading: true })
      if (result?.code !== 0) throw new Error(result?.msg || '组织架构加载失败')
      target.value = result.data?.list || []
    } catch (error) {
      ElMessage.error(error.message || '组织架构加载失败')
    }
  })
)
const search=()=>{page.value=1;load()}
const reset=()=>{keyword.value='';filterClass.value=null;filterAdvisor.value=null;filterStatus.value='';filterYear.value=null;search()}
const openCreate=()=>{editing.value=false;advisorChoice.value='auto';loadTeachers();form.value={studentNo:'',name:'',gender:0,classId:null,enrollmentYear:new Date().getFullYear(),phone:'',email:'',status:1};dialog.value=true}
const openEdit=row=>{editing.value=true;advisorChoice.value=row.advisorId ?? 'none';loadTeachers();form.value={...row};dialog.value=true}
const save=async()=>{if(!form.value.studentNo?.trim()||!form.value.name?.trim()||!form.value.classId){ElMessage.warning('请填写学号、姓名并选择班级');return}saving.value=true;try{const payload={...form.value};delete payload.advisorId;if(advisorChoice.value!=='auto'&&(!editing.value||advisorChoice.value!==(form.value.advisorId??'none')))payload.advisorId=advisorChoice.value==='none'?null:advisorChoice.value;const r=await api(editing.value?'put':'post',payload);if(r?.code!==0)throw Error(r?.msg||'保存失败');ElMessage.success('保存成功');dialog.value=false;await load()}catch(e){ElMessage.error(e.message)}finally{saving.value=false}}
const remove=async row=>{try{await ElMessageBox.confirm(`确定删除学生「${row.name}」？`,'删除确认',{type:'warning'});const r=await api('delete',undefined,{id:row.id});if(r?.code!==0)throw Error(r?.msg||'删除失败');ElMessage.success('删除成功');await load()}catch(e){if(e!=='cancel')ElMessage.error(e.message||'删除失败')}}
onMounted(() => {
  load()
  loadOptions()
  if (filterAdvisor.value) loadTeachers()
})
</script>
<style scoped>
.uni-student-tabs { display:flex; align-items:center; gap:5px; border-bottom:1px solid #e7edf5; margin:0 0 16px; padding:0 2px; overflow-x:auto; }
.uni-student-tabs button { appearance:none; cursor:pointer; white-space:nowrap; background:transparent; border:0; border-bottom:2px solid transparent; padding:12px 20px; font-size:14px; color:#65748b; font-weight:550; }
.uni-student-tabs button:hover { color:#2563eb; }
.uni-student-tabs button.active { color:#2563eb; border-bottom-color:#2563eb; font-weight:650; }
.uni-student-subpanel { overflow:visible; }
@media(max-width:768px) { .uni-student-tabs button { padding:11px 14px; } }

.uni-student-page { padding: 18px 22px 28px; min-width: 0; color: #25344d; }
.uni-student-heading { display:flex; align-items:center; justify-content:space-between; gap:16px; margin-bottom:18px; }
.uni-student-heading h1 { margin:0; font-size:22px; line-height:1.4; font-weight:650; color:#1d2b43; }
.uni-student-heading p { margin:5px 0 0; color:#7a879a; font-size:13px; }
.uni-student-panel { border:1px solid #e6ebf3; background:#fff; border-radius:12px; overflow:hidden; box-shadow:0 3px 14px rgba(27,47,84,.025); }
.uni-student-panel-title { padding:17px 20px; border-bottom:1px solid #edf0f5; }
.uni-student-panel-name { font-size:15px; font-weight:650; }
.uni-student-count { margin-left:10px; font-size:12px; font-weight:400; color:#8190a4; }
.uni-student-filter { display:flex; align-items:center; flex-wrap:wrap; gap:0 14px; padding:18px 20px 2px; }
.uni-student-filter :deep(.el-form-item) { margin-right:0; margin-bottom:16px; }
.uni-student-filter :deep(.el-form-item__label) { color:#56657b; font-weight:500; }
.uni-student-filter :deep(.el-input__wrapper), .uni-student-filter :deep(.el-select__wrapper) { border-radius:7px; }
.uni-student-filter-actions { margin-left:auto; }
.uni-student-page :deep(.el-button) { border-radius:7px; }
.uni-student-page :deep(.el-button--primary) { --el-button-bg-color:#2563eb; --el-button-border-color:#2563eb; --el-button-hover-bg-color:#1d4ed8; --el-button-hover-border-color:#1d4ed8; }
.uni-student-table-scroll { padding:0 20px; min-width:0; }
.uni-student-page :deep(.el-table) { --el-table-header-bg-color:#f7f9fc; --el-table-border-color:#edf1f6; --el-table-row-hover-bg-color:#f7faff; font-size:13px; }
.uni-student-page :deep(.el-table th.el-table__cell) { color:#617087; font-weight:600; height:46px; }
.uni-student-page :deep(.el-table td.el-table__cell) { height:48px; color:#334155; }
.uni-student-name { font-weight:600; color:#25344d; }
.uni-student-number { font-variant-numeric:tabular-nums; color:#475569; }
.uni-student-page :deep(.el-tag) { border-radius:5px; }
.uni-student-pagination { display:flex; justify-content:flex-end; padding:17px 20px 20px; }
@media (max-width: 768px) {
  .uni-student-page { padding:12px; }
  .uni-student-heading { align-items:flex-start; }
  .uni-student-heading h1 { font-size:19px; }
  .uni-student-panel-title { padding:14px; }
  .uni-student-filter { padding:14px 14px 0; display:block; }
  .uni-student-filter :deep(.el-form-item) { display:flex; margin-bottom:12px; }
  .uni-student-filter :deep(.el-form-item__content) { flex:1; min-width:0; }
  .uni-student-filter :deep(.el-input), .uni-student-filter :deep(.el-select) { width:100% !important; }
  .uni-student-filter-actions { margin-left:0; }
  .uni-student-table-scroll { padding:0 10px; }
  .uni-student-pagination { padding:14px 10px; overflow-x:auto; justify-content:flex-start; }
  .uni-student-pagination :deep(.el-pagination) { flex-wrap:nowrap; }
}
</style>
