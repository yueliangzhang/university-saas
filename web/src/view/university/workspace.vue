<template>
  <!-- UNIVERSITY_WORKSPACE_V17_6_5: single root for GVA Transition -->
  <main class="uni-workspace-simple">
    <header class="uni-workspace-simple__header">
      <div>
        <h1>高校工作台</h1>
        <p>校园数据概览与日常管理</p>
      </div>
      <el-button :loading="loading" @click="refreshContext">刷新身份信息</el-button>
    </header>

    <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" class="uni-workspace-simple__alert" />

    <section class="uni-workspace-simple__overview" aria-label="校园运营统计">
      <UniAdminOverview />
    </section>

    <section class="uni-workspace-simple__info" aria-label="当前高校身份">
      <div class="uni-workspace-simple__info-title">当前高校</div>
      <div v-loading="loading" class="uni-workspace-simple__info-content">
        <template v-if="context">
          <span>高校 ID：<strong>{{ context.tenantId ?? '—' }}</strong></span>
          <span>成员角色：<strong>{{ context.role || '—' }}</strong></span>
        </template>
        <span v-else-if="!loading" class="uni-workspace-simple__muted">暂无身份信息</span>
      </div>
    </section>
  </main>
</template>

<script setup>
import UniAdminOverview from '@/components/university-ui/UniAdminOverview.vue'
import { ref, onMounted } from 'vue'
import request from '@/utils/request'

const loading = ref(false)
const context = ref(null)
const error = ref('')

const refreshContext = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await request({ url: '/uni/workspace/context', method: 'get' })
    if (res?.code === 0) context.value = res.data
    else error.value = res?.msg || '获取高校身份失败'
  } catch (e) {
    error.value = '获取高校身份失败，请检查登录状态和权限'
  } finally {
    loading.value = false
  }
}

onMounted(refreshContext)
</script>

<style scoped>
.uni-workspace-simple {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px 24px 28px;
  min-width: 0;
  color: var(--el-text-color-primary, #243247);
}
.uni-workspace-simple__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}
.uni-workspace-simple__header h1 {
  margin: 0;
  font-size: 22px;
  line-height: 1.4;
  font-weight: 650;
}
.uni-workspace-simple__header p {
  margin: 5px 0 0;
  font-size: 13px;
  color: var(--el-text-color-secondary, #64748b);
}
.uni-workspace-simple__alert { margin: 0; }
.uni-workspace-simple__overview { min-width: 0; }
/* Keep the existing live-statistics component, but remove its large outer frame. */
.uni-workspace-simple__overview :deep(.uni-admin-overview) {
  margin: 0;
}
.uni-workspace-simple__info {
  background: var(--el-bg-color, #fff);
  border: 1px solid var(--el-border-color-lighter, #e6ebf3);
  border-radius: 12px;
  padding: 16px 18px;
}
.uni-workspace-simple__info-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
}
.uni-workspace-simple__info-content {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px 28px;
  min-height: 24px;
  font-size: 13px;
  color: var(--el-text-color-secondary, #64748b);
}
.uni-workspace-simple__info-content strong {
  font-weight: 600;
  color: var(--el-text-color-primary, #243247);
}
.uni-workspace-simple__muted { color: var(--el-text-color-placeholder, #94a3b8); }
@media (max-width: 768px) {
  .uni-workspace-simple { padding: 12px; gap: 12px; }
  .uni-workspace-simple__header h1 { font-size: 20px; }
  .uni-workspace-simple__info { padding: 14px; }
}
</style>
