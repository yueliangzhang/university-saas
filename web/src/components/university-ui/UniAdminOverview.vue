<template>
  <section class="uni-overview">
    <div class="uni-overview__heading">
      <div>
        <div class="uni-overview__eyebrow">UNIVERSITY OVERVIEW</div>
        <h2>校园运营概览</h2>
        <p>快速了解本校基础数据，统计结果来自实时业务接口</p>
      </div>
      <el-button :loading="loading" @click="refresh">刷新数据</el-button>
    </div>
    <div class="uni-overview__stats">
      <div v-for="item in metrics" :key="item.key" class="uni-overview__stat">
        <div class="uni-overview__label">{{ item.label }}</div>
        <div class="uni-overview__number">{{ item.value === null ? '—' : item.value }}</div>
        <div class="uni-overview__note">{{ item.note }}</div>
      </div>
    </div>
    <div class="uni-overview__tip">提示：无法获取总数时显示“—”，不会用分页记录数冒充全校总数。</div>
  </section>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import request from '@/utils/request'
const loading = ref(false)
const metrics = ref([
  { key: 'students', label: '学生总数', value: null, note: '全校学生档案' },
  { key: 'teachers', label: '教师总数', value: null, note: '全校教师档案' },
  { key: 'colleges', label: '学院总数', value: null, note: '学院组织架构' },
])
const endpoints = [
  ['/uni/students', { page: 1, pageSize: 1 }],
  ['/uni/teachers', { page: 1, pageSize: 1 }],
  ['/uni/org/colleges', { page: 1, pageSize: 1 }],
]
async function refresh() {
  loading.value = true
  try {
    await Promise.all(metrics.value.map(async (metric, index) => {
      metric.value = null
      try {
        const [url, params] = endpoints[index]
        const result = await request({ url, method: 'get', params })
        const total = result?.code === 0 ? result?.data?.total : null
        if (typeof total === 'number' && Number.isFinite(total) && total >= 0) metric.value = total
      } catch (_) { /* 单个统计接口失败不影响其他卡片 */ }
    }))
  } finally { loading.value = false }
}
onMounted(refresh)
</script>
<style scoped>
.uni-overview{padding:26px;margin-bottom:18px;border:1px solid #e5eaf4;border-radius:18px;background:linear-gradient(120deg,#fff,#f5f8ff);color:#22324a}
.uni-overview__heading{display:flex;align-items:center;justify-content:space-between;gap:16px;flex-wrap:wrap}
.uni-overview__eyebrow{font-size:11px;font-weight:750;letter-spacing:1.8px;color:#5075dc}
.uni-overview h2{font-size:24px;line-height:1.4;font-weight:750;margin:7px 0 5px}
.uni-overview p{margin:0;color:#71819a;font-size:13px}
.uni-overview__stats{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:14px;margin-top:22px}
.uni-overview__stat{border:1px solid #e7edf7;border-radius:14px;padding:20px;background:#fff;min-width:0}
.uni-overview__label{color:#667891;font-size:13px}.uni-overview__number{font-size:32px;font-weight:780;color:#223b74;margin:12px 0 7px;font-variant-numeric:tabular-nums}.uni-overview__note{font-size:12px;color:#91a0b4}
.uni-overview__tip{font-size:12px;color:#8998ab;margin-top:14px}
@media(max-width:700px){.uni-overview{padding:16px}.uni-overview__stats{grid-template-columns:1fr;gap:10px}.uni-overview__stat{padding:14px}}
</style>
