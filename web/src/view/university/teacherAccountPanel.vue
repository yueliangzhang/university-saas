<template>
  <div class="uni-teacher-panel">
    <div class="uni-panel-section">
      <div v-if="mode === 'accounts'" class="uni-panel-heading"><div><h3>教师登录账号</h3><p>开通账号、重置密码及启停管理</p></div><div><el-button type="warning" plain :loading="batchResetting" :disabled="loading || !selectedAccounts.length" @click="resetSelected">批量重置密码</el-button><el-button @click="load">刷新</el-button></div></div>
      <div v-else class="uni-panel-heading"><div><h3>班级辅导员分配</h3><p>班级默认辅导员仅用于新增学生时自动分配；修改此处不会改变已有学生的辅导员或审批权限。</p></div><el-button @click="load">刷新</el-button></div>
      <el-form class="uni-panel-filters" inline @submit.prevent="search">
        <el-form-item :label="mode === 'accounts' ? '姓名 / 工号' : '班级 / 辅导员'">
          <el-input v-model="filters.keyword" clearable
            :placeholder="mode === 'accounts' ? '搜索教师姓名或工号' : '搜索班级、辅导员姓名或工号'" />
        </el-form-item>
        <el-form-item label="所属学院">
          <el-select v-model="filters.collegeId" clearable filterable placeholder="全部学院">
            <el-option v-for="college in colleges" :key="college.id" :label="college.name" :value="college.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
      <el-table v-if="mode === 'accounts'" ref="accountTable" v-loading="loading" :data="pageRows" row-key="id" style="width:100%" empty-text="暂无符合条件的教师" @selection-change="selectedAccounts = $event">
        <el-table-column type="selection" width="48" :selectable="row => row.status === 1 && !!account(row.id)" />
        <el-table-column prop="teacherNo" label="工号" min-width="120" />
        <el-table-column prop="name" label="姓名" min-width="140" />
        <el-table-column label="账号状态" min-width="130"><template #default="{row}"><el-tag v-if="account(row.id)" :type="account(row.id).status===1?'success':'info'">{{account(row.id).status===1?'正常':'停用'}}</el-tag><el-tag v-else type="warning">未开通</el-tag></template></el-table-column>
        <el-table-column label="操作" min-width="230"><template #default="{row}"><el-button v-if="!account(row.id)" type="primary" link @click="passwordDialog(row,false)">开通账号</el-button><template v-else><el-button type="primary" link @click="passwordDialog(row,true)">重置密码</el-button><el-button type="warning" link @click="toggle(row)">{{account(row.id).status===1?'停用':'启用'}}</el-button></template></template></el-table-column>
      </el-table>
      <el-table v-else v-loading="loading" :data="pageRows" row-key="id" style="width:100%" empty-text="暂无符合条件的班级">
        <el-table-column prop="name" label="班级" min-width="160" />
        <el-table-column prop="code" label="班级代码" min-width="130" />
        <el-table-column label="默认辅导员" min-width="280"><template #default="{row}"><el-select :model-value="advisor(row.id)" clearable filterable placeholder="未分配" style="width:min(260px,100%)" @change="v=>setAdvisor(row.id,v)"><el-option v-for="t in activeTeachers" :key="t.id" :label="`${t.name}（${t.teacherNo}）`" :value="t.id" /></el-select></template></el-table-column>
      </el-table>
    </div>
    <el-pagination
      v-if="total"
      v-model:current-page="pagination.page"
      v-model:page-size="pagination.pageSize"
      :page-sizes="[10, 20, 50, 100]"
      :pager-count="5"
      :total="total"
      layout="total, sizes, prev, pager, next"
      class="uni-panel-pagination"
      @size-change="pagination.page = 1"
    />
    <el-dialog v-model="dialog" :title="reset?'重置教师密码':'开通教师账号'" width="min(450px,95vw)" append-to-body><p>{{selected?.name}}（{{selected?.teacherNo}}）</p><el-alert title="初始密码：123456，首次登录必须修改" type="warning" :closable="false"/><template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" :loading="saving" @click="savePassword">确认</el-button></template></el-dialog>
  </div>
</template>
<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {ElMessage,ElMessageBox} from 'element-plus'
import request from '@/utils/request'
const teachers=ref([]),accounts=ref([]),classes=ref([]),advisors=ref([]),loading=ref(false),saving=ref(false),dialog=ref(false),selected=ref(null),reset=ref(false),password=ref('')
const props = defineProps({
  mode: { type: String, default: 'accounts' },
  colleges: { type: Array, default: () => [] }
})
const majors = ref([])
const accountTable = ref(null)
const selectedAccounts = ref([])
const batchResetting = ref(false)
const resetSelected = async () => {
  const teacherIds = selectedAccounts.value.map(row => row.id)
  if (!teacherIds.length || batchResetting.value) return
  batchResetting.value = true
  try {
    await ElMessageBox.confirm(
      `将 ${teacherIds.length} 个账号重置为初始密码，并退出这些账号的当前会话？`,
      '批量重置教师密码', { type: 'warning' }
    )
    const result = await request({ url: '/uni/teacher-accounts/reset-password', method: 'post', data: { teacherIds } })
    if (result.code !== 0) throw new Error(result.msg || '重置失败')
    ElMessage.success(`已重置 ${result.data.count} 个账号`)
    await load()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error(error.response?.data?.msg || error.message || '重置失败')
  } finally {
    batchResetting.value = false
  }
}
const accountPagination = reactive({ page: 1, pageSize: 10 })
const advisorPagination = reactive({ page: 1, pageSize: 10 })
const pagination = computed(() => props.mode === 'accounts' ? accountPagination : advisorPagination)
const accountFilters = reactive({ keyword: '', collegeId: null, appliedKeyword: '', appliedCollegeId: null })
const advisorFilters = reactive({ keyword: '', collegeId: null, appliedKeyword: '', appliedCollegeId: null })
const filters = computed(() => props.mode === 'accounts' ? accountFilters : advisorFilters)
const teacherById = computed(() => new Map(teachers.value.map(teacher => [teacher.id, teacher])))
const collegeByMajor = computed(() => new Map(majors.value.map(major => [major.id, major.collegeId])))
const advisorByClass = computed(() => new Map(advisors.value.map(item => [item.classId, item.teacherId])))
const rows = computed(() => {
  const { appliedKeyword: keyword, appliedCollegeId: collegeId } = filters.value
  const matches = (...values) => !keyword || values.some(value => String(value ?? '').toLowerCase().includes(keyword))
  if (props.mode === 'accounts') {
    return teachers.value.filter(teacher =>
      (!collegeId || teacher.collegeId === collegeId) && matches(teacher.name, teacher.teacherNo)
    )
  }
  return classes.value.filter(row => {
    if (collegeId && collegeByMajor.value.get(row.majorId) !== collegeId) return false
    const teacher = teacherById.value.get(advisorByClass.value.get(row.id))
    return matches(row.name, row.code, teacher?.name, teacher?.teacherNo)
  })
})

function search() {
  filters.value.appliedKeyword = filters.value.keyword.trim().toLowerCase()
  filters.value.appliedCollegeId = filters.value.collegeId
  pagination.value.page = 1
  accountTable.value?.clearSelection()
}

function resetSearch() {
  filters.value.keyword = ''
  filters.value.collegeId = null
  search()
}
const total = computed(() => rows.value.length)
const pageRows = computed(() => {
  const { page, pageSize } = pagination.value
  return rows.value.slice((page - 1) * pageSize, page * pageSize)
})
const activeTeachers = computed(() => teachers.value.filter(teacher => teacher.status === 1))
watch(total, count => {
  const state = pagination.value
  state.page = Math.min(state.page, Math.max(1, Math.ceil(count / state.pageSize)))
})
const api=async(url,method='get',data,params)=>{const r=await request({url,method,data,params});if(r?.code!==0)throw Error(r?.msg||'操作失败');return r.data}
const loadTeachers = async () => {
  const pageSize = 100
  const first = await api('/uni/teachers', 'get', undefined, { page: 1, pageSize })
  const list = [...(first?.list || [])]
  for (let page = 2; page <= Math.ceil(first?.total / pageSize); page++) {
    const result = await api('/uni/teachers', 'get', undefined, { page, pageSize })
    list.push(...(result?.list || []))
  }
  return list
}
let loadVersion = 0
const load = async () => {
  const version = ++loadVersion
  loading.value = true
  try {
    const [t, a, c, v, m] = await Promise.all([
      loadTeachers(), api('/uni/teacher-accounts'), api('/uni/org/classes'), api('/uni/class-advisors'), api('/uni/org/majors')
    ])
    if (version !== loadVersion) return
    teachers.value = t
    accounts.value = a?.list || []
    classes.value = c?.list || []
    advisors.value = v?.list || []
    majors.value = m?.list || []
  } catch (e) {
    if (version === loadVersion) ElMessage.error(e.message || '加载失败')
  } finally {
    if (version === loadVersion) loading.value = false
  }
}
const account=id=>accounts.value.find(a=>a.teacherId===id)
const advisor = id => advisorByClass.value.get(id) || ''
const passwordDialog=(row,isReset)=>{selected.value=row;reset.value=isReset;password.value='';dialog.value=true}
const savePassword=async()=>{saving.value=true;try{await api(reset.value?'/uni/teacher-accounts/reset-password':'/uni/teacher-accounts','post',{teacherId:selected.value.id,password:"123456"});ElMessage.success('操作成功');dialog.value=false;password.value='';await load()}catch(e){ElMessage.error(e.message||'操作失败')}finally{saving.value=false}}
const toggle=async row=>{const a=account(row.id);try{await ElMessageBox.confirm(`确认${a.status===1?'停用':'启用'}该账号？`,'确认');await api('/uni/teacher-accounts/status','put',{teacherId:row.id,status:a.status===1?2:1});ElMessage.success('已更新');await load()}catch(e){if(e!=='cancel'&&e!=='close')ElMessage.error(e.message||'操作失败')}}
const setAdvisor=async(classId,teacherId)=>{try{await api('/uni/class-advisors','put',{classId,teacherId:teacherId||0});ElMessage.success('分配已更新');await load()}catch(e){ElMessage.error(e.message||'分配失败');await load()}}
onMounted(load)
watch(() => props.mode, () => {
  selectedAccounts.value = []
  load()
})
</script>
<style scoped>
.uni-panel-filters { display: flex; flex-wrap: wrap; align-items: center; gap: 4px 12px; margin-bottom: 12px; }
.uni-panel-filters :deep(.el-form-item) { margin-right: 0; margin-bottom: 8px; }
.uni-panel-filters :deep(.el-input) { width: 220px; }
.uni-panel-filters :deep(.el-select) { width: 190px; }
@media (max-width: 640px) {
  .uni-panel-filters { display: block; }
  .uni-panel-filters :deep(.el-form-item) { display: flex; }
  .uni-panel-filters :deep(.el-form-item__content) { flex: 1; min-width: 0; }
  .uni-panel-filters :deep(.el-input), .uni-panel-filters :deep(.el-select) { width: 100%; }
}

.uni-panel-pagination{justify-content:flex-end;flex-wrap:wrap;row-gap:8px;padding:16px 0}
.uni-teacher-panel{width:100%;min-width:0}.uni-panel-section{background:#fff;border:1px solid #e7edf5;border-radius:14px;padding:20px}.uni-panel-heading{display:flex;justify-content:space-between;align-items:center;gap:12px;margin-bottom:16px}.uni-panel-heading h3{margin:0;font-size:16px;color:#1e293b}.uni-panel-heading p{margin:5px 0 0;color:#64748b;font-size:12px}@media(max-width:640px){.uni-panel-section{padding:12px}.uni-panel-heading{flex-wrap:wrap}}
</style>
