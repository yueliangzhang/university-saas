<template>
  <el-badge :value="unread" :max="99" :hidden="unread === 0" class="notification-badge">
    <el-button :icon="Bell" circle :aria-label="`通知，${unread} 条未读`" @click="showInbox" />
  </el-badge>
  <el-drawer v-model="visible" title="站内通知" size="min(100%, 420px)">
    <div class="inbox-actions">
      <span>{{ unread }} 条未读</span>
      <el-button link type="primary" :disabled="!unread || busy" @click="markAll">全部已读</el-button>
    </div>
    <el-alert v-if="loadError" :title="loadError" type="error" :closable="false">
      <el-button link @click="load">重试</el-button>
    </el-alert>
    <el-empty v-else-if="!items.length && !loading" description="暂无通知" :image-size="64" />
    <ul v-loading="loading" class="inbox-list">
      <li v-for="item in items" :key="item.id">
        <button type="button" :disabled="busy" @click="openItem(item)">
          <strong>{{ labels[item.event] }}</strong>
          <span v-if="!item.readAt" class="unread-label">未读</span>
          <small>申请 #{{ item.leaveId }} · <time :datetime="item.createdAt">{{ formatTime(item.createdAt) }}</time></small>
        </button>
        <el-button v-if="!item.readAt" link :disabled="busy" @click="markOne(item.id)">标为已读</el-button>
      </li>
    </ul>
    <el-pagination v-if="total > pageSize" :current-page="page" :page-size="pageSize"
      :total="total" layout="prev, pager, next" @current-change="changePage" />
  </el-drawer>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { Bell } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  call: { type: Function, required: true },
  openLeave: { type: Function, required: true }
})
const labels = {
  submitted: '有新的请假申请待审批',
  approved: '你的请假申请已通过',
  rejected: '你的请假申请已驳回',
  cancelled: '有一条请假申请已撤销'
}
const visible = ref(false)
const unread = ref(0)
const items = ref([])
const page = ref(1)
const pageSize = ref(15)
const total = ref(0)
const loading = ref(false)
const busy = ref(false)
const loadError = ref('')
let timer
let sequence = 0

const formatTime = value => new Date(value).toLocaleString()
const errorMessage = error => error.response?.data?.msg || error.message || '操作失败'

async function refreshBadge() {
  if (document.hidden || visible.value) return
  const request = ++sequence
  try {
    const result = await props.call('/notifications', 'get', undefined, { unreadOnly: true })
    if (request === sequence) unread.value = result.unread
  } catch {
    // Opening the inbox shows the error and allows retrying.
  }
}

async function load() {
  const request = ++sequence
  loading.value = true
  loadError.value = ''
  try {
    const result = await props.call('/notifications', 'get', undefined, { page: page.value })
    if (request !== sequence) return
    items.value = result.list
    unread.value = result.unread
    total.value = result.total
    pageSize.value = result.pageSize
  } catch (error) {
    if (request === sequence) loadError.value = errorMessage(error)
  } finally {
    if (request === sequence) loading.value = false
  }
}

function showInbox() {
  page.value = 1
  visible.value = true
  load()
}

function changePage(value) {
  page.value = value
  load()
}

async function markRead(payload) {
  await props.call('/notifications/read', 'post', payload)
}

async function mark(payload) {
  busy.value = true
  try {
    await markRead(payload)
    await load()
  } catch (error) {
    ElMessage.error(errorMessage(error))
  } finally {
    busy.value = false
  }
}

const markAll = () => mark({ all: true })
const markOne = id => mark({ id })

async function openItem(item) {
  busy.value = true
  try {
    if (!item.readAt) await markRead({ id: item.id })
    await props.openLeave(item.leaveId)
    visible.value = false
    await refreshBadge()
  } catch (error) {
    ElMessage.error(errorMessage(error))
    await load()
  } finally {
    busy.value = false
  }
}

onMounted(() => {
  refreshBadge()
  timer = window.setInterval(refreshBadge, 60000)
  document.addEventListener('visibilitychange', refreshBadge)
})
onUnmounted(() => {
  sequence++
  window.clearInterval(timer)
  document.removeEventListener('visibilitychange', refreshBadge)
})
</script>

<style scoped>
.notification-badge { margin-inline-end: 8px; }
.inbox-actions { display: flex; align-items: center; justify-content: space-between; }
.inbox-list { list-style: none; padding: 0; }
.inbox-list li { display: flex; align-items: center; gap: 12px; border-bottom: 1px solid var(--el-border-color-lighter); }
.inbox-list li > button:first-child { flex: 1; padding: 12px 0; text-align: start; background: none; border: 0; color: inherit; cursor: pointer; }
.inbox-list li > button:first-child:hover strong { color: var(--el-color-primary); }
.inbox-list small { display: block; margin-top: 4px; font-size: 12px; color: var(--el-text-color-secondary); }
.unread-label { margin-inline-start: 8px; font-size: 12px; color: var(--el-color-primary); }
</style>
