<template>
  <div class="page">
    <!-- 过滤条：时间范围 / 状态 / 协议 / 密钥 / 插件 / 账号 / 模型 / 关键词 / 延迟 -->
    <t-card class="filter-card" :bordered="false">
      <div class="filters">
        <t-select v-model="filters.range" :options="rangeOptions" class="w130" @change="onRangeChange" />
        <t-date-range-picker
          v-if="filters.range === 'custom'"
          v-model="customRange"
          class="w330"
          value-type="YYYY-MM-DD HH:mm:ss"
          enable-time-picker
          allow-input
          clearable
          @change="search"
        />
        <t-select v-model="filters.status" :options="statusOptions" class="w120" />
        <t-select v-model="filters.protocol" :options="protocolOptions" class="w160" />
        <t-select v-model="filters.key_id" :options="keyOptions" class="w160" filterable clearable />
        <t-select v-model="filters.plugin_id" :options="pluginOptions" class="w150" clearable />
        <t-select v-model="filters.account_id" :options="accountOptions" class="w170" filterable clearable />
        <t-input v-model="filters.model" :placeholder="$t('logs.modelPh')" class="w170" clearable @enter="search" />
        <t-input v-model="filters.q" :placeholder="$t('logs.keywordPh')" class="w170" clearable @enter="search" />
        <t-input-number v-model="filters.min_latency" :min="0" :placeholder="$t('logs.minLatency')" class="w130" />
        <t-button theme="primary" @click="search">{{ $t('logs.search') }}</t-button>
        <t-button variant="outline" @click="reset">{{ $t('logs.reset') }}</t-button>
      </div>
    </t-card>

    <!-- 工具条 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <span class="muted">
          {{ $t('logs.totalCount', { n: total }) }}<template v-if="rangeLabel"> · {{ rangeLabel }}</template>
        </span>
        <span v-if="logs.length" class="page-sum">
          <span class="num">Σ {{ fmt(pageSummary.tokens) }}</span>
          <span class="sep">·</span>
          <span>{{ $t('logs.hitRate') }} <b class="num">{{ pageSummary.hitRate }}</b></span>

        </span>
        <label class="auto-refresh">
          <t-switch v-model="autoRefresh" size="small" />
          <span>{{ $t('logs.autoRefresh') }}</span>
        </label>
      </div>
      <div class="toolbar-right">
        <t-button variant="outline" :loading="loading" @click="refresh">{{ $t('common.refresh') }}</t-button>
        <t-button variant="outline" theme="danger" @click="cleanupVisible = true">{{ $t('logs.cleanup') }}</t-button>
      </div>
    </div>

    <data-table>
      <t-table
          row-key="ID"
          :data="logs"
          :columns="columns"
          :loading="loading"
          table-layout="auto"
          hover
          @row-click="onRowClick"
        >
          <template #model="{ row }">
          <t-tooltip
            placement="top-left"
            :content="`${$t('logs.requestedModel')}: ${row.RequestedModel || '-'}` + (modelChanged(row) ? ` / ${$t('logs.upstreamModel')}: ${row.Model || '-'}` : '')"
          >
            <span class="model-cell">
              <span class="model-alias">{{ row.RequestedModel || row.Model || '-' }}</span>
              <template v-if="modelChanged(row)">
                <span class="model-arrow">→</span>
                <span class="model-real">{{ row.Model }}</span>
              </template>
            </span>
          </t-tooltip>
        </template>
        <template #protocol="{ row }">
          <t-tooltip :content="`${dict(protocolDict, row.Protocol)} · ${row.Stream ? $t('logs.streaming') : $t('logs.nonStreaming')}`" placement="top">
            <div class="proto-cell">
              <span class="proto-dot" :class="{ live: row.Stream }" />
              <span class="proto-name">{{ shortProtocol(row.Protocol) }}</span>
            </div>
          </t-tooltip>
        </template>
        <template #key="{ row }">
          <span v-if="row.key_name">{{ row.key_name }}</span>
          <span v-else class="dim">-</span>
        </template>
          <template #status="{ row }">
          <status-tag :code="row.Status" :message="row.ErrorBrief" />
        </template>
          <template #account="{ row }">
            <span v-if="row.account_name" class="ellipsis acct" :title="row.account_name">
              {{ row.account_name }}
            </span>
            <span v-else-if="row.AccountID" class="dim">#{{ row.AccountID }}</span>
            <span v-else class="dim">-</span>
          </template>

          <template #tokens="{ row }">
          <token-cell :input="row.InputTokens" :output="row.OutputTokens" :cached="row.CachedTokens" />
        </template>
        <template #latency="{ row }">
          <!-- 首字/总耗时合并：绿条 + tooltip -->
          <t-tooltip placement="top-left">
            <div class="latency">
              <span class="latency-bar" :class="{ slow: row.LatencyMs >= 10000 }"></span>
              <div class="latency-nums">
                <div>{{ $t('logs.firstToken') }} <b>{{ fmtMs(row.FirstTokenMs) }}</b></div>
                <div>{{ $t('logs.totalTime') }} <b>{{ fmtMs(row.LatencyMs) }}</b></div>
              </div>
            </div>
            <template #content>
              <div class="tok-detail">
                <div class="tok-detail-title">{{ $t('logs.latencyTitle') }}</div>
                <div class="tok-detail-row"><span>{{ $t('logs.firstToken') }}</span><b>{{ fmtMs(row.FirstTokenMs) }}</b></div>
                <div class="tok-detail-row"><span>{{ $t('logs.totalTime') }}</span><b>{{ fmtMs(row.LatencyMs) }}</b></div>
              </div>
            </template>
          </t-tooltip>
        </template>
          <template #attempts="{ row }">
          <span :class="row.Attempts > 1 ? 'attempts-retry' : 'dim'">{{ row.Attempts || 1 }}</span>
        </template>
          <template #ua="{ row }">
          <t-tooltip v-if="row.UserAgent" :content="row.UserAgent" placement="top-left">
            <span class="ellipsis">{{ row.UserAgent }}</span>
          </t-tooltip>
          <span v-else>-</span>
        </template>
          <template #empty>
          <div class="empty">{{ $t('logs.empty') }}</div>
        </template>
      </t-table>
    </data-table>

    <div class="pager">
      <t-pagination
        v-model:current="page"
        v-model:pageSize="pageSize"
        :total="total"
        :page-size-options="[20, 50, 100]"
        show-jumper
        @current-change="load"
        @page-size-change="onPageSizeChange"
      />
    </div>

    <!-- 详情抽屉：把一行日志的所有诊断字段摊开 -->
    <t-drawer v-model:visible="detailVisible" :header="$t('logs.detail')" size="520px" :footer="false">
      <div v-if="detail" class="detail">
        <div v-for="(row, i) in detailRows" :key="i" class="dl-row">
          <span class="dl-k">{{ row.label }}</span>
          <span class="dl-v" :class="{ danger: row.danger }">{{ row.value }}</span>
        </div>
        <div v-if="detail.ErrorBrief" class="err-block">
          <div class="err-label">{{ $t('logs.errorDetail') }}</div>
          <pre class="err-text">{{ detail.ErrorBrief }}</pre>
        </div>
      </div>
    </t-drawer>

    <!-- 清理日志 -->
    <t-dialog
      v-model:visible="cleanupVisible"
      :header="$t('logs.cleanup')"
      :confirm-btn="{ theme: 'danger' }"
      @confirm="doCleanup"
    >
      <p class="cleanup-hint">{{ $t('logs.cleanupConfirm') }}</p>
      <t-radio-group v-model="cleanupScope" :options="cleanupOptions" />
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { api } from '../api/client'
import DataTable from '../components/DataTable.vue'
import StatusTag from '../components/StatusTag.vue'
import TokenCell from '../components/TokenCell.vue'
import { dict, protocolDict } from '../utils/dict'
import type { LogPage, RequestLog } from '../api/types'

const { t } = useI18n()

const logs = ref<RequestLog[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

// 关联数据：密钥 / 插件 / 账号（供过滤下拉与名称展示）
const keys = ref<{ id: number; name: string }[]>([])
const plugins = ref<{ id: number; name: string; label?: string }[]>([])
const accounts = ref<{ id: number; display_name: string; plugin_id: number }[]>([])

type FilterValue = string | number
const filters = reactive({
  range: '24h',
  status: '' as FilterValue,
  protocol: '' as FilterValue,
  key_id: '' as FilterValue,
  plugin_id: '' as FilterValue,
  account_id: '' as FilterValue,
  model: '',
  q: '',
  min_latency: 0,
})
const customRange = ref<string[]>([])

const autoRefresh = ref(false)
let timer: ReturnType<typeof setInterval> | undefined

const detailVisible = ref(false)
const detail = ref<RequestLog | null>(null)

const cleanupVisible = ref(false)
const cleanupScope = ref<FilterValue>(7)

// ---------- 下拉选项 ----------
const rangeOptions = computed(() => [
  { label: t('logs.rangeAll'), value: 'all' },
  { label: t('logs.range1h'), value: '1h' },
  { label: t('logs.range24h'), value: '24h' },
  { label: t('logs.range7d'), value: '7d' },
  { label: t('logs.range30d'), value: '30d' },
  { label: t('logs.rangeCustom'), value: 'custom' },
])
const statusOptions = computed(() => [
  { label: t('logs.allStatus'), value: '' },
  { label: t('logs.statusOk'), value: 'ok' },
  { label: t('logs.statusError'), value: 'error' },
  { label: '4xx', value: '4xx' },
  { label: '5xx', value: '5xx' },
])
const protocolOptions = computed(() => [
  { label: t('logs.allProtocol'), value: '' },
  { label: dict(protocolDict, 'messages'), value: 'messages' },
  { label: dict(protocolDict, 'chat_completions'), value: 'chat_completions' },
  { label: dict(protocolDict, 'responses'), value: 'responses' },
])
const keyOptions = computed(() => [
  { label: t('logs.allKeys'), value: '' },
  ...keys.value.map((k) => ({ label: k.name || `#${k.id}`, value: k.id })),
])
const pluginOptions = computed(() => [
  { label: t('logs.allPlugins'), value: '' },
  ...plugins.value.map((p) => ({ label: p.label || p.name, value: p.id })),
])
const accountOptions = computed(() => [
  { label: t('logs.allAccounts'), value: '' },
  ...accounts.value.map((a) => {
    const p = plugins.value.find((x) => x.id === a.plugin_id)
    const brand = p?.label || p?.name || `#${a.plugin_id}`
    return { label: `${a.display_name || '#' + a.id} · ${brand}`, value: a.id }
  }),
])
// 当前时间范围文案：让「只能看到 24 小时内的日志」这类疑惑一眼可解
const rangeLabel = computed(
  () => rangeOptions.value.find((o) => o.value === filters.range)?.label ?? '',
)
const cleanupOptions = computed(() => [
  { label: t('logs.cleanup7'), value: 7 },
  { label: t('logs.cleanup30'), value: 30 },
  { label: t('logs.cleanupAll'), value: 'all' },
])

// ---------- 表格列 ----------
const columns = computed(() => [
  {
    colKey: 'CreatedAt', title: t('common.colTime'), width: 146, align: 'center',
    cell: (_h: any, { row }: any) => fmtTime(row.CreatedAt),
  },
  { colKey: 'status', title: t('common.colStatus'), width: 72, align: 'center' },
  { colKey: 'key', title: t('logs.key'), width: 100, ellipsis: true },
  { colKey: 'account', title: t('logs.accountCol'), width: 110, ellipsis: true },
  // 请求模型与上游模型合并：一致时只显示一次，路由改写时用 → 展示真实去向
  { colKey: 'model', title: t('logs.model'), width: 170, ellipsis: true },
  {
    colKey: 'plugin', title: t('logs.plugin'), width: 92, align: 'center',
    cell: (_h: any, { row }: any) => pluginLabel(row.PluginID),
  },
  // 协议 + 流式合并，省一列宽度；流式用浅色标签区分
  { colKey: 'protocol', title: t('logs.protocol'), width: 118, align: 'center' },
  { colKey: 'tokens', title: 'Token', width: 148, align: 'center' },
  { colKey: 'latency', title: t('logs.latency'), width: 112, align: 'center' },
  { colKey: 'attempts', title: t('logs.attempts'), width: 72, align: 'center' },
  { colKey: 'ClientIP', title: 'IP', width: 104, align: 'center' },
])

// ---------- 详情 ----------
interface DetailRow { label: string; value: string; danger?: boolean }

const detailRows = computed<DetailRow[]>(() => {
  const d = detail.value
  if (!d) return []
  const rows: DetailRow[] = [
    { label: t('common.colTime'), value: fmtTime(d.CreatedAt) },
    { label: t('common.colStatus'), value: String(d.Status), danger: d.Status >= 400 },
    { label: t('logs.protocol'), value: dict(protocolDict, d.Protocol) },
    { label: t('logs.stream'), value: d.Stream ? t('logs.streaming') : t('logs.nonStreaming') },
    { label: t('logs.requestedModel'), value: d.RequestedModel || '-' },
    { label: t('logs.upstreamModel'), value: d.Model || '-' },
    { label: t('logs.routeID'), value: d.RouteID ? String(d.RouteID) : '-' },
    { label: t('logs.groupID'), value: d.GroupID ? String(d.GroupID) : '-' },
    { label: t('logs.key'), value: d.key_name || '-' },
    { label: t('logs.plugin'), value: pluginLabel(d.PluginID) },
    { label: t('logs.account'), value: d.account_name || (d.AccountID ? `#${d.AccountID}` : '-') },
    { label: t('logs.inputTokens'), value: `${d.InputTokens || 0}${d.CachedTokens ? '（含缓存命中）' : ''}` },
    { label: t('logs.outputTokens'), value: String(d.OutputTokens || 0) },
    { label: t('logs.cached'), value: `${d.CachedTokens || 0}（${hitRate(d)}）` },
    { label: t('logs.totalTokens'), value: String(totalTokens(d)) },
    { label: t('logs.firstToken'), value: fmtMs(d.FirstTokenMs) },
    { label: t('logs.totalTime'), value: fmtMs(d.LatencyMs) },
    { label: t('logs.attempts'), value: String(d.Attempts || 1), danger: (d.Attempts || 1) > 1 },
    { label: t('logs.finishReason'), value: d.FinishReason || '-' },
    { label: 'IP', value: d.ClientIP || '-' },
    { label: 'User-Agent', value: d.UserAgent || '-' },
  ]
  if (d.ErrorType) rows.push({ label: t('logs.errorType'), value: d.ErrorType, danger: true })
  return rows
})

function pluginLabel(pluginID: number | null): string {
  if (!pluginID) return '-'
  const p = plugins.value.find((x) => x.id === pluginID)
  return p?.label || p?.name || `#${pluginID}`
}

function onRowClick(payload: any) {
  const row: RequestLog | undefined = payload?.row ?? payload
  if (!row || !row.ID) return
  detail.value = row
  detailVisible.value = true
}

// ---------- 格式化 ----------
function fmt(n: number): string {
  if (!n) return '0'
  if (n < 1000) return String(n)
  if (n < 1000000) return `${(n / 1000).toFixed(1).replace(/\.0$/, '')}K`
  return `${(n / 1000000).toFixed(2)}M`
}

function fmtMs(ms: number): string {
  if (!ms) return '-'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

// 协议短名：列里只放能一眼区分的部分，完整名走 tooltip
function shortProtocol(p: string): string {
  if (p === 'chat_completions') return 'Chat'
  if (p === 'responses') return 'Responses'
  if (p === 'messages') return 'Messages'
  return p || '-'
}

// 路由改写模型名时才需要展示「请求 → 上游」，否则重复展示没意义
function modelChanged(row: RequestLog): boolean {
  const req = row.RequestedModel || row.Model
  return !!row.Model && !!req && row.Model !== req
}

function fmtTime(v: string): string {
  return v?.replace('T', ' ').slice(0, 19) ?? '-'
}

// 总 Token = 输入 + 输出。缓存命中是**输入的子集**（OpenAI / Anthropic 都是这个口径），
// 早先这里又加了一次 cached，会把命中部分算两遍。
function totalTokens(row: RequestLog): number {
  return (row.InputTokens || 0) + (row.OutputTokens || 0)
}


// 命中率 = 命中 / 输入总量（cached 已包含在 input 里，分母不能再加它）
function hitRate(row: RequestLog): string {
  return rate(row.CachedTokens || 0, row.InputTokens || 0)
}

// rate 计算百分比并夹到 100%（上游偶尔给出比输入还大的命中值时也不显示 >100%）
function rate(part: number, whole: number): string {
  if (!part || whole <= 0) return '0%'
  const pct = Math.min(100, (part / whole) * 100)
  return `${pct.toFixed(1)}%`
}

// 本页汇总：让「这一屏大概花了多少」一眼可见，不用心算
const pageSummary = computed(() => {
  let tokens = 0, cached = 0, input = 0
  for (const row of logs.value) {
    tokens += (row.InputTokens || 0) + (row.OutputTokens || 0)
    cached += row.CachedTokens || 0
    input += row.InputTokens || 0
  }
  return { tokens, hitRate: rate(cached, input) }
})

// ---------- 查询 ----------
function pad(n: number): string {
  return n < 10 ? `0${n}` : String(n)
}

function stamp(d: Date): string {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
    `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// 预设时间范围 → from/to（后端按本地时间解析）
function rangeParams(): { from?: string; to?: string } {
  const now = Date.now()
  const hours = (h: number) => stamp(new Date(now - h * 3600_000))
  switch (filters.range) {
    case '1h': return { from: hours(1) }
    case '24h': return { from: hours(24) }
    case '7d': return { from: hours(24 * 7) }
    case '30d': return { from: hours(24 * 30) }
    case 'custom': {
      const [from, to] = customRange.value ?? []
      return { from: from || undefined, to: to || undefined }
    }
    default: return {}
  }
}

function buildQuery(): string {
  const p = new URLSearchParams()
  p.set('page', String(page.value))
  p.set('page_size', String(pageSize.value))
  if (filters.status !== '') p.set('status', String(filters.status))
  if (filters.protocol !== '') p.set('protocol', String(filters.protocol))
  if (filters.key_id !== '') p.set('key_id', String(filters.key_id))
  if (filters.plugin_id !== '') p.set('plugin_id', String(filters.plugin_id))
  if (filters.account_id !== '') p.set('account_id', String(filters.account_id))
  if (filters.model.trim()) p.set('model', filters.model.trim())
  if (filters.q.trim()) p.set('q', filters.q.trim())
  if (filters.min_latency > 0) p.set('min_latency', String(filters.min_latency))
  const { from, to } = rangeParams()
  if (from) p.set('from', from)
  if (to) p.set('to', to)
  return p.toString()
}

async function load() {
  if (loading.value) return
  loading.value = true
  try {
    const r = await api.get<LogPage>(`/admin/logs?${buildQuery()}`)
    logs.value = r.logs ?? []
    total.value = r.total ?? 0
  } catch (e: any) {
    MessagePlugin.error(e.message)
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  load()
}

async function refresh() {
  await load()
}

function onPageSizeChange() {
  page.value = 1
  load()
}

function onRangeChange() {
  if (filters.range !== 'custom') search()
}

function reset() {
  filters.range = '24h'
  filters.status = ''
  filters.protocol = ''
  filters.key_id = ''
  filters.plugin_id = ''
  filters.account_id = ''
  filters.model = ''
  filters.q = ''
  filters.min_latency = 0
  customRange.value = []
  search()
}

async function loadRefs() {
  const [k, p, a] = await Promise.all([
    api.get<{ keys: { id: number; name: string }[] }>('/admin/keys').catch(() => ({ keys: [] })),
    api.get<{ plugins: { id: number; name: string; label?: string }[] }>('/admin/plugins').catch(() => ({ plugins: [] })),
    api.get<{ accounts: { id: number; display_name: string; plugin_id: number }[] }>('/admin/accounts').catch(() => ({ accounts: [] })),
  ])
  keys.value = k.keys ?? []
  plugins.value = p.plugins ?? []
  accounts.value = a.accounts ?? []
}

async function doCleanup() {
  const body = cleanupScope.value === 'all' ? { all: true } : { days: Number(cleanupScope.value) }
  try {
    const r = await api.post<{ deleted: number }>('/admin/logs/cleanup', body)
    MessagePlugin.success(t('logs.cleanupDone', { n: r.deleted }))
    cleanupVisible.value = false
    search()
  } catch (e: any) {
    MessagePlugin.error(e.message)
  }
}

// 自动刷新：仅在开启时轮询当前查询（load 内部有重入保护）
watch(autoRefresh, (on) => {
  if (timer) { clearInterval(timer); timer = undefined }
  if (on) timer = setInterval(load, 5000)
})

onMounted(async () => {
  await Promise.all([loadRefs(), load()])
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.filter-card {
  background: var(--td-bg-color-container);
  margin-bottom: 12px;
}
.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}
.w120 { width: 120px }
.w130 { width: 130px }
.w150 { width: 150px }
.w160 { width: 160px }
.w170 { width: 170px }
.w330 { width: 330px }

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
  gap: 12px;
  flex-wrap: wrap;
}
.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.page-sum {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px;
  border-radius: 14px;
  background: var(--cph-surface-2);
  border: 1px solid var(--cph-border);
  font-size: 12px;
  color: var(--cph-text-2);
}
.page-sum b {
  color: var(--cph-text-1);
}
.page-sum .sep {
  color: var(--cph-border-strong);
}
.model-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
  cursor: default;
}
.model-alias {
  font-weight: 600;
  color: var(--cph-text-1);
}
.model-arrow {
  color: var(--cph-text-3);
}
.model-real {
  color: var(--cph-text-2);
  font-size: 12px;
}
.proto-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
  cursor: default;
}
/* 流式/非流式用一个小圆点左右区分，比整块标签省一半宽度 */
.proto-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--cph-border-strong);
  flex-shrink: 0;
}
.proto-dot.live {
  background: var(--cph-success);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--cph-success) 18%, transparent);
}
.proto-name {
  color: var(--cph-text-2);
}
.acct {
  color: var(--cph-text-2);
}
.table-wrap :deep(.t-table) {
  border-radius: 0;
}
.auto-refresh {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--td-text-color-secondary);
  cursor: pointer;
}
.muted {
  font-size: 12px;
  color: var(--td-text-color-secondary);
}
.empty {
  padding: 24px 0;
  color: var(--td-text-color-placeholder);
}
.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 14px;
}

.ellipsis {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: bottom;
  cursor: default;
}
.dim {
  color: var(--td-text-color-placeholder);
}
.attempts-retry {
  color: var(--td-warning-color);
  font-weight: 600;
}

/* Token 合并列 */
.tokens {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  white-space: nowrap;
  cursor: default;
}
.tok-in { color: var(--td-success-color) }
.tok-out { color: var(--td-brand-color) }
.tok-cache { color: var(--td-warning-color); cursor: default }

.tok-detail { min-width: 200px }
.tok-detail-title { font-weight: 700; margin-bottom: 8px }
.tok-detail-row {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  padding: 2px 0;
}
.tok-detail-row b { font-variant-numeric: tabular-nums }
.tok-detail-total {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  margin-top: 6px;
  padding-top: 6px;
  border-top: 1px solid rgba(255, 255, 255, 0.2);
}
.tok-detail-total b { font-variant-numeric: tabular-nums }

/* 延迟合并列 */
.latency {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: default;
}
.latency-bar {
  width: 3px;
  height: 28px;
  border-radius: 2px;
  background: var(--td-success-color);
  flex-shrink: 0;
}
.latency-bar.slow {
  background: var(--td-warning-color);
}
.latency-nums {
  font-size: 12px;
  line-height: 1.5;
  white-space: nowrap;
}
.latency-nums div {
  display: flex;
  justify-content: space-between;
  gap: 6px;
}
.latency-nums b { font-variant-numeric: tabular-nums }

/* 详情抽屉 */
.detail { font-size: 13px }
.dl-row {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 7px 0;
  border-bottom: 1px solid var(--td-component-stroke);
}
.dl-k {
  color: var(--td-text-color-secondary);
  flex-shrink: 0;
}
.dl-v {
  text-align: right;
  word-break: break-all;
  font-variant-numeric: tabular-nums;
}
.dl-v.danger { color: var(--td-error-color) }
.err-block {
  margin-top: 16px;
}
.err-label {
  color: var(--td-text-color-secondary);
  margin-bottom: 6px;
}
.err-text {
  margin: 0;
  padding: 10px 12px;
  border-radius: 6px;
  background: var(--td-bg-color-secondarycontainer);
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
  line-height: 1.6;
  max-height: 260px;
  overflow: auto;
}
.cleanup-hint {
  margin: 0 0 12px;
  color: var(--td-text-color-secondary);
}
</style>