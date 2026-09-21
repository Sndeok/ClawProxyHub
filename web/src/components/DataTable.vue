<template>
  <!-- 表格外壳：统一横向滚动容器 + 加载/空态 + 可选分页，避免每页各写一份 .table-wrap -->
  <div class="table-shell">
    <div class="table-wrap">
      <slot />
    </div>
    <div v-if="showFooter" class="table-foot">
      <span class="muted">{{ totalText }}</span>
      <div class="foot-right">
        <slot name="footer" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ total?: number }>()
const { t } = useI18n()
const showFooter = computed(() => props.total !== undefined)
const totalText = computed(() => t('logs.totalCount', { n: props.total ?? 0 }))
</script>

<style scoped>
.table-shell { width: 100%; }
.table-wrap {
  width: 100%;
  overflow-x: auto;
  border: 1px solid var(--cph-border);
  border-radius: var(--cph-radius-lg);
  background: var(--cph-surface);
}
.table-wrap :deep(.t-table) { border-radius: 0; }
.table-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 12px;
}
.foot-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
</style>
