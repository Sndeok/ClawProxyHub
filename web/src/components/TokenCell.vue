<template>
  <!--
    用量单元格：输入 / 输出 / 缓存命中。
    - 用 SVG 图标 + 文字标签，不靠颜色或符号单独表意（可访问性）
    - 悬停展开明细（含缓存命中率），命中率分母是「输入总量」，缓存是输入的子集
  -->
  <t-tooltip placement="top-left" :overlay-style="{ minWidth: '220px' }">
    <span class="tokens">
      <span class="tok tok-in" :title="t('logs.inputTokens')">
        <arrow-down-icon />{{ fmt(input) }}
      </span>
      <span class="tok tok-out" :title="t('logs.outputTokens')">
        <arrow-up-icon />{{ fmt(output) }}
      </span>
      <span v-if="cached" class="tok tok-cache" :title="t('logs.cached')">
        {{ t('logs.cachedShort') }} {{ fmt(cached) }}
      </span>
    </span>
    <template #content>
      <div class="tok-detail">
        <div class="tok-detail-title">{{ t('logs.tokenDetail') }}</div>
        <div class="tok-detail-row"><span>{{ t('logs.inputTokens') }}</span><b>{{ fmt(input) }}</b></div>
        <div class="tok-detail-row"><span>{{ t('logs.outputTokens') }}</span><b>{{ fmt(output) }}</b></div>
        <div v-if="cached" class="tok-detail-row"><span>{{ t('logs.cached') }}</span><b>{{ fmt(cached) }}</b></div>
        <div class="tok-detail-row"><span>{{ t('logs.hitRate') }}</span><b>{{ hitRate }}</b></div>
        <div class="tok-detail-total"><span>{{ t('logs.totalTokens') }}</span><b>{{ fmt(total) }}</b></div>
      </div>
    </template>
  </t-tooltip>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowDownIcon, ArrowUpIcon } from 'tdesign-icons-vue-next'

const props = defineProps<{ input?: number; output?: number; cached?: number }>()
const { t } = useI18n()

const input = computed(() => props.input ?? 0)
const output = computed(() => props.output ?? 0)
const cached = computed(() => props.cached ?? 0)
// 总 Token = 输入 + 输出（缓存命中已含在输入里，不能再加一遍）
const total = computed(() => input.value + output.value)
const hitRate = computed(() => {
  if (!cached.value || input.value <= 0) return '0%'
  return `${Math.min(100, (cached.value / input.value) * 100).toFixed(1)}%`
})

function fmt(n: number): string {
  if (!n) return '0'
  if (n < 1000) return String(n)
  if (n < 1000000) return `${(n / 1000).toFixed(1).replace(/\.0$/, '')}K`
  return `${(n / 1000000).toFixed(2)}M`
}
</script>

<style scoped>
.tokens {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  white-space: nowrap;
  cursor: default;
  font-variant-numeric: tabular-nums;
}
.tok {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}
.tok-in { color: var(--cph-success); }
.tok-out { color: var(--cph-info); }
.tok-cache { color: var(--cph-text-3); }
.tok-detail { min-width: 200px; }
.tok-detail-title { font-weight: 600; margin-bottom: 8px; }
.tok-detail-row {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  padding: 2px 0;
}
.tok-detail-total {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  margin-top: 6px;
  padding-top: 6px;
  border-top: 1px solid rgba(255, 255, 255, 0.2);
}
</style>
