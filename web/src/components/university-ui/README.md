# V17.1 高校 SaaS UI 组件

CSS 已从 `src/main.js` 或 `src/main.ts` 加载。为避免影响现有 GVA 页面，组件样式需要在页面根元素加 `class="uni-ui"` 后才生效。

示例：

```vue
<script setup>
import { UniPageHeader, UniStatCard, UniStatusTag, UniEmptyState } from '@/components/university-ui'
</script>
<template>
  <div class="uni-ui uni-page">
    <UniPageHeader title="教师管理" description="管理高校教师档案" />
    <div class="uni-grid">
      <UniStatCard label="教师总数" :value="120" hint="示例数据，请替换为接口返回值" />
    </div>
    <div class="uni-panel" style="margin-top: 16px">
      <UniStatusTag status="active" />
      <UniEmptyState description="暂无记录" />
    </div>
  </div>
</template>
```

后续 V17.2–V17.4 逐页接入，当前版本不自动修改现有业务页面、路由或 API。
