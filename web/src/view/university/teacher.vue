<template>
  <main class="uni-record-page">
    <header class="uni-record-heading">
      <div>
        <h1>教师管理</h1>
        <p>统一管理教师档案、登录账号与班级辅导员</p>
      </div>
      <div v-if="uniTeacherTab === 'profiles'" style="display:flex;gap:8px;flex-wrap:wrap"><el-button type="primary" @click="openCreate">+ 新增教师</el-button><RecordColumns v-model="visibleColumns" :columns="optionalColumns" /><ExcelImport kind="teachers" :filters="{keyword,collegeId:filterCollege||undefined}" @imported="search" /><el-button type="danger" plain :loading="deleting" :disabled="loading || !selectedRows.length" @click="removeSelected">批量删除</el-button></div>
    </header>

    <nav class="uni-record-tabs" aria-label="教师管理功能">
      <button type="button" :class="{ active: uniTeacherTab === 'profiles' }" @click="uniTeacherTab = 'profiles'">教师档案</button>
      <button type="button" :class="{ active: uniTeacherTab === 'accounts' }" @click="uniTeacherTab = 'accounts'">登录账号</button>
      <button type="button" :class="{ active: uniTeacherTab === 'advisors' }" @click="uniTeacherTab = 'advisors'">辅导员分配</button>
    </nav>
    <section v-show="uniTeacherTab === 'profiles'" class="uni-record-panel">
      <div class="uni-record-panel-title">
        <div class="uni-record-panel-name">教师档案 <span class="uni-record-count">共 {{ total }} 名教师</span></div>
      </div>
      <el-form class="uni-record-filter" inline @submit.prevent="search">
        <el-form-item label="姓名 / 工号">
          <el-input v-model="keyword" clearable placeholder="搜索教师姓名或工号" style="width:220px" @keyup.enter="search" />
        </el-form-item>
        <el-form-item label="所属学院">
          <el-select v-model="filterCollege" clearable placeholder="全部学院" style="width:200px" @change="search">
            <el-option v-for="c in colleges" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="uni-record-table-scroll">
        <el-table :max-height="tableHeight" v-loading="loading" :data="rows" stripe style="width:100%" empty-text="暂无教师档案" @selection-change="selectedRows = $event" @sort-change="changeSort">
          <el-table-column type="selection" width="48" />
          <el-table-column prop="teacherNo" label="教师工号" min-width="125" sortable="custom">
            <template #default="{ row }"><span class="uni-record-number">{{ row.teacherNo }}</span></template>
          </el-table-column>
          <el-table-column prop="name" label="姓名" min-width="110" sortable="custom">
            <template #default="{ row }"><span class="uni-record-name">{{ row.name }}</span></template>
          </el-table-column>
          <el-table-column v-if="visibleColumns.includes('collegeName')" label="所属学院" min-width="150" sortable="custom" prop="collegeName">
            <template #default="{ row }">{{ collegeName(row.collegeId) }}</template>
          </el-table-column>
          <el-table-column v-if="visibleColumns.includes('title')" prop="title" label="职称" min-width="100" sortable="custom" />
          <el-table-column v-if="visibleColumns.includes('phone')" prop="phone" label="电话" min-width="130" sortable="custom" />
          <el-table-column v-if="visibleColumns.includes('email')" prop="email" label="邮箱" min-width="160" show-overflow-tooltip sortable="custom" />
          <el-table-column v-if="visibleColumns.includes('status')" label="状态" width="100" sortable="custom" prop="status">
            <template #default="{ row }"><el-tag size="small" :type="row.status===1?'success':'info'" effect="light">{{ row.status===1?'在职':'停用' }}</el-tag></template>
          </el-table-column>
          <el-table-column label="操作" width="176" fixed="right" align="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="detail = row">详情</el-button>
              <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
              <el-button link type="danger" @click="remove(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <div class="uni-record-pagination">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total"
          :page-sizes="[10,20,50,100]" layout="total, sizes, prev, pager, next"
          @current-change="load" @size-change="search" />
      </div>
    </section>
    <el-drawer :model-value="!!detail" title="教师详情" size="min(480px, 100vw)" append-to-body @update:model-value="value => { if (!value) detail = null }">
      <el-descriptions v-if="detail" :column="1" border>
        <el-descriptions-item label="教师工号">{{ detail.teacherNo }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ detail.name }}</el-descriptions-item>
        <el-descriptions-item label="所属学院">{{ collegeName(detail.collegeId) }}</el-descriptions-item>
        <el-descriptions-item label="职称">{{ detail.title || '—' }}</el-descriptions-item>
        <el-descriptions-item label="负责学生">{{ detail.studentCount ?? 0 }}</el-descriptions-item>
        <el-descriptions-item label="负责班级">{{ detail.classCount ?? 0 }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ detail.phone || '—' }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ detail.email || '—' }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ detail.status === 1 ? '在职' : '停用' }}</el-descriptions-item>
      </el-descriptions>
    </el-drawer>
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
    <TeacherAccountPanel v-if="uniTeacherTab!=='profiles'" :mode="uniTeacherTab" :colleges="colleges" />
  </main>
</template>
<script setup>
import { useSessionStorage, StorageSerializers, useWindowSize } from '@vueuse/core'
import RecordColumns from './RecordColumns.vue'
import ExcelImport from './ExcelImport.vue' 
import TeacherAccountPanel from './teacherAccountPanel.vue'
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import { useRecordTable } from './useRecordTable'
const uniTeacherTab = ref('profiles')
const detail = ref(null)
const { height: viewportHeight } = useWindowSize()
const tableHeight = computed(() => Math.min(720, Math.max(260, viewportHeight.value * 0.65)))
const optionalColumns = { collegeName: '所属学院', title: '职称', phone: '电话', email: '邮箱', status: '状态' }
const visibleColumns = useSessionStorage('university-teacher-columns', Object.keys(optionalColumns))
const rows=ref([]),colleges=ref([]),keyword=useSessionStorage('university-teacher-keyword', ''),filterCollege=useSessionStorage('university-teacher-college', null, { serializer: StorageSerializers.object }),page=ref(1),pageSize=ref(20),total=ref(0)
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
<style scoped src="./recordPage.css"></style>
