<template>
  <div class="uni-stat-card">
    <div class="uni-stat-card__label">{{ label }}</div>
    <div class="uni-stat-card__value">{{ value }}</div>
    <div v-if="hint" class="uni-stat-card__hint">{{ hint }}</div>
    <slot />
  </div>
</template>
<script setup>
defineProps({
  label: { type: String, required: true },
  value: { type: [String, Number], default: '—' },
  hint: { type: String, default: '' },
})
</script>
<style scoped>
.uni-stat-card { background:var(--uni-surface,#fff); border:1px solid var(--uni-border,#e5eaf2); border-radius:var(--uni-radius-lg,16px); padding:20px 24px; box-shadow:var(--uni-shadow,0 4px 20px rgba(15,23,42,.045)); min-width:0; }
.uni-stat-card__label { color:var(--uni-muted,#64748b); font-size:13px; }
.uni-stat-card__value { color:var(--uni-text,#17243b); font-size:30px; font-weight:750; line-height:1.4; margin:8px 0; overflow-wrap:anywhere; }
.uni-stat-card__hint { color:var(--uni-muted,#64748b); font-size:12px; }
</style>
