<template><el-tag :type="tagType" effect="light" size="small">{{ displayText }}</el-tag></template>
<script setup>
import { computed } from 'vue'
const props = defineProps({ status: { type: [String, Number], default: '' }, text: { type: String, default: '' } })
const statusMap = {
  pending: ['warning', '待审批'], approved: ['success', '已通过'], rejected: ['danger', '已拒绝'],
  cancelled: ['info', '已取消'], active: ['success', '正常'], inactive: ['info', '停用'],
}
const normalized = computed(() => String(props.status ?? '').toLowerCase())
const tagType = computed(() => statusMap[normalized.value]?.[0] || 'info')
const displayText = computed(() => props.text || statusMap[normalized.value]?.[1] || String(props.status ?? '—'))
</script>
