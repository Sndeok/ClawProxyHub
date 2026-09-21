<template>
  <!-- 状态码标签：2xx 成功、4xx 警告、5xx 失败；带错误摘要时用 tooltip 展开 -->
  <t-tooltip v-if="message" :content="message" placement="top-left" :overlay-style="overlayStyle">
    <t-tag :theme="theme" variant="light">{{ code }}</t-tag>
  </t-tooltip>
  <t-tag v-else :theme="theme" variant="light">{{ code }}</t-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ code: number; message?: string }>()

const theme = computed(() => {
  if (props.code < 400) return 'success'
  if (props.code < 500) return 'warning'
  return 'danger'
})
const overlayStyle = { maxWidth: '640px', whiteSpace: 'pre-wrap' as const }
</script>
