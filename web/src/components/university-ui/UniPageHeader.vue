<template>
  <header class="uni-page-header">
    <div class="uni-page-header__copy">
      <div v-if="eyebrow" class="uni-page-header__eyebrow">{{ eyebrow }}</div>
      <h1 class="uni-page-header__title">{{ title }}</h1>
      <p v-if="description" class="uni-page-header__description">{{ description }}</p>
    </div>
    <div v-if="$slots.actions" class="uni-page-header__actions"><slot name="actions" /></div>
  </header>
</template>
<script setup>
defineProps({
  title: { type: String, required: true },
  description: { type: String, default: '' },
  eyebrow: { type: String, default: '' },
})
</script>
<style scoped>
.uni-page-header { display:flex; align-items:flex-start; justify-content:space-between; gap:16px; flex-wrap:wrap; margin-bottom:24px; }
.uni-page-header__eyebrow { color:var(--uni-primary,#2563eb); font-size:12px; font-weight:700; letter-spacing:.06em; margin-bottom:6px; }
.uni-page-header__title { margin:0; color:var(--uni-text,#17243b); font-size:clamp(22px,2.2vw,28px); font-weight:700; line-height:1.3; }
.uni-page-header__description { margin:8px 0 0; color:var(--uni-muted,#64748b); font-size:14px; }
.uni-page-header__actions { display:flex; align-items:center; gap:8px; flex-wrap:wrap; }
</style>
