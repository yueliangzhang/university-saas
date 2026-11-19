<template>
  <div class="student-portal">
    <template v-if="!token">
      <div class="login-layout">
        <section class="login-intro">
          <div class="brand-mark"><el-icon :size="24"><School /></el-icon></div>
          <div class="eyebrow">CAMPUS STUDENT SERVICES</div>
          <h1>校园服务，<br />从这里开始。</h1>
          <p>统一的学生服务入口，随时查看个人信息、提交请假申请、追踪审批进度。</p>
          <div class="intro-features">
            <span><el-icon><CircleCheck /></el-icon> 便捷申请</span>
            <span><el-icon><Clock /></el-icon> 进度透明</span>
            <span><el-icon><Lock /></el-icon> 安全访问</span>
          </div>
        </section>
        <section class="login-panel">
          <div class="login-card">
            <div class="mobile-brand"><el-icon><School /></el-icon> 高校学生服务平台</div>
            <div class="login-title">欢迎回来</div>
            <p class="muted">请使用学校管理员为你开通的学生账号登录</p>
            <el-form label-position="top" @submit.prevent="login">
              <el-form-item label="高校 ID"><el-input v-model="form.tenantId" size="large" placeholder="请输入高校 ID，例如 1" /></el-form-item>
              <el-form-item label="学号"><el-input v-model="form.studentNo" size="large" placeholder="请输入学号" autocomplete="username" /></el-form-item>
              <el-form-item label="登录密码"><el-input v-model="form.password" size="large" type="password" show-password autocomplete="current-password" placeholder="请输入密码" /></el-form-item>
              <el-button class="login-button" type="primary" size="large" native-type="submit" :loading="loading">登录学生平台 <el-icon class="ml-8"><Right /></el-icon></el-button>
            </el-form>
            <p class="login-help">忘记密码？请联系本校管理员重置。</p>
          </div>
        </section>
      </div>
    </template>

    <template v-else>
      <el-container class="app-shell">
        <el-aside class="sidebar" width="236px">
          <div class="sidebar-brand"><span class="brand-icon"><el-icon><School /></el-icon></span><span>学生服务平台<small>STUDENT PORTAL</small></span></div>
          <div class="nav-label">工作台</div>
          <button v-for="item in navigation" :key="item.key" class="nav-item" :class="{active:activeTab===item.key}" @click="activeTab=item.key">
            <el-icon :size="18"><component :is="item.icon" /></el-icon><span>{{item.label}}</span>
          </button>
          <div class="sidebar-bottom">
            <div class="sidebar-user"><el-avatar :size="36">{{ (me?.name || '学').slice(0,1) }}</el-avatar><div><strong>{{me?.name||'学生'}}</strong><small>{{me?.studentNo||'—'}}</small></div></div>
            <el-button class="logout-button" text @click="logout"><el-icon><SwitchButton /></el-icon> 退出登录</el-button>
          </div>
        </el-aside>

        <el-container class="main-container">
          <el-header class="topbar">
            <div class="topbar-left"><el-icon class="topbar-icon"><School /></el-icon><span>高校学生服务平台</span><span class="topbar-separator">/</span><strong>{{ currentTabLabel }}</strong></div>
            <div class="topbar-right"><el-tag type="success" effect="plain" round>已登录</el-tag><el-avatar :size="32">{{ (me?.name || '学').slice(0,1) }}</el-avatar></div>
          </el-header>
          <el-main class="content">
            <template v-if="activeTab==='dashboard'">
              <div class="welcome-card">
                <div><div class="welcome-eyebrow">STUDENT DASHBOARD</div><h1>你好，{{me?.name||'同学'}} <span class="wave">👋</span></h1><p>欢迎回来。你可以在这里管理请假申请，及时了解审批进度。</p><el-button type="primary" size="large" @click="openLeaveDialog"><el-icon><Plus /></el-icon> 发起请假申请</el-button></div>
                <div class="welcome-decoration"><el-icon :size="90"><School /></el-icon></div>
              </div>
              <div class="stats-grid">
                <div v-for="stat in stats" :key="stat.key" class="stat-card"><div class="stat-icon" :class="stat.key"><el-icon :size="22"><component :is="stat.icon" /></el-icon></div><div><div class="stat-label">{{stat.label}}</div><div class="stat-value">{{stat.value}}</div></div></div>
              </div>
              <div class="section-card">
                <div class="section-header"><div><h2>最近请假记录</h2><p>随时了解申请的最新审批状态</p></div><el-button text type="primary" @click="activeTab='leaves'">查看全部 <el-icon><Right /></el-icon></el-button></div>
                <el-table v-loading="leaveLoading" :data="leaveRows.slice(0,5)" class="data-table" empty-text="暂无请假记录，点击上方按钮提交第一条申请">
                  <el-table-column label="请假类型" min-width="95"><template #default="{row}"><strong>{{leaveTypes[row.leaveType]||row.leaveType}}</strong></template></el-table-column>
                  <el-table-column label="请假时间" min-width="230"><template #default="{row}">{{leaveFmt(row.startAt)}} 至 {{leaveFmt(row.endAt)}}</template></el-table-column>
                  <el-table-column prop="reason" label="请假原因" min-width="130" show-overflow-tooltip />
                  <el-table-column label="状态" width="108"><template #default="{row}"><el-tag :type="statusType(row.status)" effect="light" round>{{leaveStatuses[row.status]||row.status}}</el-tag></template></el-table-column>
                </el-table>
              </div>
            </template>

            <template v-else-if="activeTab==='leaves'">
              <div class="page-heading"><div><h1>我的请假</h1><p>管理请假申请，查看审批结果和历史记录</p></div><el-button type="primary" @click="openLeaveDialog"><el-icon><Plus /></el-icon> 申请请假</el-button></div>
              <div class="stats-grid compact-stats"><div v-for="stat in stats" :key="stat.key" class="stat-card"><div class="stat-icon" :class="stat.key"><el-icon :size="20"><component :is="stat.icon" /></el-icon></div><div><div class="stat-label">{{stat.label}}</div><div class="stat-value">{{stat.value}}</div></div></div></div>
              <div class="section-card">
                <div class="section-header filters"><div><h2>申请记录</h2><p>共 {{leaveTotal}} 条记录</p></div><div class="filter-actions"><el-select v-model="statusFilter" placeholder="全部状态" style="width:136px" @change="leavePage=1"><el-option label="全部状态" value="all"/><el-option v-for="(label,key) in leaveStatuses" :key="key" :label="label" :value="key"/></el-select><el-button :loading="leaveLoading" @click="refreshLeaves"><el-icon><Refresh /></el-icon> 刷新</el-button></div></div>
                <el-table v-loading="leaveLoading" :data="filteredLeaveRows" class="data-table" empty-text="当前没有符合条件的请假记录">
                  <el-table-column label="类型" width="90"><template #default="{row}"><strong>{{leaveTypes[row.leaveType]||row.leaveType}}</strong></template></el-table-column>
                  <el-table-column label="开始时间" min-width="150"><template #default="{row}">{{leaveFmt(row.startAt)}}</template></el-table-column>
                  <el-table-column label="结束时间" min-width="150"><template #default="{row}">{{leaveFmt(row.endAt)}}</template></el-table-column>
                  <el-table-column prop="reason" label="请假原因" min-width="170" show-overflow-tooltip />
                  <el-table-column label="审批状态" width="105"><template #default="{row}"><el-tag :type="statusType(row.status)" round>{{leaveStatuses[row.status]||row.status}}</el-tag></template></el-table-column>
                  <el-table-column prop="reviewNote" label="审批备注" min-width="150" show-overflow-tooltip><template #default="{row}">{{row.reviewNote||'—'}}</template></el-table-column>
                  <el-table-column label="操作" width="85" fixed="right"><template #default="{row}"><el-button v-if="row.status==='pending'" type="danger" link @click="cancelLeave(row)">撤销</el-button><span v-else class="muted">—</span></template></el-table-column>
                </el-table>
                <div class="pagination"><el-pagination v-model:current-page="leavePage" :page-size="10" :total="leaveTotal" layout="total, prev, pager, next" @current-change="loadLeaves" /></div>
              </div>
              <p v-if="statusFilter!=='all'" class="filter-note">当前状态筛选仅作用于本页记录；如需查看其他记录，请翻页。</p>
            </template>

            <template v-else-if="activeTab==='profile'">
              <div class="page-heading"><div><h1>个人中心</h1><p>查看你的在校基本信息</p></div></div>
              <div class="section-card profile-card"><div class="profile-banner"><el-avatar :size="66">{{ (me?.name||'学').slice(0,1) }}</el-avatar><div><h2>{{me?.name||'学生'}}</h2><p>学号 {{me?.studentNo||'—'}}</p></div></div><el-descriptions :column="profileColumns" border class="profile-details"><el-descriptions-item label="姓名">{{me?.name||'—'}}</el-descriptions-item><el-descriptions-item label="学号">{{me?.studentNo||'—'}}</el-descriptions-item><el-descriptions-item label="班级 ID">{{me?.classId??'—'}}</el-descriptions-item><el-descriptions-item label="入学年份">{{me?.enrollmentYear||'—'}}</el-descriptions-item></el-descriptions><p class="muted profile-hint">如需修改学籍信息，请联系学校管理员。</p></div>
            </template>

            <template v-else>
              <div class="page-heading"><div><h1>账号安全</h1><p>管理登录密码，保护你的学生账号</p></div></div>
              <div class="section-card security-card"><div class="security-heading"><span class="security-icon"><el-icon :size="23"><Lock /></el-icon></span><div><h2>修改登录密码</h2><p>建议使用至少 12 位的强密码，并避免与其他网站重复。</p></div></div><el-form label-position="top" class="password-form" @submit.prevent="changePassword"><el-form-item label="当前密码"><el-input v-model="oldPassword" type="password" show-password autocomplete="current-password" placeholder="请输入当前密码" /></el-form-item><el-form-item label="新密码"><el-input v-model="newPassword" type="password" show-password autocomplete="new-password" placeholder="至少 12 位，最多 72 字节" /></el-form-item><el-button type="primary" native-type="submit" :loading="passwordSaving">保存新密码</el-button></el-form><el-alert class="password-tip" title="密码修改成功后，当前会话将自动退出，请使用新密码重新登录。" type="info" :closable="false" show-icon /></div>
            </template>
          </el-main>
        </el-container>
      </el-container>

      <nav class="mobile-nav" aria-label="学生服务导航"><button v-for="item in navigation" :key="item.key" :class="{active:activeTab===item.key}" @click="activeTab=item.key"><el-icon :size="20"><component :is="item.icon" /></el-icon><span>{{item.shortLabel}}</span></button><button @click="logout"><el-icon :size="20"><SwitchButton /></el-icon><span>退出</span></button></nav>

      <el-dialog v-model="leaveDialog" title="申请请假" class="leave-dialog" width="min(560px, 94vw)" append-to-body destroy-on-close>
        <el-alert title="请如实填写请假信息，提交后由学校管理员审批。" type="info" :closable="false" class="dialog-alert" />
        <el-form label-position="top" @submit.prevent="submitLeave">
          <el-form-item label="请假类型" required><el-select v-model="leaveForm.leaveType" style="width:100%"><el-option label="事假" value="personal"/><el-option label="病假" value="sick"/><el-option label="其他" value="other"/></el-select></el-form-item>
          <el-form-item label="开始时间" required><el-date-picker v-model="leaveForm.startAt" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width:100%" placeholder="请选择开始时间"/></el-form-item>
          <el-form-item label="结束时间" required><el-date-picker v-model="leaveForm.endAt" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width:100%" placeholder="请选择结束时间"/></el-form-item>
          <el-form-item label="请假原因" required><el-input v-model="leaveForm.reason" type="textarea" :rows="4" maxlength="2000" show-word-limit placeholder="请简要说明请假原因"/></el-form-item>
        </el-form>
        <template #footer><el-button @click="leaveDialog=false">取消</el-button><el-button type="primary" :loading="leaveSaving" @click="submitLeave">提交申请</el-button></template>
      </el-dialog>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { School, CircleCheck, Clock, Lock, Right, Plus, House, Calendar, User, SwitchButton, Refresh, Files, CircleClose } from '@element-plus/icons-vue'
import axios from 'axios'

const form = ref({ tenantId: '', studentNo: '', password: '' })
const token = ref(sessionStorage.getItem('uni_student_token') || '')
const me = ref(null)
const loading = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const passwordSaving = ref(false)
const activeTab = ref('dashboard')
const navigation = [
  { key: 'dashboard', label: '首页概览', shortLabel: '首页', icon: House },
  { key: 'leaves', label: '我的请假', shortLabel: '请假', icon: Calendar },
  { key: 'profile', label: '个人中心', shortLabel: '我的', icon: User },
  { key: 'security', label: '账号安全', shortLabel: '安全', icon: Lock },
]
const currentTabLabel = computed(() => navigation.find(n => n.key === activeTab.value)?.label || '首页概览')
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
const loadMe = async () => {
  try { me.value = await call('/uni/student-portal/me') }
  catch (e) { clear(); ElMessage.error(leaveError(e)) }
}
const login = async () => {
  if (!form.value.tenantId || !form.value.studentNo.trim() || !form.value.password) { ElMessage.warning('请填写高校 ID、学号和密码'); return }
  loading.value = true
  try {
    const data = await call('/uni/student-portal/login', 'post', { tenantId: Number(form.value.tenantId), studentNo: form.value.studentNo.trim(), password: form.value.password })
    token.value = data.token
    sessionStorage.setItem('uni_student_token', data.token)
    form.value.password = ''
    await loadMe()
    if (token.value) await loadLeaves()
  } catch (e) { ElMessage.error(leaveError(e)) }
  finally { loading.value = false }
}
const logout = async () => {
  try { await call('/uni/student-portal/logout', 'post', {}) }
  catch (e) { /* local session still cleared */ }
  finally { clear() }
}
const changePassword = async () => {
  if (!oldPassword.value || new TextEncoder().encode(newPassword.value).length < 12 || new TextEncoder().encode(newPassword.value).length > 72) {
    ElMessage.warning('请输入当前密码，新密码长度应为 12–72 字节'); return
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
const leaveForm = ref({ leaveType: 'personal', reason: '', startAt: '', endAt: '' })
const statusFilter = ref('all')
const filteredLeaveRows = computed(() => statusFilter.value === 'all' ? leaveRows.value : leaveRows.value.filter(r => r.status === statusFilter.value))
const stats = computed(() => [
  { key: 'total', label: '请假总数', value: leaveTotal.value, icon: Files },
  { key: 'pending', label: '本页待审批', value: leaveRows.value.filter(r => r.status === 'pending').length, icon: Clock },
  { key: 'approved', label: '本页已通过', value: leaveRows.value.filter(r => r.status === 'approved').length, icon: CircleCheck },
  { key: 'rejected', label: '本页已驳回', value: leaveRows.value.filter(r => r.status === 'rejected').length, icon: CircleClose },
])
const statusType = status => ({ pending: 'warning', approved: 'success', rejected: 'danger', cancelled: 'info' })[status] || 'info'
const leaveFmt = v => v ? String(v).replace('T', ' ').slice(0, 16) : ''
const loadLeaves = async () => {
  if (!token.value) return
  leaveLoading.value = true
  try {
    const data = await call('/uni/student-portal/leaves', 'get', undefined, { page: leavePage.value, pageSize: 10 })
    leaveRows.value = data?.list || []
    leaveTotal.value = data?.total || 0
  } catch (e) { ElMessage.error(leaveError(e)) }
  finally { leaveLoading.value = false }
}
const refreshLeaves = () => { loadLeaves() }
const openLeaveDialog = () => { leaveForm.value = { leaveType: 'personal', reason: '', startAt: '', endAt: '' }; leaveDialog.value = true }
const submitLeave = async () => {
  const f = leaveForm.value
  if (!f.reason?.trim() || !f.startAt || !f.endAt || f.endAt <= f.startAt) { ElMessage.warning('请填写完整信息，结束时间必须晚于开始时间'); return }
  leaveSaving.value = true
  try {
    await call('/uni/student-portal/leaves', 'post', { leaveType: f.leaveType, reason: f.reason.trim(), startAt: f.startAt, endAt: f.endAt })
    ElMessage.success('请假申请已提交')
    leaveDialog.value = false
    leavePage.value = 1
    await loadLeaves()
  } catch (e) { ElMessage.error(leaveError(e)) }
  finally { leaveSaving.value = false }
}
const cancelLeave = async row => {
  try {
    await ElMessageBox.confirm('确认撤销这条待审批请假申请？', '撤销确认', { type: 'warning', confirmButtonText: '确认撤销', cancelButtonText: '取消' })
    await call('/uni/student-portal/leaves/cancel', 'post', { id: row.id })
    ElMessage.success('已撤销')
    await loadLeaves()
  } catch (e) { if (e !== 'cancel' && e !== 'close') ElMessage.error(leaveError(e)) }
}
onMounted(async () => { window.addEventListener('resize', onResize); if (token.value) { await loadMe(); if (token.value) await loadLeaves() } })
onUnmounted(() => window.removeEventListener('resize', onResize))
</script>

<style scoped>
.student-portal{min-height:100vh;background:#f5f7fb;color:#17233c;font-family:Inter,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}
.student-portal :deep(.el-button){font-weight:550}.student-portal :deep(.el-button--primary){--el-button-bg-color:#3266e8;--el-button-border-color:#3266e8;--el-button-hover-bg-color:#2457d5;--el-button-hover-border-color:#2457d5}
.app-shell{min-height:100vh}.sidebar{background:#fff;border-right:1px solid #e9edf5;display:flex;flex-direction:column;min-height:100vh;padding:27px 14px 20px;position:sticky;top:0;height:100vh}.sidebar-brand{display:flex;align-items:center;gap:11px;font-size:17px;font-weight:750;padding:0 10px 36px}.sidebar-brand small{display:block;font-size:10px;font-weight:650;letter-spacing:1.5px;color:#94a1b6;margin-top:3px}.brand-icon,.brand-mark{background:#3266e8;color:#fff;border-radius:12px;display:flex;align-items:center;justify-content:center}.brand-icon{width:39px;height:39px;font-size:21px}.nav-label{padding:0 16px 12px;color:#9aa6b7;font-size:11px;letter-spacing:1px;font-weight:700}.nav-item{width:100%;display:flex;align-items:center;gap:13px;background:transparent;border:0;border-radius:10px;padding:13px 15px;margin-bottom:6px;text-align:left;color:#66758d;cursor:pointer;font-size:14px;font-weight:550}.nav-item:hover{background:#f4f7ff;color:#3266e8}.nav-item.active{background:#edf3ff;color:#3266e8;font-weight:700}.sidebar-bottom{margin-top:auto;border-top:1px solid #eef1f6;padding:20px 8px 0}.sidebar-user{display:flex;gap:10px;align-items:center;margin-bottom:16px}.sidebar-user :deep(.el-avatar),.topbar-right :deep(.el-avatar),.profile-banner :deep(.el-avatar){background:#dbe7ff;color:#3266e8;font-weight:750}.sidebar-user strong,.sidebar-user small{display:block}.sidebar-user strong{font-size:13px}.sidebar-user small{font-size:12px;color:#95a1b3;margin-top:3px}.logout-button{color:#758298!important}.main-container{min-width:0}.topbar{height:70px;background:#fff;border-bottom:1px solid #e9edf5;display:flex;justify-content:space-between;align-items:center;padding:0 36px}.topbar-left,.topbar-right{display:flex;align-items:center;gap:14px;font-size:13px;color:#8a98ad}.topbar-left strong{color:#34415a;font-weight:650}.topbar-icon{color:#3266e8;font-size:18px}.topbar-separator{color:#ccd4df}.content{max-width:1400px;width:100%;margin:0 auto;padding:32px 36px 80px}.welcome-card{background:linear-gradient(110deg,#244ed0 0%,#3875f1 100%);color:white;border-radius:18px;min-height:212px;padding:34px 40px;display:flex;justify-content:space-between;align-items:center;overflow:hidden;position:relative}.welcome-eyebrow{font-size:11px;letter-spacing:2px;opacity:.75;font-weight:700}.welcome-card h1{font-size:29px;letter-spacing:-.5px;margin:13px 0 9px}.welcome-card p{font-size:13px;color:#dbe7ff;margin:0 0 23px;line-height:1.7}.welcome-card :deep(.el-button){background:#fff;border-color:#fff;color:#2455d4}.welcome-decoration{opacity:.16;transform:rotate(-13deg);padding:0 5%}.wave{font-size:24px}.stats-grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:18px;margin:22px 0}.stat-card{background:#fff;border:1px solid #e9edf5;border-radius:14px;padding:23px 18px;display:flex;align-items:center;gap:15px;min-width:0}.stat-icon{width:48px;height:48px;flex:none;border-radius:12px;display:flex;align-items:center;justify-content:center;background:#edf3ff;color:#3972e9}.stat-icon.pending{background:#fff5e7;color:#f0a43d}.stat-icon.approved{background:#e9f9f1;color:#28b47a}.stat-icon.rejected{background:#fff0f0;color:#e96a72}.stat-label{font-size:12px;color:#8895a9}.stat-value{font-size:26px;font-weight:750;margin-top:6px;line-height:1}.section-card{background:#fff;border:1px solid #e9edf5;border-radius:15px;padding:24px;margin-top:22px}.section-header{display:flex;align-items:center;justify-content:space-between;gap:15px;margin-bottom:18px}.section-header h2,.security-heading h2{font-size:17px;margin:0 0 7px}.section-header p,.security-heading p{color:#9aa5b6;font-size:12px;margin:0}.data-table{width:100%;--el-table-header-bg-color:#f8faff;--el-table-border-color:#edf1f6;--el-table-row-hover-bg-color:#f7faff}.data-table :deep(th.el-table__cell){color:#7e8ca3;font-weight:600}.page-heading{display:flex;justify-content:space-between;align-items:center;gap:16px;margin-bottom:22px}.page-heading h1{font-size:25px;margin:0 0 9px}.page-heading p{color:#8d9aaf;font-size:13px;margin:0}.compact-stats{margin-top:0}.filter-actions{display:flex;gap:10px}.pagination{display:flex;justify-content:flex-end;margin-top:20px}.filter-note{font-size:12px;color:#8b98ab}.muted{color:#9aa6b6}.profile-card{padding:0;overflow:hidden}.profile-banner{background:linear-gradient(100deg,#f0f5ff,#f9fbff);padding:35px;display:flex;align-items:center;gap:18px}.profile-banner h2{margin:0 0 7px}.profile-banner p{color:#8a98aa;margin:0}.profile-details{margin:30px}.profile-hint{margin:0 30px 30px;font-size:12px}.security-card{max-width:680px}.security-heading{display:flex;align-items:center;gap:16px;margin-bottom:28px}.security-icon{width:48px;height:48px;border-radius:12px;background:#edf3ff;color:#3266e8;display:flex;align-items:center;justify-content:center}.password-form{max-width:440px}.password-tip{margin-top:25px}.mobile-nav{display:none}
.login-layout{display:grid;grid-template-columns:1fr 1fr;min-height:100vh}.login-intro{background:linear-gradient(150deg,#1d43bc,#3674ed);padding:10vh 9%;color:white;display:flex;flex-direction:column;justify-content:center}.brand-mark{width:58px;height:58px;margin-bottom:50px}.eyebrow{font-size:12px;letter-spacing:2.5px;color:#bed3ff;font-weight:700}.login-intro h1{font-size:clamp(36px,4.5vw,62px);line-height:1.2;margin:22px 0}.login-intro p{max-width:410px;color:#d6e3ff;font-size:15px;line-height:1.9}.intro-features{display:flex;flex-wrap:wrap;gap:22px;margin-top:46px}.intro-features span{display:flex;align-items:center;gap:7px;font-size:13px;color:#e2eaff}.login-panel{display:flex;align-items:center;justify-content:center;padding:30px;background:#f8faff}.login-card{width:min(100%,420px);background:white;padding:48px 40px;border-radius:18px;box-shadow:0 16px 60px rgba(33,65,120,.08)}.login-title{font-size:28px;font-weight:750;margin-bottom:12px}.login-card .muted{font-size:13px;line-height:1.7;margin:0 0 32px}.login-card :deep(.el-form-item){margin-bottom:23px}.login-button{width:100%;margin-top:8px}.ml-8{margin-left:8px}.login-help{text-align:center;color:#98a5b6;font-size:12px;margin-top:26px}.mobile-brand{display:none}
@media(max-width:1100px){.stats-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.content{padding:26px 24px 80px}.topbar{padding:0 24px}.sidebar{width:205px!important}.welcome-decoration{display:none}}
@media(max-width:760px){.sidebar{display:none}.topbar{height:58px;padding:0 17px}.topbar-left{gap:8px;font-size:12px}.topbar-right{gap:7px}.topbar-icon,.topbar-separator{display:none}.content{padding:18px 14px 90px}.welcome-card{padding:25px 22px;min-height:185px}.welcome-card h1{font-size:24px}.stats-grid{gap:10px;margin:14px 0}.stat-card{padding:15px 12px;gap:10px}.stat-icon{width:37px;height:37px}.stat-value{font-size:22px}.stat-label{font-size:11px}.section-card{padding:15px;margin-top:14px}.section-header{align-items:flex-start}.page-heading h1{font-size:22px}.page-heading{align-items:flex-start}.filters{flex-direction:column}.filter-actions{width:100%;justify-content:space-between}.mobile-nav{display:flex;position:fixed;bottom:0;left:0;right:0;background:#fff;border-top:1px solid #e9edf5;z-index:100;padding:8px 4px max(8px,env(safe-area-inset-bottom));justify-content:space-around;box-shadow:0 -3px 16px rgba(32,50,85,.04)}.mobile-nav button{border:0;background:transparent;display:flex;flex-direction:column;align-items:center;gap:4px;color:#8996a9;font-size:10px;padding:4px 6px}.mobile-nav button.active{color:#3266e8;font-weight:700}.login-layout{grid-template-columns:1fr}.login-intro{display:none}.login-panel{min-height:100vh;padding:18px}.login-card{padding:35px 25px}.mobile-brand{display:flex;align-items:center;gap:8px;color:#3266e8;font-weight:750;margin-bottom:40px}.profile-banner{padding:22px}.profile-details{margin:18px}.profile-hint{margin:0 18px 22px}}
</style>
