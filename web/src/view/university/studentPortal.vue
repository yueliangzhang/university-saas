<template>
  <div class="student-portal uni-ui">
    <div v-if="token && mustChange" style="max-width:440px;margin:60px auto;padding:24px"><el-card><h2>首次登录，请修改密码</h2><el-form label-position="top" @submit.prevent="changePassword"><el-form-item label="当前密码"><el-input v-model="oldPassword" type="password"/></el-form-item><el-form-item label="新密码（6位）"><el-input v-model="newPassword" type="password"/></el-form-item><el-form-item label="确认新密码"><el-input v-model="confirmPassword" type="password"/></el-form-item><el-button native-type="submit" type="primary">确认修改</el-button><el-button @click="logout">退出</el-button></el-form></el-card></div>
<template v-else-if="token && me">
      <el-container class="app-shell">
        <el-aside class="sidebar" width="236px">
          <div class="sidebar-brand"><span class="brand-icon"><el-icon><School /></el-icon></span><span>学生服务平台<small>STUDENT PORTAL</small></span></div>
          <div class="nav-label">工作台</div>
          <button v-for="item in navigation" :key="item.key" class="nav-item" :class="{active:activeTab===item.key}" :aria-current="activeTab === item.key ? 'page' : undefined" @click="navigate(item.key)">
            <el-icon :size="18"><component :is="item.icon" /></el-icon><span>{{item.label}}</span>
          </button>
          <div class="sidebar-bottom">
            <div class="sidebar-user"><el-avatar :size="36">{{ (me?.name || '学').slice(0,1) }}</el-avatar><div><strong>{{me?.name||'学生'}}</strong><small>{{me?.studentNo||'—'}}</small></div></div>
            <el-button class="logout-button" text @click="logout"><el-icon><SwitchButton /></el-icon> 退出登录</el-button>
          </div>
        </el-aside>

        <el-container class="main-container">
          <el-header class="topbar">
            <div class="topbar-left"><el-icon class="topbar-icon"><School /></el-icon><span>{{ me.schoolName || '高校学生服务平台' }}</span><span class="topbar-separator">/</span><strong>{{ currentTabLabel }}</strong></div>
            <div class="topbar-right"><LeaveNotifications :call="callNotifications" :open-leave="openNotificationLeave" /><el-avatar :size="32">{{ (me?.name || '学').slice(0,1) }}</el-avatar></div>
          </el-header>
          <el-main class="content">
            <template v-if="activeTab==='dashboard'">
              <div class="welcome-card">
                <div><h1>你好，{{me?.name||'同学'}}</h1><p>{{ [me.majorName, me.className].filter(Boolean).join(' · ') || me.studentNo }}</p><el-button type="primary" size="large" @click="openLeaveDialog"><el-icon><Plus /></el-icon> 发起请假申请</el-button></div>
                <div class="welcome-decoration"><el-icon :size="90"><School /></el-icon></div>
              </div>
              <div class="stats-grid">
                <div v-for="stat in stats" :key="stat.key" class="stat-card"><div class="stat-icon" :class="stat.key"><el-icon :size="22"><component :is="stat.icon" /></el-icon></div><div><div class="stat-label">{{stat.label}}</div><div class="stat-value">{{stat.value}}</div></div></div>
              </div>
              <div class="section-card">
                <div class="section-header"><div><h2>最近请假记录</h2><p>随时了解申请的最新审批状态</p></div><el-button text type="primary" @click="activeTab='leaves'">查看全部 <el-icon><Right /></el-icon></el-button></div>
                <el-skeleton v-if="dashboardLoading" :rows="3" animated />
                <el-alert v-else-if="dashboardError" :title="dashboardError" type="error" :closable="false">
                  <el-button link @click="loadDashboard">重试</el-button>
                </el-alert>
                <el-empty v-else-if="!recentLeaves.length" description="暂无请假记录" :image-size="64" />
                <button v-for="row in recentLeaves" :key="row.id" class="leave-card" @click="selectedLeave = row">
                  <strong>{{ leaveTypes[row.leaveType] || row.leaveType }}</strong>
                  <el-tag :type="statusType(row.status)" round>{{ leaveStatuses[row.status] }}</el-tag>
                  <span>{{ leaveFmt(row.startAt) }} 至 {{ leaveFmt(row.endAt) }}</span>
                </button>
              </div>
            </template>

            <template v-else-if="activeTab==='leaves'">
              <div class="page-heading"><h1>请假记录</h1><el-button type="primary" @click="openLeaveDialog">申请请假</el-button></div>
              <div class="section-card">
                <div class="record-filters">
                  <el-select v-model="statusFilter" aria-label="审批状态" @change="refreshLeaves">
                    <el-option label="全部状态" value="all" />
                    <el-option v-for="(label, key) in leaveStatuses" :key="key" :label="label" :value="key" />
                  </el-select>
                  <label>开始日期 <input v-model="dateFrom" type="date" :max="dateTo || undefined" @change="applyDates" /></label>
                  <label>结束日期 <input v-model="dateTo" type="date" :min="dateFrom || undefined" @change="applyDates" /></label>
                  <el-button v-if="dateFrom || dateTo" @click="clearDates">清除日期</el-button>
                </div>
                <p v-if="dateError" class="date-error" role="alert">{{ dateError }}</p>
                <el-skeleton v-if="leaveLoading" :rows="4" animated />
                <el-alert v-else-if="recordsError" :title="recordsError" type="error" :closable="false">
                  <el-button link @click="loadLeaves">重试</el-button>
                </el-alert>
                <el-empty v-else-if="!leaveRows.length" description="暂无符合条件的请假记录" :image-size="64" />
                <button v-else v-for="row in leaveRows" :key="row.id" class="leave-card" @click="selectedLeave = row">
                  <strong>{{ leaveTypes[row.leaveType] || row.leaveType }}</strong>
                  <el-tag :type="statusType(row.status)" round>{{ leaveStatuses[row.status] }}</el-tag>
                  <span>{{ leaveFmt(row.startAt) }} 至 {{ leaveFmt(row.endAt) }}</span>
                </button>
                <div v-if="leaveTotal > 10" class="pagination">
                  <el-pagination v-model:current-page="leavePage" :page-size="10" :total="leaveTotal"
                    :pager-count="5" layout="prev, pager, next" @current-change="loadLeaves" />
                </div>
              </div>
            </template>

            <template v-else-if="activeTab==='profile'">
              <div class="page-heading"><div><h1>个人中心</h1><p>查看你的在校基本信息</p></div></div>
              <div class="section-card profile-card"><div class="profile-banner"><el-avatar :size="66">{{ (me?.name||'学').slice(0,1) }}</el-avatar><div><h2>{{me?.name||'学生'}}</h2><p>学号 {{me?.studentNo||'—'}}</p></div></div><el-descriptions :column="profileColumns" border class="profile-details"><el-descriptions-item label="姓名">{{me?.name||'—'}}</el-descriptions-item><el-descriptions-item label="学号">{{me?.studentNo||'—'}}</el-descriptions-item><el-descriptions-item label="高校">{{me.schoolName || '—'}}</el-descriptions-item><el-descriptions-item label="学院">{{me.collegeName || '—'}}</el-descriptions-item><el-descriptions-item label="专业">{{me.majorName || '—'}}</el-descriptions-item><el-descriptions-item label="班级">{{me.className || '—'}}</el-descriptions-item><el-descriptions-item label="所属辅导员">{{me?.advisorName||'暂未分配'}}</el-descriptions-item><el-descriptions-item label="入学年份">{{me?.enrollmentYear||'—'}}</el-descriptions-item></el-descriptions><p class="muted profile-hint">如需修改学籍信息，请联系学校管理员。</p><div class="profile-actions"><el-button @click="activeTab='security'">修改密码</el-button><el-button @click="appStore.toggleTheme(!appStore.isDark)">{{ appStore.isDark ? '浅色模式' : '深色模式' }}</el-button><el-button @click="logout">退出登录</el-button></div></div>
            </template>

            <template v-else>
              <div class="page-heading"><div><h1>账号安全</h1><p>管理登录密码，保护你的学生账号</p></div></div>
              <div class="section-card security-card"><div class="security-heading"><span class="security-icon"><el-icon :size="23"><Lock /></el-icon></span><div><h2>修改登录密码</h2><p>建议使用6 位密码，并避免与其他网站重复。</p></div></div><el-form label-position="top" class="password-form" @submit.prevent="changePassword"><el-form-item label="当前密码"><el-input v-model="oldPassword" type="password" show-password autocomplete="current-password" placeholder="请输入当前密码" /></el-form-item><el-form-item label="新密码"><el-input v-model="newPassword" type="password" show-password autocomplete="new-password" placeholder="固定6位" /></el-form-item><el-button type="primary" native-type="submit" :loading="passwordSaving" :disabled="passwordSaving">保存新密码</el-button></el-form><el-alert class="password-tip" title="密码修改成功后，当前会话将自动退出，请使用新密码重新登录。" type="info" :closable="false" show-icon /></div>
            </template>
          </el-main>
        </el-container>
      </el-container>

      <nav class="mobile-nav" aria-label="学生服务导航"><button v-for="item in navigation" :key="item.key" :class="{active:activeTab===item.key}" :aria-current="activeTab === item.key ? 'page' : undefined" @click="navigate(item.key)"><el-icon :size="20"><component :is="item.icon" /></el-icon><span>{{item.shortLabel}}</span></button></nav>

      <el-dialog v-model="leaveDialog" title="申请请假" :close-on-click-modal="!leaveSaving" :close-on-press-escape="!leaveSaving" :show-close="!leaveSaving" class="leave-dialog" width="min(560px, 94vw)" append-to-body destroy-on-close>
        <div class="application-advisor">
          <span>审批辅导员</span><strong>{{ me.advisorName || '暂未分配' }}</strong>
        </div>
        <el-form label-position="top" @submit.prevent="submitLeave">
          <el-form-item label="请假类型" required :error="leaveErrors.leaveType">
            <el-radio-group v-model="leaveForm.leaveType" @change="delete leaveErrors.leaveType">
              <el-radio-button v-for="(label, key) in leaveTypes" :key="key" :value="key">{{ label }}</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="开始时间" required :error="leaveErrors.startAt"><el-date-picker v-model="leaveForm.startAt" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width:100%" placeholder="请选择开始时间" :disabled-date="beforeToday" @change="clearLeaveTimeErrors"/></el-form-item>
          <el-form-item label="结束时间" required :error="leaveErrors.endAt"><el-date-picker v-model="leaveForm.endAt" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width:100%" placeholder="请选择结束时间" :disabled-date="beforeStartDate" @change="clearLeaveTimeErrors"/></el-form-item>
          <p v-if="leaveDuration" class="application-duration">请假时长 <strong>{{ leaveDuration }}</strong></p>
          <el-form-item label="请假原因" required :error="leaveErrors.reason"><el-input v-model="leaveForm.reason" type="textarea" :rows="4" maxlength="2000" show-word-limit placeholder="请简要说明请假原因" @input="delete leaveErrors.reason"/></el-form-item>
          <el-form-item label="证明文件（选填）">
            <input type="file" multiple aria-label="证明文件（选填）" accept=".pdf,.jpg,.jpeg,.png" :disabled="leaveSaving" @change="selectProof" />
            <div class="muted w-full">最多 5 个文件，支持 PDF、JPG、PNG，每个不超过 5 MB</div>
            <ul v-if="leaveProofs.length" class="w-full">
              <li v-for="(file, index) in leaveProofs" :key="index" class="flex items-center gap-2">
                <span class="min-w-0 break-all">{{ file.name }}</span>
                <el-button type="danger" link :disabled="leaveSaving" :aria-label="'移除 ' + file.name" @click="leaveProofs.splice(index, 1)">移除</el-button>
              </li>
            </ul>
          </el-form-item>
        </el-form>
        <template #footer><el-button :disabled="leaveSaving" @click="leaveDialog=false">取消</el-button><el-button type="primary" :loading="leaveSaving" :disabled="leaveSaving" @click="submitLeave">提交申请</el-button></template>
      </el-dialog>
      <el-dialog :model-value="Boolean(selectedLeave)" title="请假详情" width="min(560px, 94vw)"
        @update:model-value="value => { if (!value) selectedLeave = null }">
        <dl v-if="selectedLeave" class="leave-details">
          <dt>请假类型</dt><dd>{{ leaveTypes[selectedLeave.leaveType] }}</dd>
          <dt>审批状态</dt><dd>{{ leaveStatuses[selectedLeave.status] }}</dd>
          <dt>开始时间</dt><dd>{{ leaveFmt(selectedLeave.startAt) }}</dd>
          <dt>结束时间</dt><dd>{{ leaveFmt(selectedLeave.endAt) }}</dd>
          <dt>请假时长</dt><dd>{{ formatDuration(selectedLeave.startAt, selectedLeave.endAt) }}</dd>
          <dt>提交时间</dt><dd>{{ leaveFmt(selectedLeave.createdAt) || '—' }}</dd>
          <dt>请假原因</dt><dd>{{ selectedLeave.reason }}</dd>
          <dt>审批时间</dt><dd>{{ leaveFmt(selectedLeave.reviewedAt) || '—' }}</dd>
          <dt>审批意见</dt><dd>{{ selectedLeave.reviewNote || '—' }}</dd>
          <dt v-if="selectedLeave.hasProof">证明文件</dt>
          <dd v-if="selectedLeave.hasProof">
            <el-button v-for="index in (selectedLeave.proofCount || 1)" :key="index"
              link type="primary" @click="downloadProof(selectedLeave, index)">附件 {{ index }}</el-button>
          </dd>
        </dl>
        <template #footer>
          <el-button @click="selectedLeave = null">关闭</el-button>
          <el-button v-if="selectedLeave?.status === 'pending'" type="danger" plain
            :loading="cancelling" @click="cancelLeave(selectedLeave)">撤销申请</el-button>
        </template>
      </el-dialog>
    </template>
    <section v-else-if="token" class="profile-loading" aria-live="polite">
      <el-skeleton v-if="profileLoading" :rows="4" animated />
      <el-alert v-else-if="profileError" :title="profileError" type="error" :closable="false">
        <el-button link @click="initialize">重试</el-button>
      </el-alert>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { School, CircleCheck, Clock, Lock, Right, Plus, House, Calendar, User, SwitchButton, Files } from '@element-plus/icons-vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/pinia'
import { downloadLeaveProof } from '@/utils/leaveProof'
import LeaveNotifications from '@/components/university-ui/LeaveNotifications.vue'

const mustChange = ref(false)
const confirmPassword = ref('')
const token = ref(sessionStorage.getItem('uni_student_token') || '')
const router = useRouter()
const appStore = useAppStore()
watch(token, value => {
  if (!value) router.replace({ name: 'Login' })
})
const me = ref(null)
const profileLoading = ref(true)
const profileError = ref('')
const oldPassword = ref('')
const newPassword = ref('')
const passwordSaving = ref(false)
const activeTab = ref('dashboard')
const navigation = [
  { key: 'dashboard', label: '首页概览', shortLabel: '首页', icon: House },
  { key: 'apply', label: '请假申请', shortLabel: '请假', icon: Calendar },
  { key: 'leaves', label: '请假记录', shortLabel: '记录', icon: Files },
  { key: 'profile', label: '个人中心', shortLabel: '我的', icon: User },

]
const currentTabLabel = computed(() => navigation.find(n => n.key === activeTab.value)?.label || '账号安全')
const profileColumns = ref(window.innerWidth < 640 ? 1 : 2)
const onResize = () => { profileColumns.value = window.innerWidth < 640 ? 1 : 2 }

const base = (import.meta.env.VITE_BASE_API || '/api').replace(/\/$/, '')
const client = axios.create({ baseURL: base, timeout: 15000 })
const call = async (url, method = 'get', data, params) => {
  const r = await client({ url, method, data, params, headers: token.value ? { Authorization: `Bearer ${token.value}` } : {} })
  if (r.data?.code !== 0) throw Error(r.data?.msg || '请求失败')
  return r.data.data
}
const leaveError = e => e.response?.data?.msg || e.message || '操作失败'
const clear = () => {
 mustChange.value=false; confirmPassword.value=""
  token.value = ''
  sessionStorage.removeItem('uni_student_token')
  me.value = null
  oldPassword.value = ''
  newPassword.value = ''
  activeTab.value = 'dashboard'
  leaveRows.value = []
  leaveTotal.value = 0
  leavePage.value = 1
}
const initialize = async () => {
  profileLoading.value = true
  profileError.value = ''
  try {
    const profile = await call('/uni/student-portal/me')
    mustChange.value = Boolean(profile.mustChangePassword)
    if (!mustChange.value) {
      me.value = profile
      await loadDashboard()
    }
  } catch (error) {
    if (error.response?.status === 401) clear()
    else profileError.value = leaveError(error)
  } finally {
    profileLoading.value = false
  }
}
const logout = async () => {
  try { await call('/uni/student-portal/logout', 'post', {}) }
  catch (e) { /* local session still cleared */ }
  finally { clear() }
}
const changePassword = async () => {
  if (passwordSaving.value) return
  if (!oldPassword.value || Array.from(newPassword.value).length !== 6 || newPassword.value === "123456" || (mustChange.value && confirmPassword.value !== newPassword.value)) {
    ElMessage.warning('请输入当前密码，新密码长度应为 6位'); return
  }
  passwordSaving.value = true
  try {
    await call('/uni/student-portal/change-password', 'post', { oldPassword: oldPassword.value, newPassword: newPassword.value })
    clear()
    ElMessage.success('密码修改成功，请重新登录')
  } catch (e) { ElMessage.error(leaveError(e)) }
  finally { passwordSaving.value = false }
}

const leaveTypes = { sick: '病假', personal: '事假', other: '其他' }
const leaveStatuses = { pending: '待审批', approved: '已通过', rejected: '已驳回', cancelled: '已撤销' }
const leaveRows = ref([])
const leaveTotal = ref(0)
const leavePage = ref(1)
const leaveLoading = ref(false)
const leaveSaving = ref(false)
const leaveDialog = ref(false)
const leaveProofs = ref([])
const maxProofCount = 5
const maxProofSize = 5 * 1024 * 1024
const leaveForm = ref({ leaveType: 'personal', reason: '', startAt: '', endAt: '' })
const leaveErrors = ref({})
const statusFilter = ref('all')
const dateFilter = ref(null)
const dateFrom = ref('')
const dateTo = ref('')
const dateError = ref('')
const selectedLeave = ref(null)
const callNotifications = (url, method, data, params) => call('/uni/student-portal' + url, method, data, params)
const openNotificationLeave = async id => {
  const result = await call('/uni/student-portal/leaves', 'get', undefined, { id, pageSize: 1 })
  const leave = result.list?.[0]
  if (!leave) throw new Error('该请假记录已不可访问')
  selectedLeave.value = leave
}

const cancelling = ref(false)
const recentLeaves = ref([])
const summary = ref(null)
const recordsError = ref('')
const dashboardError = ref('')
const dashboardLoading = ref(false)
let leaveRequestVersion = 0
let dashboardRequestVersion = 0

const stats = computed(() => [
  { key: 'total', label: '累计请假', value: summary.value?.total ?? '—', icon: Files },
  { key: 'pending', label: '待审批', value: summary.value?.pending ?? '—', icon: Clock },
  { key: 'approved', label: '已批准', value: summary.value?.approved ?? '—', icon: CircleCheck }
])
const statusType = status => ({ pending: 'warning', approved: 'success', rejected: 'danger', cancelled: 'info' })[status] || 'info'
const leaveFmt = value => value ? String(value).replace('T', ' ').slice(0, 16) : ''
const formatDuration = (start, end) => {
  const minutes = Math.round((Date.parse(leaveFmt(end).replace(' ', 'T')) - Date.parse(leaveFmt(start).replace(' ', 'T'))) / 60000)
  if (!(minutes > 0)) return ''
  const days = Math.floor(minutes / 1440)
  const hours = Math.floor(minutes / 60) % 24
  const remainder = minutes % 60
  return [days && `${days}天`, hours && `${hours}小时`, remainder && `${remainder}分钟`].filter(Boolean).join('')
}
const leaveDuration = computed(() => formatDuration(leaveForm.value.startAt, leaveForm.value.endAt))
const beforeToday = day => day < new Date(new Date().setHours(0, 0, 0, 0))
const beforeStartDate = day => beforeToday(day) ||
  Boolean(leaveForm.value.startAt && day < new Date(leaveForm.value.startAt.slice(0, 10) + 'T00:00:00'))

const loadDashboard = async () => {
  if (!token.value) return
  const version = ++dashboardRequestVersion
  dashboardLoading.value = true
  dashboardError.value = ''
  try {
    const result = await call('/uni/student-portal/leaves', 'get', undefined, { pageSize: 3, includeSummary: true })
    if (version !== dashboardRequestVersion) return
    recentLeaves.value = result.list || []
    summary.value = result.summary
  } catch (error) {
    if (error.response?.status === 401) clear()
    else if (version === dashboardRequestVersion) dashboardError.value = leaveError(error)
  } finally {
    if (version === dashboardRequestVersion) dashboardLoading.value = false
  }
}

const loadLeaves = async () => {
  if (!token.value) return
  const version = ++leaveRequestVersion
  leaveLoading.value = true
  recordsError.value = ''
  const [from, to] = dateFilter.value || []
  try {
    const result = await call('/uni/student-portal/leaves', 'get', undefined, {
      page: leavePage.value, pageSize: 10, from, to,
      status: statusFilter.value === 'all' ? undefined : statusFilter.value
    })
    if (version !== leaveRequestVersion) return
    leaveTotal.value = result.total
    leaveRows.value = result.list || []
    const lastPage = Math.max(1, Math.ceil(leaveTotal.value / 10))
    if (leavePage.value > lastPage) leavePage.value = lastPage
  } catch (error) {
    if (version !== leaveRequestVersion) return
    if (error.response?.status === 401) clear()
    else recordsError.value = leaveError(error)
  } finally {
    if (version === leaveRequestVersion) leaveLoading.value = false
  }
}
const refreshLeaves = () => {
  if (leavePage.value === 1) loadLeaves()
  else leavePage.value = 1
}
const applyDates = () => {
  if (!dateFrom.value || !dateTo.value || dateFrom.value > dateTo.value) {
    dateError.value = '请选择有效的开始和结束日期'
    return
  }
  dateError.value = ''
  dateFilter.value = [dateFrom.value, dateTo.value]
  refreshLeaves()
}
const clearDates = () => {
  dateFrom.value = ''
  dateTo.value = ''
  dateError.value = ''
  dateFilter.value = null
  refreshLeaves()
}
const navigate = tab => {
  if (tab === 'apply') openLeaveDialog()
  else activeTab.value = tab
}
watch(activeTab, tab => {
  if (tab === 'leaves') loadLeaves()
  if (tab === 'dashboard') loadDashboard()
})
const selectProof = event => {
  const files = Array.from(event.target.files)
  event.target.value = ''
  if (leaveProofs.value.length + files.length > maxProofCount) {
    ElMessage.warning('最多选择 5 个证明文件')
    return
  }
  if (files.some(file => !/\.(pdf|jpe?g|png)$/i.test(file.name) || !file.size || file.size > maxProofSize)) {
    ElMessage.warning('请选择 5 MB 以内的 PDF、JPG 或 PNG 文件')
    return
  }
  leaveProofs.value.push(...files)
}
const openLeaveDialog = () => {
  leaveForm.value = { leaveType: 'personal', reason: '', startAt: '', endAt: '' }
  leaveProofs.value = []
  leaveErrors.value = {}
  leaveDialog.value = true
}
const clearLeaveTimeErrors = () => {
  delete leaveErrors.value.startAt
  delete leaveErrors.value.endAt
}
const validateLeave = fields => {
  const errors = {}
  const start = Date.parse(fields.startAt?.replace(' ', 'T'))
  const end = Date.parse(fields.endAt?.replace(' ', 'T'))
  if (!Object.hasOwn(leaveTypes, fields.leaveType)) errors.leaveType = '请选择请假类型'
  if (!Number.isFinite(start)) errors.startAt = '请选择有效的开始时间'
  else if (start <= Date.now()) errors.startAt = '开始时间必须晚于当前时间'
  if (!Number.isFinite(end)) errors.endAt = '请选择有效的结束时间'
  else if (end <= start) errors.endAt = '结束时间必须晚于开始时间'
  if (!fields.reason) errors.reason = '请填写请假原因'
  else if (new TextEncoder().encode(fields.reason).length > 2000) errors.reason = '请假原因不能超过 2000 字节'
  leaveErrors.value = errors
  return Object.keys(errors).length === 0
}
const downloadProof = async (row, index) => {
  try {
    await downloadLeaveProof(client, `/uni/student-portal/leaves/${row.id}/proof?index=${index}`, token.value, `${row.id}-${index}`)
  } catch {
    ElMessage.error('下载证明失败，请重试')
  }
}
const submitLeave = async () => {
  if (leaveSaving.value) return
  const f = leaveForm.value
  const fields = { leaveType: f.leaveType, reason: f.reason.trim(), startAt: f.startAt, endAt: f.endAt }
  if (!validateLeave(fields)) return
  leaveSaving.value = true
  try {
    if (leaveProofs.value.length) {
      const body = new FormData()
      for (const [key, value] of Object.entries(fields)) body.append(key, value)
      for (const file of leaveProofs.value) body.append('proof', file)
      // Let the browser set Content-Type with the multipart boundary.
      const response = await fetch(`${base}/uni/student-portal/leaves`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token.value}` },
        body,
        signal: AbortSignal.timeout(90000)
      })
      if (response.status === 413) throw Error('证明文件超出服务器上传限制')
      const result = await response.json().catch(() => null)
      if (response.status === 400 && result?.msg === '请求格式不合法') {
        throw Error('服务器未识别证明附件请求，请确认后端已更新，再重试提交')
      }
      if (!response.ok || result?.code !== 0) throw Error(result?.msg || '请假申请提交失败')
    } else {
      await call('/uni/student-portal/leaves', 'post', fields)
    }
    ElMessage.success('请假申请已提交')
    leaveDialog.value = false
    if (activeTab.value === 'leaves') refreshLeaves()
    await loadDashboard()
  } catch (e) { ElMessage.error(leaveError(e)) }
  finally { leaveSaving.value = false }
}
const cancelLeave = async row => {
  if (cancelling.value) return
  cancelling.value = true
  try {
    await ElMessageBox.confirm('确认撤销这条待审批请假申请？', '撤销确认', { type: 'warning', confirmButtonText: '确认撤销', cancelButtonText: '取消' })
    await call('/uni/student-portal/leaves/cancel', 'post', { id: row.id })
    ElMessage.success('已撤销')
    selectedLeave.value = null
    if (activeTab.value === 'leaves') refreshLeaves()
    await loadDashboard()
  } catch (e) { if (e !== 'cancel' && e !== 'close') ElMessage.error(leaveError(e)) }
  finally { cancelling.value = false }
}
onMounted(() => {
  window.addEventListener('resize', onResize)
  if (token.value) initialize()
})
onUnmounted(() => window.removeEventListener('resize', onResize))
</script>

<style scoped>
.student-portal{min-height:100vh;background:#f5f7fb;color:#17233c;font-family:Inter,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}
.student-portal :deep(.el-button){font-weight:550}.student-portal :deep(.el-button--primary){--el-button-bg-color:#3266e8;--el-button-border-color:#3266e8;--el-button-hover-bg-color:#2457d5;--el-button-hover-border-color:#2457d5}
.app-shell{min-height:100vh}.sidebar{background:#fff;border-right:1px solid #e9edf5;display:flex;flex-direction:column;min-height:100vh;padding:27px 14px 20px;position:sticky;top:0;height:100vh}.sidebar-brand{display:flex;align-items:center;gap:11px;font-size:17px;font-weight:750;padding:0 10px 36px}.sidebar-brand small{display:block;font-size:10px;font-weight:650;letter-spacing:1.5px;color:#94a1b6;margin-top:3px}.brand-icon,.brand-mark{background:#3266e8;color:#fff;border-radius:12px;display:flex;align-items:center;justify-content:center}.brand-icon{width:39px;height:39px;font-size:21px}.nav-label{padding:0 16px 12px;color:#9aa6b7;font-size:11px;letter-spacing:1px;font-weight:700}.nav-item{width:100%;display:flex;align-items:center;gap:13px;background:transparent;border:0;border-radius:10px;padding:13px 15px;margin-bottom:6px;text-align:left;color:#66758d;cursor:pointer;font-size:14px;font-weight:550}.nav-item:hover{background:#f4f7ff;color:#3266e8}.nav-item.active{background:#edf3ff;color:#3266e8;font-weight:700}.sidebar-bottom{margin-top:auto;border-top:1px solid #eef1f6;padding:20px 8px 0}.sidebar-user{display:flex;gap:10px;align-items:center;margin-bottom:16px}.sidebar-user :deep(.el-avatar),.topbar-right :deep(.el-avatar),.profile-banner :deep(.el-avatar){background:#dbe7ff;color:#3266e8;font-weight:750}.sidebar-user strong,.sidebar-user small{display:block}.sidebar-user strong{font-size:13px}.sidebar-user small{font-size:12px;color:#95a1b3;margin-top:3px}.logout-button{color:#758298!important}.main-container{min-width:0}.topbar{height:70px;background:#fff;border-bottom:1px solid #e9edf5;display:flex;justify-content:space-between;align-items:center;padding:0 36px}.topbar-left,.topbar-right{display:flex;align-items:center;gap:14px;font-size:13px;color:#8a98ad}.topbar-left strong{color:#34415a;font-weight:650}.topbar-icon{color:#3266e8;font-size:18px}.topbar-separator{color:#ccd4df}.content{max-width:1400px;width:100%;margin:0 auto;padding:32px 36px 80px}.welcome-card{background:linear-gradient(110deg,#244ed0 0%,#3875f1 100%);color:white;border-radius:18px;min-height:212px;padding:34px 40px;display:flex;justify-content:space-between;align-items:center;overflow:hidden;position:relative}.welcome-eyebrow{font-size:11px;letter-spacing:2px;opacity:.75;font-weight:700}.welcome-card h1{font-size:29px;letter-spacing:-.5px;margin:13px 0 9px}.welcome-card p{font-size:13px;color:#dbe7ff;margin:0 0 23px;line-height:1.7}.welcome-card :deep(.el-button){background:#fff;border-color:#fff;color:#2455d4}.welcome-decoration{opacity:.16;transform:rotate(-13deg);padding:0 5%}.wave{font-size:24px}.stats-grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:18px;margin:22px 0}.stat-card{background:#fff;border:1px solid #e9edf5;border-radius:14px;padding:23px 18px;display:flex;align-items:center;gap:15px;min-width:0}.stat-icon{width:48px;height:48px;flex:none;border-radius:12px;display:flex;align-items:center;justify-content:center;background:#edf3ff;color:#3972e9}.stat-icon.pending{background:#fff5e7;color:#f0a43d}.stat-icon.approved{background:#e9f9f1;color:#28b47a}.stat-icon.rejected{background:#fff0f0;color:#e96a72}.stat-label{font-size:12px;color:#8895a9}.stat-value{font-size:26px;font-weight:750;margin-top:6px;line-height:1}.section-card{background:#fff;border:1px solid #e9edf5;border-radius:15px;padding:24px;margin-top:22px}.section-header{display:flex;align-items:center;justify-content:space-between;gap:15px;margin-bottom:18px}.section-header h2,.security-heading h2{font-size:17px;margin:0 0 7px}.section-header p,.security-heading p{color:#9aa5b6;font-size:12px;margin:0}.data-table{width:100%;--el-table-header-bg-color:#f8faff;--el-table-border-color:#edf1f6;--el-table-row-hover-bg-color:#f7faff}.data-table :deep(th.el-table__cell){color:#7e8ca3;font-weight:600}.page-heading{display:flex;justify-content:space-between;align-items:center;gap:16px;margin-bottom:22px}.page-heading h1{font-size:25px;margin:0 0 9px}.page-heading p{color:#8d9aaf;font-size:13px;margin:0}.compact-stats{margin-top:0}.filter-actions{display:flex;gap:10px}.pagination{display:flex;justify-content:flex-end;margin-top:20px}.filter-note{font-size:12px;color:#8b98ab}.muted{color:#9aa6b6}.profile-card{padding:0;overflow:hidden}.profile-banner{background:linear-gradient(100deg,#f0f5ff,#f9fbff);padding:35px;display:flex;align-items:center;gap:18px}.profile-banner h2{margin:0 0 7px}.profile-banner p{color:#8a98aa;margin:0}.profile-details{margin:30px}.profile-hint{margin:0 30px 30px;font-size:12px}.security-card{max-width:680px}.security-heading{display:flex;align-items:center;gap:16px;margin-bottom:28px}.security-icon{width:48px;height:48px;border-radius:12px;background:#edf3ff;color:#3266e8;display:flex;align-items:center;justify-content:center}.password-form{max-width:440px}.password-tip{margin-top:25px}.mobile-nav{display:none}
@media(max-width:1100px){.stats-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.content{padding:26px 24px 80px}.topbar{padding:0 24px}.sidebar{width:205px!important}.welcome-decoration{display:none}}
@media(max-width:760px){.sidebar{display:none}.topbar{height:58px;padding:0 17px}.topbar-left{gap:8px;font-size:12px}.topbar-right{gap:7px}.topbar-icon,.topbar-separator{display:none}.content{padding:18px 14px 90px}.welcome-card{padding:25px 22px;min-height:185px}.welcome-card h1{font-size:24px}.stats-grid{gap:10px;margin:14px 0}.stat-card{padding:15px 12px;gap:10px}.stat-icon{width:37px;height:37px}.stat-value{font-size:22px}.stat-label{font-size:11px}.section-card{padding:15px;margin-top:14px}.section-header{align-items:flex-start}.page-heading h1{font-size:22px}.page-heading{align-items:flex-start}.filters{flex-direction:column}.filter-actions{width:100%;justify-content:space-between}.mobile-nav{display:flex;position:fixed;bottom:0;left:0;right:0;background:#fff;border-top:1px solid #e9edf5;z-index:100;padding:8px 4px max(8px,env(safe-area-inset-bottom));justify-content:space-around;box-shadow:0 -3px 16px rgba(32,50,85,.04)}.mobile-nav button{border:0;background:transparent;display:flex;flex-direction:column;align-items:center;gap:4px;color:#8996a9;font-size:10px;padding:4px 6px}.mobile-nav button.active{color:#3266e8;font-weight:700}.profile-banner{padding:22px}.profile-details{margin:18px}.profile-hint{margin:0 18px 22px}}

/* V17 Scroll Fix */
html, body {
  overflow-y: auto;
}

.uni-portal,
.teacher-portal,
.student-portal,
.portal-layout,
.portal-main,
.portal-content {
  height: auto;
  min-height: 100vh;
  overflow-y: visible;
}

.el-main {
  overflow-y: auto;
}

</style>

<style scoped>
/* V17.2 STUDENT PORTAL VISUAL UPGRADE */
.student-portal.uni-ui { --uni-primary:#2563eb; --uni-primary-hover:#1d4ed8; --uni-border:#e5eaf2; --uni-bg:#f5f7fb; background:var(--uni-bg); color:var(--uni-text,#17243b); }
.student-portal.uni-ui :deep(.el-button--primary) { --el-button-bg-color:var(--uni-primary); --el-button-border-color:var(--uni-primary); --el-button-hover-bg-color:var(--uni-primary-hover); --el-button-hover-border-color:var(--uni-primary-hover); }
.student-portal.uni-ui :deep(.el-input__wrapper), .student-portal.uni-ui :deep(.el-select__wrapper) { border-radius:10px; min-height:38px; }
.student-portal.uni-ui :deep(.el-button) { border-radius:10px; }
.student-portal.uni-ui :deep(.el-table) { --el-table-header-bg-color:#f8fafc; --el-table-row-hover-bg-color:#f3f7ff; }
.student-portal.uni-ui .sidebar { border-right:1px solid var(--uni-border); box-shadow:4px 0 24px rgba(15,23,42,.018); }
.student-portal.uni-ui .nav-item { transition:background-color .18s ease,color .18s ease,transform .18s ease; }
.student-portal.uni-ui .nav-item:focus-visible,.student-portal.uni-ui .mobile-nav button:focus-visible { outline:3px solid #93c5fd; outline-offset:2px; }
.student-portal.uni-ui .nav-item.active { background:#eaf1ff; color:var(--uni-primary); }
.student-portal.uni-ui .brand-icon { background:var(--uni-primary); }
.student-portal.uni-ui .topbar { border-bottom:1px solid var(--uni-border); }
.student-portal.uni-ui .content { max-width:1460px; }
.student-portal.uni-ui .welcome-card { background:linear-gradient(112deg,#183b9b 0%,#2563eb 65%,#4385f6 100%); border-radius:20px; box-shadow:0 14px 34px rgba(37,99,235,.15); }
.student-portal.uni-ui .welcome-card :deep(.el-button) { color:#1d4ed8; }
.student-portal.uni-ui .stat-card,.student-portal.uni-ui .section-card { border-color:var(--uni-border); border-radius:16px; box-shadow:0 5px 22px rgba(15,23,42,.035); transition:box-shadow .2s ease,transform .2s ease; }
@media(hover:hover) { .student-portal.uni-ui .stat-card:hover { transform:translateY(-2px); box-shadow:0 10px 28px rgba(15,23,42,.07); } }
.student-portal.uni-ui .stat-value { letter-spacing:-.7px; }
.student-portal.uni-ui .page-heading h1 { letter-spacing:-.6px; }
.student-portal.uni-ui .section-header h2 { color:#1e293b; }
.student-portal.uni-ui .profile-banner { background:linear-gradient(105deg,#edf4ff,#f8fbff); }
.student-portal.uni-ui .security-icon { background:#eaf1ff; color:var(--uni-primary); }
.student-portal.uni-ui .mobile-nav { border-color:var(--uni-border); }
.student-portal.uni-ui .mobile-nav button.active { color:var(--uni-primary); }
@media(max-width:760px) {
 .student-portal.uni-ui .welcome-card { border-radius:16px; }
 .student-portal.uni-ui .stat-card,.student-portal.uni-ui .section-card { border-radius:13px; }
 .student-portal.uni-ui .mobile-nav button { min-width:44px; min-height:44px; justify-content:center; }
 .student-portal.uni-ui .pagination { overflow-x:auto; }
}
@media(prefers-reduced-motion:reduce) { .student-portal.uni-ui .nav-item,.student-portal.uni-ui .stat-card { transition:none; transform:none; } }
</style>

<style scoped>
/* V17.7.2: mobile usability without changing student portal scroll rules */
.student-portal.uni-ui .main-container { min-width: 0; }
.student-portal.uni-ui .content { min-width: 0; }
.student-portal.uni-ui .section-card { min-width: 0; }
.student-portal.uni-ui .section-card :deep(.el-table) { max-width: 100%; }
.student-portal.uni-ui .filter-actions { flex-wrap: wrap; }
.student-portal.uni-ui .password-form :deep(.el-input) { max-width: 100%; }
@media (max-width: 760px) {
  .student-portal.uni-ui .page-heading,
  .student-portal.uni-ui .section-header { flex-wrap: wrap; }
  .student-portal.uni-ui .filter-actions { gap: 8px; }
  .student-portal.uni-ui .welcome-card { min-width: 0; }
  .student-portal.uni-ui .stat-card { min-width: 0; }
  .student-portal.uni-ui .stat-value { overflow-wrap: anywhere; }
  .student-portal.uni-ui .pagination { justify-content: center; }
}
</style>

<style scoped>
.student-portal {
  --student-surface: #fff;
  --student-ink: #17233c;
  --student-muted: #52647d;
  --student-line: #e3eaf4;
}
.student-portal.uni-ui .stats-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
.student-portal.uni-ui .stat-label { color: var(--student-muted); }
.leave-card {
  width: 100%;
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: center;
  gap: 10px 14px;
  min-height: 70px;
  padding: 14px 0;
  border: 0;
  border-top: 1px solid var(--student-line);
  background: transparent;
  color: var(--student-ink);
  text-align: left;
  cursor: pointer;
}
.leave-card > span:last-child { grid-column: 1 / -1; color: var(--student-muted); font-size: 13px; }
.leave-card:focus-visible { outline: 2px solid #5384ef; outline-offset: 2px; border-radius: 6px; }
.record-filters { display: flex; gap: 12px; flex-wrap: wrap; margin-bottom: 16px; }
.record-filters .el-select { width: 150px; }
.record-filters label { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.record-filters input { min-height: 34px; color: var(--student-ink); background: var(--student-surface); border: 1px solid var(--student-line); border-radius: 7px; padding: 4px 8px; }
.date-error { color: var(--el-color-danger); margin-bottom: 12px; }
:global(html.dark) .record-filters input { color-scheme: dark; }
.application-advisor { display: flex; justify-content: space-between; padding: 12px 14px; border-radius: 10px; margin-bottom: 16px; background: var(--el-fill-color-light); }
.application-duration { margin: -4px 0 16px; display: flex; justify-content: space-between; }
.leave-details { display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 12px 16px; }
.leave-details dt { color: var(--student-muted); }
.leave-details dd { margin: 0; white-space: pre-wrap; overflow-wrap: anywhere; }
.profile-actions { display: flex; flex-wrap: wrap; gap: 10px; padding: 0 24px 24px; }
.profile-actions :deep(.el-button) { margin-left: 0; }
.profile-loading { max-width: 560px; margin: 0 auto; padding: 36px 18px; }

@media (max-width: 760px) {
  .student-portal.uni-ui .stats-grid { gap: 8px; }
  .student-portal.uni-ui .stat-card { padding: 14px 6px; justify-content: center; text-align: center; }
  .student-portal .stat-icon { display: none; }
  .student-portal .welcome-decoration { display: none; }
  .student-portal.uni-ui .mobile-nav { padding-bottom: max(6px, env(safe-area-inset-bottom)); }
  .student-portal.uni-ui .content { padding-bottom: calc(84px + env(safe-area-inset-bottom)); }
  .record-filters { flex-direction: column; }
  .record-filters .el-select, .record-filters input { width: 100%; }
  .record-filters label { display: grid; grid-template-columns: 64px minmax(0, 1fr); }
}

:global(html.dark) .student-portal {
  --student-surface: #152033;
  --student-ink: #e6edf8;
  --student-muted: #aebdd2;
  --student-line: #2a3951;
  background: #101827;
  color: var(--student-ink);
}
:global(html.dark) .student-portal :is(.sidebar, .topbar, .section-card, .stat-card, .mobile-nav, .profile-banner) {
  background: var(--student-surface);
  border-color: var(--student-line);
}
:global(html.dark) .student-portal :is(.section-header h2, .topbar-left strong, .nav-item, .mobile-nav button, .page-heading p, .muted) {
  color: var(--student-ink);
}
</style>
