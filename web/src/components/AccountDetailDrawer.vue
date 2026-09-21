<template>
  <!--
    账号详情抽屉：账号基础信息 + 插件声明的动态区块（签到/成长/积分包）+ 最近任务执行。
    数据由父级拉到 detail 后整体传入，组件只负责渲染（不自己再打接口）。
  -->
  <t-drawer :visible="visible" :header="header" size="720px" @update:visible="(v: boolean) => emit('update:visible', v)">
    <t-space v-if="detail" direction="vertical" style="width: 100%" size="large">
      <t-descriptions :column="1" bordered size="small">
        <t-descriptions-item :label="t('accounts.account')">{{ detail.display_name || `#${detail.id}` }}</t-descriptions-item>
        <t-descriptions-item :label="t('accounts.status')">{{ dict(accountStatusDict, detail.status) }}</t-descriptions-item>
        <t-descriptions-item v-if="detail.pause_reason" :label="t('accounts.pauseReason')">{{ detail.pause_reason }}</t-descriptions-item>
        <t-descriptions-item v-if="detail.paused_until && !detail.manual_pause" :label="t('accounts.resumeAt')">
          {{ detail.paused_until?.replace('T', ' ').slice(0, 19) }}
        </t-descriptions-item>
        <t-descriptions-item :label="t('accounts.lastRefresh')">{{ fmtTime(detail.last_refresh_at) }}</t-descriptions-item>
        <t-descriptions-item :label="t('accounts.lastUsed')">{{ fmtTime(detail.last_used_at) }}</t-descriptions-item>
        <t-descriptions-item :label="t('accounts.credits')">{{ creditSummaryLine }}</t-descriptions-item>
      </t-descriptions>

      <!-- 动态渲染块：插件声明的 ProfileSection（签到 / 成长计划 / 积分包…），不定义即不渲染 -->
      <div v-for="sec in sections" :key="sec.id">
        <div class="section-title">{{ label(sec.title, sec.id) }}</div>
        <t-descriptions v-if="sec.entries?.length" :column="1" bordered size="small">
          <t-descriptions-item v-for="(e, i) in sec.entries" :key="i" :label="label(e.label, '')">
            <t-tag v-if="isStatus(e.value)" :theme="sectionStatusTheme(e.value)" variant="light" size="small">
              {{ statusValue(e.value) }}
            </t-tag>
            <template v-else>{{ e.value }}</template>
          </t-descriptions-item>
        </t-descriptions>
        <t-table
          v-if="sec.items?.length && sec.columns?.length"
          :data="sec.items"
          size="small"
          height="260"
          :row-key="(_r: Record<string, string>, i?: number) => String(i)"
          :columns="sectionColumns(sec.columns)"
          :show-header="true"
        >
          <template #section-cell="{ col, row }">
            <t-tag v-if="col.kind === 'status'" :theme="sectionStatusTheme(row.cells?.[col.colKey])" variant="light" size="small">
              {{ statusValue(row.cells?.[col.colKey]) }}
            </t-tag>
            <template v-else>{{ row.cells?.[col.colKey] || '-' }}</template>
          </template>
        </t-table>
      </div>

      <div>
        <div class="section-title">{{ t('accounts.runsTitle') }}</div>
        <t-table
          v-if="detail.runs?.length"
          row-key="ID"
          :data="detail.runs"
          size="small"
          :columns="detailRunColumns"
          :show-header="true"
        >
          <template #run-status="{ row: run }">
            <t-tag :theme="run.status === 'success' ? 'success' : run.status === 'failed' ? 'danger' : 'warning'" variant="light">
              {{ dict(runStatusDict, run.status) }}
            </t-tag>
          </template>
        </t-table>
        <t-empty v-else :description="t('accounts.noRuns')" />
      </div>
    </t-space>
  </t-drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { dict, accountStatusDict, runStatusDict } from '../utils/dict'
import type { AccountDetail } from '../api/types'
import { fmtNum, fmtTime } from '../utils/format'

const props = defineProps<{ visible: boolean; detail: AccountDetail | null }>()
const emit = defineEmits<{ 'update:visible': [boolean] }>()
const { t } = useI18n()

// 详情里的动态区块由插件声明（profile_json 随账号持久化）
interface ProfileSection {
  id: string
  title: Record<string, string>
  entries?: { label: Record<string, string>; value: string; kind?: string }[]
  columns?: { key: string; title: Record<string, string>; kind?: string }[]
  items?: Record<string, string>[]
}

const detailRunColumns = computed(() => [
  { colKey: 'run-status', title: t('tasks.colResult'), width: 80 },
  { colKey: 'capability', title: t('tasks.colTask'), width: 110, align: 'center' },
  { colKey: 'summary', title: t('tasks.colSummary'), ellipsis: true, align: 'center' },
  { colKey: 'started_at', title: t('common.colTime'), width: 160, cell: (_h: any, { row }: any) => fmtTime(row.started_at), align: 'center' },
])

// 抽屉标题：账号详情（带名称后缀）
const header = computed(() =>
  props.detail?.display_name ? `${t('accounts.detailTitle')} · ${props.detail.display_name}` : t('accounts.detailTitle'),
)

// 积分合并一行：
// - 有积分包/免费池的插件（如 lobsterai）：可用 X · N 个积分包 · 免费池 used/limit
// - 其余（如 workbuddy）：剩余 X / 已用 X / 总 X
const creditSummaryLine = computed(() => {
  const c = props.detail?.credits as any
  if (!c) return '-'
  const pkgCount = Array.isArray(c.packages) ? c.packages.length : 0
  if (pkgCount > 0 || c.free_limit !== undefined) {
    const parts: string[] = []
    if (c.remaining !== undefined) parts.push(t('accounts.avail', { v: fmtNum(c.remaining) }))
    if (pkgCount > 0) parts.push(t('accounts.packagesN', { n: pkgCount }))
    if (c.free_limit !== undefined) {
      parts.push(t('accounts.freePool', { used: fmtNum(c.free_used), limit: fmtNum(c.free_limit) }))
    }
    return parts.join(' · ') || '-'
  }
  const legacy = [
    t('accounts.legacyRemaining', { v: fmtNum(c.remaining) }),
    t('accounts.legacyUsed', { v: fmtNum(c.used) }),
    t('accounts.legacyTotal', { v: fmtNum(c.total) }),
  ].filter((p) => !p.endsWith(' -'))
  return legacy.join(' / ') || '-'
})

// profile 快照里的动态块（核心随 profile_json 持久化，插件声明 → 主框架渲染）
const sections = computed<ProfileSection[]>(() => {
  const raw = props.detail?.profile?.sections
  return Array.isArray(raw) ? (raw as ProfileSection[]) : []
})

// status 值协议："status:xxx" 前缀 = 徽章；其余为纯文本
function isStatus(v: string): boolean {
  return v.startsWith('status:')
}
function statusValue(v: string): string {
  return v.startsWith('status:') ? v.slice('status:'.length) : v
}
// status → 徽章色（成功类绿 / 警示类橙 / 失效类红 / 其余默认）
function sectionStatusTheme(v: string): string {
  const s = statusValue(v)
  if (/已签到|active|正常|有效/.test(s)) return 'success'
  if (/expiringSoon|临期|待领奖|未签到/.test(s)) return 'warning'
  if (/expired|已过期|已领奖/.test(s)) return s === '已领奖' ? 'success' : 'danger'
  return 'default'
}

// 区块标题/表头按当前语言取值
function label(v: Record<string, string> | undefined, fallback: string): string {
  if (!v) return fallback
  return v.zh || v.en || fallback
}

// 动态表格列：TDesign 列描述（统一走 section-cell 插槽渲染）
function sectionColumns(cols: { key: string; title: Record<string, string>; kind?: string }[]) {
  return cols.map((c) => ({
    colKey: c.key,
    title: label(c.title, c.key),
    kind: c.kind,
    cell: 'section-cell',
    ellipsis: c.kind !== 'status',
  }))
}

</script>

<style scoped>
.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--cph-text-2);
  margin-bottom: 8px;
}
</style>
