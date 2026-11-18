<template>
  <div class="p-6">
    <el-card>
      <template #header>高校工作台</template>
      <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" />
      <div v-loading="loading">
        <el-descriptions v-if="context" :column="1" border>
          <el-descriptions-item label="当前高校 ID">{{ context.tenantId }}</el-descriptions-item>
          <el-descriptions-item label="成员角色">{{ context.role }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import request from '@/utils/request'

const loading = ref(false)
const context = ref(null)
const error = ref('')

onMounted(async () => {
  loading.value = true
  try {
    const res = await request({ url: '/uni/workspace/context', method: 'get' })
    if (res?.code === 0) context.value = res.data
    else error.value = res?.msg || '获取高校身份失败'
  } catch (e) {
    error.value = '获取高校身份失败，请检查登录状态和权限'
  } finally {
    loading.value = false
  }
})
</script>
