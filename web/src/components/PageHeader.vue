<template>
  <!-- 页面标题区：标题 + 说明 + 右侧操作槽（各页统一，不再各写一份 flex 布局） -->
  <div class="page-header">
    <div class="ph-left">
      <slot name="title">
        <h1 class="ph-title">{{ title }}</h1>
      </slot>
      <p v-if="subtitle" class="ph-sub">{{ subtitle }}</p>
    </div>
    <div v-if="$slots.actions" class="ph-actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title?: string; subtitle?: string }>()
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}
.ph-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  letter-spacing: -0.02em;
  color: var(--cph-text-1);
}
.ph-sub {
  margin: 2px 0 0;
  font-size: 12.5px;
  color: var(--cph-text-3);
}
.ph-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
</style>
