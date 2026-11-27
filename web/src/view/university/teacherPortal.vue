<template>
  <div class="teacher-portal uni-ui">
    <div v-if="token" class="topbar">
      <div class="brand">{{ advisorMode ? '辅导员工作台' : '🎓 高校教师服务平台' }}</div>
      <div>
        <el-button v-if="advisorMode" plain @click="returnToWorkspace">{{ userStore.userInfo.authorityId === 2002 ? '切换为请假监管员' : '切换为高校管理员' }}</el-button>
        <el-button plain @click="logout">退出登录</el-button>
      </div>
    </div>
    <div v-if="token && mustChange" style="max-width:440px;margin:60px auto;padding:24px"><el-card><h2>首次登录，请修改密码</h2><el-form label-position="top" @submit.prevent="changePassword"><el-form-item label="当前密码"><el-input v-model="oldPassword" type="password"/></el-form-item><el-form-item label="新密码（6位）"><el-input v-model="newPassword" type="password"/></el-form-item><el-form-item label="确认新密码"><el-input v-model="confirmPassword" type="password"/></el-form-item><el-button native-type="submit" type="primary">确认修改</el-button><el-button @click="logout">退出</el-button></el-form></el-card></div>
<main v-else-if="token" class="main"><div class="welcome"><div><h2>你好，{{me?.name || '老师'}}</h2><p class="muted">工号：{{me?.teacherNo}} · {{me?.title || '教师'}}</p></div><el-button type="primary" :loading="refreshing" @click="refresh">刷新数据</el-button></div>
      <section class="teacher-metrics" aria-label="负责范围概览">
        <article v-for="metric in metrics" :key="metric.label">
          <span>{{ metric.label }}</span><strong>{{ metric.value }}</strong>
        </article>
      </section>
      <el-card shadow="never" class="section">
        <template #header><b>近 7 天已批准请假人数</b></template>
        <div v-if="trend.length" class="teacher-trend">
          <div v-for="day in trend" :key="day.date">
            <strong>{{ day.count }}</strong>
            <span class="trend-bar" :style="{ height: (day.count / trendMax * 64) + 'px' }"></span>
            <span>{{ day.date.slice(5) }}</span>
          </div>
        </div>
        <el-empty v-else description="暂无趋势数据" :image-size="48" />
      </el-card>
      <el-card shadow="never" class="section">
        <template #header><b>学生请假审批</b></template>
        <form class="teacher-filters" @submit.prevent="search">
          <el-input v-model="leaveKeyword" placeholder="学生姓名或学号" aria-label="学生姓名或学号" clearable />
          <el-select v-model="selectedClassId" aria-label="审批班级" placeholder="全部班级" clearable>
            <el-option v-for="cl in classes" :key="cl.id" :label="cl.name" :value="String(cl.id)" />
          </el-select>
          <el-select v-model="status" aria-label="审批状态">
            <el-option label="全部记录" value="" />
            <el-option v-for="(label, key) in statuses" :key="key" :label="label" :value="key" />
          </el-select>
          <label>开始日期 <input v-model="dateFrom" type="date" :max="dateTo || undefined" /></label>
          <label>结束日期 <input v-model="dateTo" type="date" :min="dateFrom || undefined" /></label>
          <el-button native-type="submit" type="primary">查询</el-button>
          <el-button :loading="exporting" @click="exportLeaves">导出 Excel</el-button>
        </form>
        <div class="batch-tools">
          <el-checkbox :model-value="allPendingSelected" :disabled="!pendingRows.length" @change="selectPending">全选本页待审批</el-checkbox>
          <el-button :disabled="!selectedIds.length || reviewing" @click="review(selectedIds, 'approved')">批量通过</el-button>
          <el-button :disabled="!selectedIds.length || reviewing" @click="review(selectedIds, 'rejected')">批量驳回</el-button>
        </div>
        <el-skeleton v-if="loading" :rows="4" animated />
        <el-alert v-else-if="leaveError" :title="leaveError" type="error" :closable="false"><el-button link @click="loadLeaves">重试</el-button></el-alert>
        <el-empty v-else-if="!rows.length" description="暂无符合条件的请假记录" :image-size="64" />
        <article v-for="row in rows" :key="row.id" class="approval-card">
          <el-checkbox v-if="row.status==='pending'" :model-value="selectedIds.includes(row.id)" :aria-label="'选择 ' + row.studentName + ' 的申请'" @change="checked => selectRow(row.id, checked)" />
          <button class="approval-overview" @click="selectedLeave = row">
            <strong>{{ row.studentName }}</strong>
            <span>{{ row.className }} · {{ types[row.leaveType] || row.leaveType }} · {{ duration(row) }}</span>
            <small>{{ fmt(row.startAt) }} 至 {{ fmt(row.endAt) }}</small>
          </button>
          <el-tag :type="colors[row.status] || 'info'">{{ statuses[row.status] }}</el-tag>
          <div v-if="row.status==='pending'" class="approval-actions">
            <el-button type="success" plain :disabled="reviewing" @click="review([row.id], 'approved')">通过</el-button>
            <el-button type="danger" plain :disabled="reviewing" @click="review([row.id], 'rejected')">驳回</el-button>
          </div>
        </article>
        <el-pagination class="pager" v-model:current-page="page" :page-size="10" :pager-count="5"
          :total="total" layout="total,prev,pager,next" @current-change="loadLeaves" />
      </el-card>
      <el-card shadow="never" class="section">
        <template #header><b>负责班级</b></template>
        <el-empty v-if="!classes.length" description="暂无负责班级" :image-size="64" />
        <div v-else class="class-cards">
          <article v-for="cl in classes" :key="cl.id">
            <strong>{{ cl.name }}</strong>
            <span>{{ cl.studentCount }} 名学生 · {{ cl.onLeave }} 人正在请假 · {{ leaveRate(cl) }}</span>
          </article>
        </div>
      </el-card>
      <el-card shadow="never" class="section">
        <template #header><b>我负责的学生</b></template>
        <form class="teacher-filters" @submit.prevent="searchStudents">
          <el-input v-model="studentKeyword" placeholder="姓名或学号" aria-label="姓名或学号" clearable />
          <el-select v-model="studentClass" clearable placeholder="全部班级" aria-label="学生班级">
            <el-option v-for="cl in classes" :key="cl.id" :label="cl.name" :value="cl.id" />
          </el-select>
          <el-button native-type="submit" type="primary">查询</el-button>
        </form>
        <el-table :data="studentRows" v-loading="studentLoading" empty-text="暂无符合条件的学生">
          <el-table-column prop="studentNo" label="学号" min-width="104" />
          <el-table-column prop="name" label="姓名" min-width="78" />
          <el-table-column prop="className" label="班级" min-width="100" />
          <el-table-column label="操作" width="70"><template #default="{ row }"><el-button link type="primary" @click="selectedStudent = row">详情</el-button></template></el-table-column>
        </el-table>
        <el-pagination class="pager" v-model:current-page="studentPage" :page-size="20" :pager-count="5"
          :total="studentTotal" layout="total,prev,pager,next" @current-change="loadStudents" />
      </el-card>
      <el-dialog :model-value="!!selectedLeave" title="请假详情" width="min(520px,94vw)" @update:model-value="value => { if (!value) selectedLeave = null }">
        <dl v-if="selectedLeave" class="teacher-detail">
          <dt>学生</dt><dd>{{ selectedLeave.studentName }}（{{ selectedLeave.studentNo }}）</dd>
          <dt>班级</dt><dd>{{ selectedLeave.className }}</dd>
          <dt>请假日期</dt><dd>{{ fmt(selectedLeave.startAt) }} 至 {{ fmt(selectedLeave.endAt) }}</dd>
          <dt>请假原因</dt><dd>{{ selectedLeave.reason }}</dd>
          <dt>审批状态</dt><dd>{{ statuses[selectedLeave.status] }}</dd>
          <dt>审批意见</dt><dd>{{ selectedLeave.reviewNote || '—' }}</dd>
          <dt v-if="selectedLeave.hasProof">证明文件</dt><dd v-if="selectedLeave.hasProof"><el-button v-for="index in (selectedLeave.proofCount || 1)" :key="index" link type="primary" @click="downloadProof(selectedLeave, index)">附件 {{ index }}</el-button></dd>
        </dl>
      </el-dialog>
      <el-dialog :model-value="!!selectedStudent" title="学生详情" width="min(480px,94vw)" @update:model-value="value => { if (!value) selectedStudent = null }">
        <dl v-if="selectedStudent" class="teacher-detail">
          <dt>姓名</dt><dd>{{ selectedStudent.name }}</dd><dt>学号</dt><dd>{{ selectedStudent.studentNo }}</dd>
          <dt>班级</dt><dd>{{ selectedStudent.className }}</dd><dt>入学年份</dt><dd>{{ selectedStudent.enrollmentYear || '—' }}</dd>
          <dt>状态</dt><dd>{{ selectedStudent.status === 1 ? '在读' : '停用' }}</dd>
        </dl>
      </el-dialog>
      <el-card v-if="!advisorMode" shadow="never" class="section"><template #header><b>账号安全</b></template><el-button @click="changeDialog=true">修改密码</el-button></el-card>
    </main>
    <el-dialog v-if="!advisorMode" v-model="changeDialog" title="修改登录密码" width="min(460px,94vw)"><el-form label-position="top"><el-form-item label="旧密码"><el-input v-model="oldPassword" type="password" show-password autocomplete="current-password"/></el-form-item><el-form-item label="新密码（6位）"><el-input v-model="newPassword" type="password" show-password autocomplete="new-password"/></el-form-item></el-form><template #footer><el-button @click="changeDialog=false">取消</el-button><el-button type="primary" :loading="busy" @click="changePassword">确认修改</el-button></template></el-dialog>
  </div>
</template>
<script setup>
import {computed,ref,onMounted,watch} from 'vue'
import { useUserStore } from '@/pinia/modules/user'
import request from '@/utils/request'
import {ElMessage,ElMessageBox} from 'element-plus'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { downloadLeaveProof } from '@/utils/leaveProof'
const base=(import.meta.env.VITE_BASE_API||'/api').replace(/\/$/,'')
const client=axios.create({baseURL:base,timeout:15000})
const mustChange = ref(false)
const confirmPassword = ref('')
const props = defineProps({ advisorMode: Boolean })
const userStore = useUserStore()
const teacherToken = ref(sessionStorage.getItem('uni_teacher_token') || '')
const token = computed(() => props.advisorMode ? userStore.token : teacherToken.value)
const portalBase = computed(() => props.advisorMode ? '/uni/advisor-workspace' : '/uni/teacher-portal')
const router = useRouter()
watch(token, value => {
  if (!value) router.replace({ name: 'Login' })
})
const me=ref(null),classes=ref([]),rows=ref([]),total=ref(0),page=ref(1),status=ref('pending'),selectedClassId=ref(''),loading=ref(false),busy=ref(false),refreshing=ref(false)
const studentRows = ref([])
const studentPage = ref(1)
const studentTotal = ref(0)
const studentClass = ref(null)
const studentKeyword = ref('')
const studentLoading = ref(false)
const studentFilters = ref({})
const selectedStudent = ref(null)
const selectedLeave = ref(null)
const selectedIds = ref([])
const reviewing = ref(false)
const exporting = ref(false)
const leaveKeyword = ref('')
const dateFrom = ref('')
const dateTo = ref('')
const leaveFilters = ref({ status: 'pending' })
const leaveError = ref('')
const summary = ref(null)
const trend = computed(() => summary.value?.trend || [])
const trendMax = computed(() => Math.max(1, ...trend.value.map(day => day.count)))
const metrics = computed(() => [
  { label: '负责学生', value: classes.value.reduce((sum, row) => sum + row.studentCount, 0) },
  { label: '负责班级', value: classes.value.length },
  { label: '待审批', value: summary.value?.pending ?? '—' },
  { label: '正在请假', value: classes.value.reduce((sum, row) => sum + row.onLeave, 0) }
])
const leaveRate = row => row.studentCount ? (row.onLeave * 100 / row.studentCount).toFixed(1) + '%' : '0%'
const duration = row => {
  const minutes = Math.round((new Date(row.endAt) - new Date(row.startAt)) / 60000)
  if (!(minutes > 0)) return '—'
  const days = Math.floor(minutes / 1440)
  const hours = Math.floor(minutes / 60) % 24
  const rest = minutes % 60
  return [days && days + '天', hours && hours + '小时', rest && rest + '分钟'].filter(Boolean).join('')
}
const pendingRows = computed(() => rows.value.filter(row => row.status === 'pending'))
const allPendingSelected = computed(() => pendingRows.value.length > 0 && pendingRows.value.every(row => selectedIds.value.includes(row.id)))
const selectPending = checked => { selectedIds.value = checked ? pendingRows.value.map(row => row.id) : [] }
const selectRow = (id, checked) => {
  selectedIds.value = checked ? [...selectedIds.value, id] : selectedIds.value.filter(item => item !== id)
}
let studentVersion = 0
let leaveVersion = 0
const loadStudents = async () => {
  const version = ++studentVersion
  studentLoading.value = true
  try {
    const data = await call('/students', 'get', undefined, { page: studentPage.value, pageSize: 20, ...studentFilters.value })
    if (version !== studentVersion) return
    studentRows.value = data.list || []
    studentTotal.value = data.total
  } catch (e) {
    if (version === studentVersion) ElMessage.error(error(e))
  } finally {
    if (version === studentVersion) studentLoading.value = false
  }
}
const searchStudents = () => {
  studentFilters.value = { classId: studentClass.value || undefined, keyword: studentKeyword.value.trim() }
  if (studentPage.value === 1) loadStudents()
  else studentPage.value = 1
}
const changeDialog=ref(false),oldPassword=ref(''),newPassword=ref('')
const types={personal:'事假',sick:'病假',other:'其他'},statuses={pending:'待审批',approved:'已通过',rejected:'已驳回',cancelled:'已撤销'},colors={pending:'warning',approved:'success',rejected:'danger',cancelled:'info'}
const fmt=s=>s?String(s).replace('T',' ').slice(0,16):''
const error=e=>e.response?.data?.msg||e.message||'操作失败'
const call = async (url, method = 'get', data, params) => {
  const config = { url: portalBase.value + url, method, data, params }
  if (props.advisorMode) {
    const result = await request(config)
    if (result.code !== 0) throw new Error(result.msg || '请求失败')
    return result.data
  }
  const result = await client({ ...config, headers: { Authorization: `Bearer ${token.value}` } })
  if (result.data?.code !== 0) throw new Error(result.data?.msg || '请求失败')
  return result.data.data
}
const clear = () => {
  teacherToken.value = ''
  sessionStorage.removeItem('uni_teacher_token')
}
const loadLeaves = async (includeSummary = false) => {
  const version = ++leaveVersion
  loading.value = true
  leaveError.value = ''
  rows.value = []
  selectedIds.value = []
  try {
    const data = await call('/leaves', 'get', undefined, {
      page: page.value, pageSize: 10, ...leaveFilters.value, includeSummary: includeSummary === true
    })
    if (data.summary) summary.value = data.summary
    if (version !== leaveVersion) return
    rows.value = data.list || []
    total.value = data.total
    const lastPage = Math.max(1, Math.ceil(total.value / 10))
    if (page.value > lastPage) page.value = lastPage
  } catch (e) {
    if (version === leaveVersion) leaveError.value = error(e)
  } finally {
    if (version === leaveVersion) loading.value = false
  }
}
const refresh = async () => {
  if (!token.value || refreshing.value) return
  refreshing.value = true
  try {
    const profile = await call('/me')
    mustChange.value = Boolean(profile.mustChangePassword)
    if (mustChange.value) return
    me.value = profile
    const data = await call('/classes')
    classes.value = data.list || []
    if (!classes.value.some(cl => String(cl.id) === selectedClassId.value)) {
      selectedClassId.value = ''
      delete leaveFilters.value.classId
    }
    if (!classes.value.some(cl => cl.id === studentClass.value)) {
      studentClass.value = null
      delete studentFilters.value.classId
    }
    await Promise.all([loadLeaves(true), loadStudents()])
  } catch (e) {
    ElMessage.error(error(e))
    if (props.advisorMode && e.response?.status === 403) await returnToWorkspace()
    else if (!props.advisorMode && e.response?.status === 401) clear()
  } finally {
    refreshing.value = false
  }
}
const returnToWorkspace = () => router.replace({ name: 'Login' })
const logout=async()=>{if(props.advisorMode)return userStore.LoginOut();try{await call('/logout','post',{})}catch(e){}finally{clear()}}
const downloadProof = async (row, index) => {
  try {
    await downloadLeaveProof(
      props.advisorMode ? request : client,
      `${portalBase.value}/leaves/${row.id}/proof?index=${index}`,
      props.advisorMode ? undefined : token.value,
      `${row.id}-${index}`
    )
  } catch {
    ElMessage.error('下载证明失败，请重试')
  }
}
const search = () => {
  if ((dateFrom.value || dateTo.value) && (!dateFrom.value || !dateTo.value || dateFrom.value > dateTo.value)) {
    ElMessage.warning('请选择有效的开始和结束日期')
    return
  }
  leaveFilters.value = {
    status: status.value, classId: selectedClassId.value || undefined, keyword: leaveKeyword.value.trim(),
    from: dateFrom.value || undefined, to: dateTo.value || undefined
  }
  if (page.value === 1) loadLeaves()
  else page.value = 1
}
const exportLeaves = async () => {
  exporting.value = true
  try {
    const config = { url: portalBase.value + '/leaves', method: 'get', params: { ...leaveFilters.value, format: 'xlsx' }, responseType: 'blob' }
    const result = props.advisorMode ? await request(config) :
      await client({ ...config, headers: { Authorization: `Bearer ${token.value}` } })
    const url = URL.createObjectURL(result.data)
    const link = document.createElement('a')
    link.href = url
    link.download = '学生请假记录.xlsx'
    link.click()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
  } catch (e) {
    const payload = e.response?.data
    let message
    if (payload instanceof Blob) {
      try { message = JSON.parse(await payload.text()).msg } catch { /* use default error */ }
    }
    ElMessage.error(message || '导出失败，请重试')
  } finally {
    exporting.value = false
  }
}
const review = async (ids, action) => {
  if (reviewing.value) return
  reviewing.value = true
  try {
    const title = action === 'rejected' ? '驳回申请' : '通过申请'
    const { value } = await ElMessageBox.prompt(
      `确认${action === 'rejected' ? '驳回' : '通过'} ${ids.length} 条请假申请？`,
      title,
      { inputPlaceholder: action === 'rejected' ? '填写驳回原因' : '审批备注（可留空）',
        inputValidator: note => action !== 'rejected' || !!note?.trim(),
        inputErrorMessage: '驳回原因不能为空' }
    )
    const result = await call('/leaves/review', 'post', { ids, action, note: value || '' })
    ElMessage.success(`已处理 ${result.count} 条申请`)
    selectedLeave.value = null
    await refresh()
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') {
      ElMessage.error(error(e))
      if (e.response?.status === 409) await refresh()
    }
  } finally {
    reviewing.value = false
  }
}
const changePassword=async()=>{if(busy.value)return;if(!oldPassword.value||Array.from(newPassword.value).length!==6||newPassword.value==="123456"||(mustChange.value&&confirmPassword.value!==newPassword.value)){ElMessage.warning('请输入旧密码，新密码长度须为 6位');return}busy.value=true;try{await call('/change-password','post',{oldPassword:oldPassword.value,newPassword:newPassword.value});ElMessage.success('密码已修改，请重新登录');changeDialog.value=false;oldPassword.value='';newPassword.value='';clear()}catch(e){ElMessage.error(error(e))}finally{busy.value=false}}
onMounted(()=>{if(token.value)refresh()})
</script>
<style scoped>.leave-filters{display:flex;align-items:center;gap:10px;flex-wrap:wrap}@media(max-width:640px){.leave-filters{width:100%}.leave-filters :deep(.el-select){flex:1;min-width:140px}}
.teacher-portal{min-height:100vh;background:#f4f7fb;color:#1e293b}.topbar{background:#fff;padding:18px max(24px,calc((100vw - 1200px)/2));display:flex;justify-content:space-between;align-items:center;border-bottom:1px solid #e2e8f0}.brand{font-size:19px;font-weight:700;color:#1d4ed8}.login-wrap{max-width:430px;margin:90px auto;padding:0 16px}.login-wrap h2{margin:0 0 8px}.muted{color:#64748b;font-size:13px}.main{max-width:1200px;margin:0 auto;padding:28px 18px}.welcome,.section-header{display:flex;justify-content:space-between;align-items:center;gap:12px}.welcome h2{margin:0}.stats{margin:20px 0}.stats h2{margin:8px 0 0;font-size:30px}.section{margin-bottom:18px}.pager{margin-top:16px;display:flex;justify-content:flex-end}@media(max-width:600px){.welcome{align-items:flex-start;flex-direction:column}.main{padding:14px 10px}.topbar{padding:14px}}

/* Legacy scoped scroll rules replaced by portal-specific global rules below. */

</style>

<style scoped>
/* V17.3 TEACHER PORTAL VISUAL UPGRADE */
.teacher-portal.uni-ui { --uni-primary:#2563eb; --uni-border:#e5eaf2; --uni-muted:#64748b; min-height:100vh; background:radial-gradient(circle at 90% 0%,#eaf2ff 0%,transparent 32%),#f5f7fb; color:#17243b; }
.teacher-portal.uni-ui .topbar { min-height:72px; padding-top:14px; padding-bottom:14px; background:rgba(255,255,255,.96); border-bottom:1px solid var(--uni-border); box-shadow:0 2px 16px rgba(15,23,42,.035); }
.teacher-portal.uni-ui .brand { font-size:19px; font-weight:750; letter-spacing:-.4px; color:#1d4ed8; }
.teacher-portal.uni-ui .main { max-width:1280px; padding-top:32px; padding-bottom:56px; }
.teacher-portal.uni-ui .login-wrap { max-width:470px; margin:clamp(45px,10vh,110px) auto; }
.teacher-portal.uni-ui .login-wrap :deep(.el-card) { border:1px solid var(--uni-border); border-radius:22px; overflow:hidden; box-shadow:0 22px 65px rgba(23,54,112,.10); }
.teacher-portal.uni-ui .login-wrap :deep(.el-card__body) { padding:38px 36px 40px; }
.teacher-portal.uni-ui .login-wrap h2 { font-size:27px; font-weight:760; letter-spacing:-.7px; }
.teacher-portal.uni-ui .login-wrap .muted { margin-bottom:26px; line-height:1.7; }
.teacher-portal.uni-ui .login-wrap :deep(.el-form-item__label) { font-weight:600; color:#334155; }
.teacher-portal.uni-ui :deep(.el-input__wrapper), .teacher-portal.uni-ui :deep(.el-select__wrapper) { min-height:40px; border-radius:10px; }
.teacher-portal.uni-ui :deep(.el-button) { border-radius:10px; font-weight:550; }
.teacher-portal.uni-ui :deep(.el-button--primary) { --el-button-bg-color:#2563eb; --el-button-border-color:#2563eb; --el-button-hover-bg-color:#1d4ed8; --el-button-hover-border-color:#1d4ed8; }
.teacher-portal.uni-ui .welcome { position:relative; overflow:hidden; background:linear-gradient(110deg,#183b9b,#2563eb 65%,#4589f9); color:#fff; padding:30px 34px; border-radius:20px; box-shadow:0 14px 35px rgba(37,99,235,.15); }
.teacher-portal.uni-ui .welcome::after { content:''; position:absolute; width:220px; height:220px; right:-90px; top:-120px; border-radius:50%; background:rgba(255,255,255,.08); pointer-events:none; }
.teacher-portal.uni-ui .welcome h2 { font-size:27px; font-weight:750; letter-spacing:-.5px; }
.teacher-portal.uni-ui .welcome .muted { color:rgba(255,255,255,.85); line-height:1.7; }
.teacher-portal.uni-ui .welcome :deep(.el-button--primary) { --el-button-bg-color:#fff; --el-button-border-color:#fff; --el-button-text-color:#1d4ed8; --el-button-hover-bg-color:#eff6ff; --el-button-hover-border-color:#eff6ff; --el-button-hover-text-color:#1d4ed8; }
.teacher-portal.uni-ui .stats { margin:22px 0 8px; }
.teacher-portal.uni-ui .stats :deep(.el-col) { margin-bottom:16px; }
.teacher-portal.uni-ui .stats :deep(.el-card) { height:100%; min-height:122px; border:1px solid var(--uni-border); border-radius:17px; box-shadow:0 5px 20px rgba(15,23,42,.035); transition:transform .2s ease,box-shadow .2s ease; }
.teacher-portal.uni-ui .stats :deep(.el-card__body) { padding:24px; }
.teacher-portal.uni-ui .stats h2 { color:#1d4ed8; font-size:34px; font-weight:780; letter-spacing:-1px; }
.teacher-portal.uni-ui .muted { color:var(--uni-muted); }
.teacher-portal.uni-ui .section { margin-bottom:20px; border:1px solid var(--uni-border); border-radius:17px; box-shadow:0 5px 20px rgba(15,23,42,.035); }
.teacher-portal.uni-ui .section :deep(.el-card__header) { padding:19px 23px; border-bottom:1px solid #edf1f7; font-size:16px; color:#1e293b; }
.teacher-portal.uni-ui .section :deep(.el-card__body) { padding:22px 23px; }
.teacher-portal.uni-ui .section-header { flex-wrap:wrap; }
.teacher-portal.uni-ui :deep(.el-table) { --el-table-header-bg-color:#f8fafc; --el-table-row-hover-bg-color:#f2f7ff; --el-table-border-color:#edf1f7; border-radius:10px; }
.teacher-portal.uni-ui :deep(.el-table th.el-table__cell) { color:#475569; font-weight:650; }
.teacher-portal.uni-ui :deep(.el-tag) { border-radius:7px; }
.teacher-portal.uni-ui .pager { flex-wrap:wrap; gap:6px; }
.teacher-portal.uni-ui :deep(.el-button:focus-visible), .teacher-portal.uni-ui :deep(.el-input__wrapper:focus-within) { outline:2px solid #93c5fd; outline-offset:2px; }
@media (hover:hover) { .teacher-portal.uni-ui .stats :deep(.el-card:hover) { transform:translateY(-2px); box-shadow:0 12px 30px rgba(15,23,42,.075); } }
@media (max-width:640px) {
 .teacher-portal.uni-ui .topbar { min-height:62px; padding:12px 16px; gap:8px; }
 .teacher-portal.uni-ui .brand { font-size:16px; }
 .teacher-portal.uni-ui .main { padding:14px 12px 40px; }
 .teacher-portal.uni-ui .welcome { padding:23px 20px; border-radius:16px; }
 .teacher-portal.uni-ui .welcome h2 { font-size:23px; }
 .teacher-portal.uni-ui .stats :deep(.el-card__body) { padding:17px; }
 .teacher-portal.uni-ui .stats h2 { font-size:28px; }
 .teacher-portal.uni-ui .section { border-radius:14px; }
 .teacher-portal.uni-ui .section :deep(.el-card__header), .teacher-portal.uni-ui .section :deep(.el-card__body) { padding:16px; }
 .teacher-portal.uni-ui .login-wrap { margin:45px auto; padding:0 14px; }
 .teacher-portal.uni-ui .login-wrap :deep(.el-card__body) { padding:30px 23px; }
 .teacher-portal.uni-ui .pager { justify-content:center; }
}
@media (prefers-reduced-motion:reduce) { .teacher-portal.uni-ui .stats :deep(.el-card) { transition:none; transform:none; } }
</style>


<style>
/* V17.7.1 TEACHER PORTAL SCROLL FIX */
/* Only while teacher portal is mounted. Restore document scrolling when outer app styles lock it. */
html:has(.teacher-portal),
body:has(.teacher-portal) {
  height: auto !important;
  min-height: 100% !important;
  overflow-y: auto !important;
}
body:has(.teacher-portal) #app {
  height: auto !important;
  min-height: 100vh !important;
  max-height: none !important;
  overflow-y: visible !important;
}
body:has(.teacher-portal) .teacher-portal {
  height: auto !important;
  min-height: 100vh;
  max-height: none !important;
  overflow: visible !important;
}
</style>

<style scoped>
.teacher-metrics, .class-cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(145px, 1fr)); gap: 12px; margin: 18px 0; }
.teacher-metrics article, .class-cards article { display: grid; gap: 6px; border: 1px solid #e4eaf3; border-radius: 12px; padding: 14px; background: var(--el-bg-color); }
.teacher-metrics span, .class-cards span { color: var(--el-text-color-secondary); }
.teacher-metrics strong { font-size: 27px; }
.teacher-trend { display: grid; grid-template-columns: repeat(7, minmax(0, 1fr)); align-items: end; gap: 6px; min-height: 100px; }
.teacher-trend > div { display: grid; justify-items: center; gap: 5px; font-size: 12px; }
.trend-bar { width: min(32px, 65%); min-height: 2px; background: var(--el-color-primary); border-radius: 4px 4px 0 0; }
.teacher-filters, .batch-tools { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; margin-bottom: 14px; }
.teacher-filters .el-input, .teacher-filters .el-select { width: 155px; }
.teacher-filters label { display: flex; align-items: center; gap: 6px; }
.teacher-filters input { color: var(--el-text-color-primary); background: var(--el-bg-color); border: 1px solid var(--el-border-color); border-radius: 6px; padding: 5px; }
.approval-card { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; padding: 12px 0; border-top: 1px solid var(--el-border-color-lighter); }
.approval-overview { min-width: 125px; flex: 1; display: grid; gap: 4px; background: none; border: 0; text-align: left; color: inherit; cursor: pointer; padding: 5px; }
.approval-overview span, .approval-overview small { color: var(--el-text-color-secondary); }
.approval-actions { display: flex; gap: 6px; }
.approval-actions .el-button, .batch-tools .el-button, .teacher-filters .el-button { margin-left: 0; }
.teacher-detail { display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 12px; }
.teacher-detail dd { margin: 0; white-space: pre-wrap; overflow-wrap: anywhere; }
@media (max-width: 600px) {
  .teacher-metrics { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .teacher-filters .el-input, .teacher-filters .el-select { width: 100%; }
  .approval-actions { width: 100%; justify-content: flex-end; }
  .approval-actions .el-button { min-height: 42px; }
  .teacher-portal .pager { flex-wrap: wrap; }
}
</style>
