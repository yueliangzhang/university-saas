<template>
  <div class="uni-teacher-panel">
    <div v-if="mode === 'accounts'" class="uni-panel-section">
      <div class="uni-panel-heading"><div><h3>教师登录账号</h3><p>开通账号、重置密码及启停管理</p></div><el-button @click="load">刷新</el-button></div>
      <el-table v-loading="loading" :data="pageRows" row-key="id" style="width:100%" empty-text="暂无教师">
        <el-table-column prop="teacherNo" label="工号" min-width="120" />
        <el-table-column prop="name" label="姓名" min-width="140" />
        <el-table-column label="账号状态" min-width="130"><template #default="{row}"><el-tag v-if="account(row.id)" :type="account(row.id).status===1?'success':'info'">{{account(row.id).status===1?'正常':'停用'}}</el-tag><el-tag v-else type="warning">未开通</el-tag></template></el-table-column>
        <el-table-column label="操作" min-width="230"><template #default="{row}"><el-button v-if="!account(row.id)" type="primary" link @click="passwordDialog(row,false)">开通账号</el-button><template v-else><el-button type="primary" link @click="passwordDialog(row,true)">重置密码</el-button><el-button type="warning" link @click="toggle(row)">{{account(row.id).status===1?'停用':'启用'}}</el-button></template></template></el-table-column>
      </el-table>
    </div>
    <div v-else class="uni-panel-section">
      <div class="uni-panel-heading"><div><h3>班级辅导员分配</h3><p>班级默认辅导员仅用于新增学生时自动分配；修改此处不会改变已有学生的辅导员或审批权限。</p></div><el-button @click="load">刷新</el-button></div>
      <el-table v-loading="loading" :data="pageRows" row-key="id" style="width:100%" empty-text="暂无班级">
        <el-table-column prop="name" label="班级" min-width="160" />
        <el-table-column prop="code" label="班级代码" min-width="130" />
        <el-table-column label="默认辅导员" min-width="280"><template #default="{row}"><el-select :model-value="advisor(row.id)" clearable placeholder="未分配" style="width:min(260px,100%)" @change="v=>setAdvisor(row.id,v)"><el-option v-for="t in activeTeachers" :key="t.id" :label="`${t.name}（${t.teacherNo}）`" :value="t.id" /></el-select></template></el-table-column>
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
const props=defineProps({mode:{type:String,default:'accounts'}})
const accountPagination = reactive({ page: 1, pageSize: 10 })
const advisorPagination = reactive({ page: 1, pageSize: 10 })
const pagination = computed(() => props.mode === 'accounts' ? accountPagination : advisorPagination)
const rows = computed(() => props.mode === 'accounts' ? teachers.value : classes.value)
const total = computed(() => rows.value.length)
const pageRows = computed(() => {
  const { page, pageSize } = pagination.value
  return rows.value.slice((page - 1) * pageSize, page * pageSize)
})
const activeTeachers = computed(() => teachers.value.filter(teacher => teacher.status === 1))
watch([teachers, classes], ([teacherRows, classRows]) => {
  accountPagination.page = Math.min(accountPagination.page, Math.max(1, Math.ceil(teacherRows.length / accountPagination.pageSize)))
  advisorPagination.page = Math.min(advisorPagination.page, Math.max(1, Math.ceil(classRows.length / advisorPagination.pageSize)))
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
    const [t, a, c, v] = await Promise.all([
      loadTeachers(), api('/uni/teacher-accounts'), api('/uni/org/classes'), api('/uni/class-advisors')
    ])
    if (version !== loadVersion) return
    teachers.value = t
    accounts.value = a?.list || []
    classes.value = c?.list || []
    advisors.value = v?.list || []
  } catch (e) {
    if (version === loadVersion) ElMessage.error(e.message || '加载失败')
  } finally {
    if (version === loadVersion) loading.value = false
  }
}
const account=id=>accounts.value.find(a=>a.teacherId===id)
const advisor=id=>advisors.value.find(a=>a.classId===id)?.teacherId||''
const passwordDialog=(row,isReset)=>{selected.value=row;reset.value=isReset;password.value='';dialog.value=true}
const savePassword=async()=>{saving.value=true;try{await api(reset.value?'/uni/teacher-accounts/reset-password':'/uni/teacher-accounts','post',{teacherId:selected.value.id,password:"123456"});ElMessage.success('操作成功');dialog.value=false;password.value='';await load()}catch(e){ElMessage.error(e.message||'操作失败')}finally{saving.value=false}}
const toggle=async row=>{const a=account(row.id);try{await ElMessageBox.confirm(`确认${a.status===1?'停用':'启用'}该账号？`,'确认');await api('/uni/teacher-accounts/status','put',{teacherId:row.id,status:a.status===1?2:1});ElMessage.success('已更新');await load()}catch(e){if(e!=='cancel'&&e!=='close')ElMessage.error(e.message||'操作失败')}}
const setAdvisor=async(classId,teacherId)=>{try{await api('/uni/class-advisors','put',{classId,teacherId:teacherId||0});ElMessage.success('分配已更新');await load()}catch(e){ElMessage.error(e.message||'分配失败');await load()}}
onMounted(load)
watch(()=>props.mode,()=>load())
</script>
<style scoped>
.uni-panel-pagination{justify-content:flex-end;flex-wrap:wrap;row-gap:8px;padding:16px 0}
.uni-teacher-panel{width:100%;min-width:0}.uni-panel-section{background:#fff;border:1px solid #e7edf5;border-radius:14px;padding:20px}.uni-panel-heading{display:flex;justify-content:space-between;align-items:center;gap:12px;margin-bottom:16px}.uni-panel-heading h3{margin:0;font-size:16px;color:#1e293b}.uni-panel-heading p{margin:5px 0 0;color:#64748b;font-size:12px}@media(max-width:640px){.uni-panel-section{padding:12px}.uni-panel-heading{flex-wrap:wrap}}
</style>
