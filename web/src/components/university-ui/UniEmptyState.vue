<template>
  <div class="uni-empty-state">
    <el-empty :description="description" :image-size="imageSize">
      <slot />
    </el-empty>
  </div>
</template>
<script setup>
defineProps({ description: { type: String, default: '暂无数据' }, imageSize: { type: Number, default: 100 } })
</script>
<style scoped>.uni-empty-state { padding:12px 0; }</style>
