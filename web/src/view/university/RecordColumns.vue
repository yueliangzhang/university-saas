<template>
  <el-popover trigger="click" placement="bottom-end" width="180">
    <template #reference><el-button>显示列</el-button></template>
    <el-checkbox-group v-model="visible">
      <el-checkbox v-for="(label, key) in columns" :key="key" :value="key">{{ label }}</el-checkbox>
    </el-checkbox-group>
  </el-popover>
</template>

<script setup>
defineProps({ columns: { type: Object, required: true } })
const visible = defineModel({ type: Array, required: true })
</script>
