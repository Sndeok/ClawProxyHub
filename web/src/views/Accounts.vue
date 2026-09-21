<template>
  <div class="page">
    <div class="page-header">
      
      <div class="header-actions">
  <t-button variant="outline" :loading="refreshingAll" @click="refreshAll">
    <template #icon><refresh-icon /></template>
    {{ $t('accounts.refreshAll') }}
  </t-button>
  <t-button theme="primary" :disabled="!plugins.length" @click="openAdd">{{ $t('accounts.add') }}</t-button>
</div>
    </div>

    <DataTable>
      <t-table row-key="id" :data="accounts" :columns="columns" :loading="loading" table-layout="auto">
      <template #display_name="{ row }">
        <span class="acct-name" @click="openDetail(row.id)">{{ row.display_name || `#${row.id}` }}</span>
      </template>
      <template #group="{ row }">
        <div class="group-tags" @click.stop>
          <t-tag
            v-for="gid in row.group_ids ?? []"
            :key="gid"
            size="small"
            variant="light"
            closable
            @close="() => toggleGroup(row, gid, false)"
          >
            {{ groupName(gid) }}
          </t-tag>
          <t-popup trigger="click">
            <t-tag size="small" theme="default" variant="light" class="group-add">＋</t-tag>
            <template #content>
              <div class="group-picker">
                <div v-if="!groupsOf(row.plugin_id).length" class="group-picker-empty">{{ $t('accounts.noGroups') }}</div>
                <div
                  v-for="g in groupsOf(row.plugin_id)"
                  :key="g.id"
                  class="group-picker-item"
                  :class="{ active: (row.group_ids ?? []).includes(g.id) }"
                  @click="() => toggleGroup(row, g.id, !(row.group_ids ?? []).includes(g.id))"
                >
                  {{ g.name }}
                  <check-icon v-if="(row.group_ids ?? []).includes(g.id)" />
                </div>
              </div>
            </template>
          </t-popup>
        </div>
      </template>
      <template #schedule="{ row }">
        <t-tooltip v-if="pausedInfo(row)" :content="pausedInfo(row)!" placement="top">
          <t-tag theme="warning" variant="light">{{ pausedLabel(row) }}</t-tag>
        </t-tooltip>
        <t-switch
          v-else
          :value="row.status === 'active'"
          size="small"
          :disabled="row.status === 'expired'"
          @change="() => toggleSchedule(row)"
        />
      </template>
      <template #status="{ row }">
        <t-tag v-if="row.status === 'active'" theme="success" variant="light">{{ $t('accounts.statusActive') }}</t-tag>
        <t-tag v-else :theme="row.status === 'expired' ? 'danger' : 'default'" variant="light">
          {{ dict(accountStatusDict, row.status) }}
        </t-tag>
      </template>
      <template #expiry="{ row }">
        <t-tooltip v-if="row.credits?.next_expiry" placement="top-left">
          <span class="expiry" :class="{ soon: expiringSoon(row.credits) }">
            {{ expiryLabel(row.credits) }}
          </span>
          <template #content>
            <div class="tok-detail">
              <div class="tok-detail-title">{{ $t('accounts.expiry') }}</div>
              <div class="tok-detail-row"><span>{{ $t('accounts.nextExpiry') }}</span><b>{{ row.credits?.next_expiry }}</b></div>
              <div class="tok-detail-row"><span>{{ $t('accounts.nextLeft') }}</span><b>{{ row.credits?.next_left ?? 0 }}</b></div>
              <div class="tok-detail-row"><span>{{ $t('accounts.expiring7') }}</span><b>{{ row.credits?.expiring ?? 0 }}</b></div>
            </div>
          </template>
        </t-tooltip>
        <span v-else class="dim">-</span>
      </template>
      <template #today_tokens="{ row }">
        <t-tooltip v-if="row.today_tokens || row.today_cached" placement="top-left">
          <span class="today num">{{ fmtCompact(row.today_tokens) }}</span>
          <template #content>
            <div class="tok-detail">
              <div class="tok-detail-title">{{ $t('accounts.todayTokens') }}</div>
              <div class="tok-detail-row"><span>Token</span><b>{{ row.today_tokens || 0 }}</b></div>
              <div class="tok-detail-row"><span>{{ $t('logs.cached') }}</span><b>{{ row.today_cached || 0 }}</b></div>
              <div class="tok-detail-row"><span>{{ $t('logs.totalCount', { n: row.today_requests || 0 }) }}</span><b></b></div>
            </div>
          </template>
        </t-tooltip>
        <span v-else class="dim">-</span>
      </template>
      <template #today_credits="{ row }">
        <t-tooltip v-if="row.today_credits" placement="top-left">
          <span class="today num">
            {{ fmtCredit(row.today_credits) }}
            <em v-if="row.today_credits_estimated" class="est">{{ $t('accounts.estimated') }}</em>
          </span>
          <template #content>
            <div class="tok-detail">
              <div class="tok-detail-title">{{ $t('accounts.todayCredits') }}</div>
              <div class="tok-detail-row">
                <span>{{ row.today_credits_estimated ? $t('accounts.estimated') : $t('accounts.todayCredits') }}</span>
                <b>{{ row.today_credits }}</b>
              </div>
            </div>
          </template>
        </t-tooltip>
        <span v-else class="dim">-</span>
      </template>
      <template #credits="{ row }">
        <div v-if="row.credits" class="credit-cell">
          <span>{{ $t('accounts.remaining') }}: {{ fmtNum(row.credits.remaining) }}</span>
          <span>{{ $t('accounts.totalCredits') }}: {{ fmtNum(row.credits.total) }}</span>
        </div>
        <span v-else>-</span>
      </template>
      <template #op="{ row }">
        <t-space size="small">
          <t-link theme="primary" @click="openEdit(row)">{{ $t('common.edit') }}</t-link>
          <t-link theme="primary" @click="openTest(row)">{{ $t('accounts.test') }}</t-link>
          <t-link theme="primary" @click="refresh(row.id)">{{ $t('common.refresh') }}</t-link>
          <t-popconfirm :content="$t('accounts.confirmDelete')" @confirm="remove(row.id)">
            <t-link theme="danger">{{ $t('common.delete') }}</t-link>
          </t-popconfirm>
        </t-space>
      </template>
      </t-table></DataTable>


    <!-- 账号详情抽屉（已抽成组件） -->
    <account-detail-drawer v-model:visible="detailVisible" :detail="detail" />

    <!-- 编辑账号（组件：改名 / 绑分组 / 绑代理 / 同步模型） -->
    <account-edit-dialog
      v-model:visible="editVisible"
      :row="editRow"
      :groups="groups"
      :proxies="proxies"
      @saved="loadAll"
    />

    <!-- 在线测试（组件：直调插件 Chat，绕路由与密钥） -->
    <account-test-drawer v-model:visible="testVisible" :row="testRow" />

    <!-- 添加账号：向导（选择客户端 → 授权 → 配置） -->
    <t-dialog v-model:visible="addVisible" :header="$t('accounts.add')" :footer="false" width="680px" :close-on-overlay-click="false">

      <!-- 第一步：选择客户端（卡片平铺，每行四个） -->
      <template v-if="wizardStep === 'select'">
        <t-empty v-if="!plugins.length" :description="$t('accounts.noPlugins')" />
        <t-row v-else :gutter="[12, 12]">
          <t-col v-for="p in plugins" :key="p.id" :span="6">
            <div class="client-card" @click="choosePlugin(p)">
              <div class="client-icon">
                <img v-if="p.icon" :src="p.icon" :alt="p.label || p.name" />
                <span v-else>{{ (p.label || p.name).slice(0, 1) }}</span>
              </div>
              <div class="client-name">{{ p.label || p.name }}</div>
              <div class="client-caps">
                <t-tag v-for="c in (p.capabilities ?? []).slice(0, 3)" :key="c" size="small" variant="light">
                  {{ dict(capabilityDict, c) }}
                </t-tag>
              </div>
            </div>
          </t-col>
        </t-row>
      </template>

      <!-- 第二步：授权（tab = 插件声明的登录方式，表单按 schema 动态渲染） -->
      <t-space v-else-if="wizardStep === 'auth'" direction="vertical" style="width: 100%" size="large">
        <div class="wizard-back">
          <t-link theme="primary" @click="wizardStep = 'select'">{{ $t('accounts.backToSelect') }}</t-link>
          <span class="wizard-client">{{ selectedPluginLabel }}</span>
        </div>

        <t-tabs v-if="methods.length" v-model="methodId">
          <t-tab-panel v-for="m in methods" :key="m.id" :value="m.id" :label="label(m.label, m.id)">
            <div class="tab-body">
              <t-form v-if="currentFields?.length" label-width="90px">
                <t-form-item v-for="f in currentFields" :key="f.name" :label="f.type === 'textarea' ? '' : label(f.label, f.name)" :label-width="f.type === 'textarea' ? 0 : 90" :mark="f.required && f.type !== 'textarea'">
                  <div
                    v-if="f.type === 'textarea'"
                    class="drop-zone"
                    @drop.prevent="onDrop($event, f.name)"
                    @dragover.prevent
                  >
                    <t-textarea
                      v-model="form[f.name]"
                      :placeholder="f.placeholder || $t('accounts.pastePh')"
                      :autosize="{ minRows: 6, maxRows: 12 }"
                      class="scroll-textarea"
                    />
                    <span class="drop-hint">{{ $t('accounts.dropHint') }}</span>
                  </div>
                  <t-input                    v-else
                    v-model="form[f.name]"
                    :type="f.type === 'password' ? 'password' : 'text'"
                    :placeholder="f.placeholder"
                  />
                </t-form-item>
              </t-form>
              <t-alert v-else theme="info" :message="$t('accounts.noFieldsHint')" />
            </div>
          </t-tab-panel>
        </t-tabs>

        <!-- 浏览器授权：链接可复制可打开；auto 模式自动轮询 -->
        <t-alert v-if="nextStep" :theme="nextStep.action === 'open_url' ? 'warning' : 'info'">
          <template #message>
            <div>{{ label(nextStep.prompt, '') }}</div>
            <div v-if="nextStep.wait" class="mode-hint">
              {{ showCallbackInput ? $t('accounts.callbackManual') : $t('accounts.callbackAuto') }}
            </div>
            <div v-if="nextStep.url" class="login-url">
              <span class="login-url-text">{{ nextStep.url }}</span>
              <t-space size="small">
                <t-link theme="primary" @click="copyText(nextStep.url!)">{{ $t('accounts.copy') }}</t-link>
                <t-link theme="primary" @click="openURL(nextStep.url!)">{{ $t('accounts.open') }}</t-link>
              </t-space>
            </div>
          </template>
        </t-alert>
        <t-form v-if="nextStep?.action === 'input_form' && nextStep.fields?.length" label-width="90px">
          <t-form-item v-for="f in nextStep.fields" :key="f.name" :label="label(f.label, f.name)" :mark="f.required">
            <t-textarea
              v-if="f.type === 'textarea'"
              v-model="stepForm[f.name]"
              :placeholder="f.placeholder"
              :autosize="{ minRows: 2, maxRows: 6 }"
              class="scroll-textarea"
            />
            <t-input v-else v-model="stepForm[f.name]" :placeholder="f.placeholder" />
          </t-form-item>
        </t-form>
        <t-form v-else-if="nextStep?.action === 'open_url' && nextStep.fields?.length && (!nextStep.wait || showCallbackInput)" label-width="90px">
          <t-form-item v-for="f in nextStep.fields" :key="f.name" :label="label(f.label, f.name)" :mark="f.required">
            <t-textarea
              v-model="stepForm[f.name]"
              :placeholder="f.placeholder"
              :autosize="{ minRows: 2, maxRows: 6 }"
              class="scroll-textarea"
            />
          </t-form-item>
        </t-form>

        <t-button theme="primary" block :loading="submitting" @click="submit">
          {{ submitLabel }}
        </t-button>
      </t-space>

      <!-- 第三步：授权成功 → 基本信息 / 模型列表 / 分组 -->
      <t-space v-else direction="vertical" style="width: 100%" size="large">
        <t-alert theme="success" :message="$t('accounts.successHint')" />
        <div>
          <div class="section-title">{{ $t('accounts.basicInfo') }}</div>
          <t-form label-width="90px">
            <t-form-item :label="$t('accounts.name')">
              <t-input v-model="newAccountName" :placeholder="wizardProfileName ? $t('accounts.namePh', { name: wizardProfileName }) : $t('accounts.namePhNone')" />
            </t-form-item>
          </t-form>
        </div>
        <div>
          <div class="section-title">
            {{ $t('accounts.modelsTitle') }}
            <t-link theme="primary" style="margin-left: 8px" @click="syncModels">{{ modelsSyncing ? $t('accounts.syncing') : $t('accounts.sync') }}</t-link>
          </div>
          <div v-if="wizardModels.length" class="model-list">
            <t-tag v-for="m in wizardModels" :key="m.id" variant="light-outline" style="margin: 0 6px 6px 0">{{ m.id }}</t-tag>
          </div>
          <span v-else class="hint">{{ $t('accounts.noModels') }}</span>
        </div>
        <div>
          <div class="section-title">{{ $t('accounts.groupsTitle') }}</div>
          <bind-select v-model="newAccountGroups" :options="wizardGroupOptions" :placeholder="$t('accounts.groupsPh')" />
        </div>
        <t-button theme="primary" block :loading="savingConfig" @click="finishWizard">{{ $t('accounts.finish') }}</t-button>
      </t-space>
    </t-dialog>


  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { RefreshIcon } from 'tdesign-icons-vue-next'
import { CheckIcon } from 'tdesign-icons-vue-next'
import { api } from '../api/client'
import AccountDetailDrawer from '../components/AccountDetailDrawer.vue'
import AccountEditDialog from '../components/AccountEditDialog.vue'
import AccountTestDrawer from '../components/AccountTestDrawer.vue'
import { fmtNum } from '../utils/format'
import DataTable from '../components/DataTable.vue'
import BindSelect from '../components/BindSelect.vue'
import { accountStatusDict, capabilityDict, dict, label, runStatusDict } from '../utils/dict'
import type { Account, AccountDetail, AuthMethod, GroupInfo, LoginResp, ModelInfo, NextStep, PluginInfo } from '../api/types'

const { t } = useI18n()

const plugins = ref<PluginInfo[]>([])
const accounts = ref<Account[]>([])
const groups = ref<GroupInfo[]>([])
const proxies = ref<{ ID: number; Scheme: string; Host: string; Port: number }[]>([])
const loading = ref(false)
const refreshingAll = ref(false)

// 编辑弹窗
const editVisible = ref(false)
const editRow = ref<Account | null>(null)

// 在线测试抽屉
const testVisible = ref(false)
const testRow = ref<Account | null>(null)

const addVisible = ref(false)
const wizardStep = ref<'select' | 'auth' | 'done'>('select')
const pluginName = ref('')
const methods = ref<AuthMethod[]>([])
const methodId = ref('')
const form = ref<Record<string, string>>({})
const nextStep = ref<NextStep | null>(null)
const stepForm = ref<Record<string, string>>({})
const submitting = ref(false)

// 向导第三步（授权成功后的配置）
const newAccountId = ref(0)
const newAccountName = ref('')
const newAccountGroups = ref<number[]>([])
const savingConfig = ref(false)
const wizardModels = ref<{ id: string }[]>([])
const modelsSyncing = ref(false)
const wizardProfileName = ref('')
let pollTimer: ReturnType<typeof setTimeout> | null = null

const selectedPluginLabel = computed(() => {
  const p = plugins.value.find((x) => x.name === pluginName.value)
  return p?.label || p?.name || ''
})

// 授权按钮文案：按登录方式形态给出（发送验证码 / 生成授权链接 / 授权）
const submitLabel = computed(() => {
  if (nextStep.value?.wait) return showCallbackInput.value ? t('accounts.submitCallback') : t('accounts.waitingAuth')
  if (nextStep.value) return t('accounts.submitNext')
  const types = currentFields.value.map((f) => f.type)
  if (types.includes('phone')) return t('accounts.sendOtp')
  if (!currentFields.value.length) return t('accounts.genAuthUrl')
  return t('accounts.authorize')
})

const columns = computed(() => [
  { colKey: 'display_name', title: t('accounts.account'), width: 150, ellipsis: true },
  { colKey: 'plugin', title: t('accounts.colPlugin'), width: 92, ellipsis: true, cell: (_h: any, { row }: any) => pluginLabel(row.plugin_id), align: 'center' },
  { colKey: 'group', title: t('accounts.groups'), width: 268, align: 'center' },
  { colKey: 'credits', title: t('accounts.credits'), width: 108, align: 'center' },
  { colKey: 'today_tokens', title: t('accounts.todayTokens'), width: 108, align: 'center' },
  { colKey: 'today_credits', title: t('accounts.todayCredits'), width: 102, align: 'center' },
  { colKey: 'expiry', title: t('accounts.expiry'), width: 132, align: 'center' },
  { colKey: 'status', title: t('accounts.status'), width: 84, align: 'center' },
  { colKey: 'schedule', title: t('accounts.schedule'), width: 86, align: 'center' },
  { colKey: 'last_refresh_at', title: t('accounts.lastRefresh'), width: 106, cell: (_h: any, { row }: any) => row.last_refresh_at ? timeAgo(row.last_refresh_at) : '-', align: 'center' },
  { colKey: 'op', title: t('common.colOp'), width: 186, align: 'center' },
])

// 相对时间：如 5分钟前 / 1天前 / 3个月前
function timeAgo(ts: string): string {
  const diff = Date.now() - new Date(ts.replace(' ', 'T')).getTime()
  const min = Math.floor(diff / 60000)
  if (min < 1) return t('common.justNow')
  if (min < 60) return t('common.minutesAgo', { n: min })
  const h = Math.floor(min / 60)
  if (h < 24) return t('common.hoursAgo', { n: h })
  const d = Math.floor(h / 24)
  if (d < 30) return t('common.daysAgo', { n: d })
  const mo = Math.floor(d / 30)
  if (mo < 12) return t('common.monthsAgo', { n: mo })
  return t('common.yearsAgo', { n: Math.floor(mo / 12) })
}

// 调度开关：active ↔ disabled（expired 需重新授权，不可直接开关）
async function toggleSchedule(row: Account) {
  if (row.status === 'active') {
    await api.post(`/admin/accounts/${row.id}/pause`)
    MessagePlugin.success(t('accounts.pausedSchedule'))
  } else {
    await api.post(`/admin/accounts/${row.id}/resume`)
    MessagePlugin.success(t('accounts.resumedSchedule'))
  }
  await loadAll()
}

// 插件品牌名映射：关联字段统一显示品牌而非 id
function pluginLabel(pluginID: number): string {
  const p = plugins.value.find((x) => x.id === pluginID)
  return p?.label || p?.name || `#${pluginID}`
}

// ---------- 暂停展示 ----------

// 自动暂停（429 限时 / 402 手动）判定：active 但 paused_until 在未来
function pausedInfo(row: Account): string {
  if (row.status !== 'active' || !row.paused_until) return ''
  const until = new Date(row.paused_until).getTime()
  if (!until || until <= Date.now()) return ''
  const untilText = row.paused_until.replace('T', ' ').slice(0, 19)
  return until - Date.now() > 365 * 24 * 3600 * 1000
    ? t('accounts.pausedManual', { reason: row.pause_reason || t('accounts.autoPause') })
    : t('accounts.pausedRateLimited', { until: untilText, reason: row.pause_reason || '' })
}

function pausedLabel(row: Account): string {
  const until = row.paused_until ? new Date(row.paused_until).getTime() : 0
  return until - Date.now() > 365 * 24 * 3600 * 1000 ? t('accounts.pausedManualTag') : t('accounts.pausedRateLimitedTag')
}

// ---------- 账号详情 ----------

const detailVisible = ref(false)
const detail = ref<AccountDetail | null>(null)




// ---------- 动态渲染块（插件声明的 ProfileSection） ----------








async function openDetail(id: number) {
  detail.value = await api.get<AccountDetail>(`/admin/accounts/${id}/detail`)
  detailVisible.value = true
}

const currentFields = computed(() => methods.value.find((m) => m.id === methodId.value)?.fields ?? [])
const currentMethod = computed(() => methods.value.find((m) => m.id === methodId.value))
const pluginGroups = computed(() => groups.value.filter((g) => g.plugin === pluginName.value))

// 管理界面是否本机访问（决定 auto_wait 的走向：本机 127.0.0.1 回调可达）
const isLocal = ['localhost', '127.0.0.1', '::1'].includes(location.hostname)

// wait 步骤是否渲染回调粘贴框：按插件声明的 callback 模式，未声明按 next 下发推断
const showCallbackInput = computed(() => {
  switch (currentMethod.value?.callback) {
    case 'auto': return false
    case 'wait': return true
    case 'auto_wait': return !isLocal
    default: return !!nextStep.value?.fields?.length
  }
})

// 是否自动轮询等待插件侧完成（auto_wait 非本机时等用户手动提交）
const autoPolling = computed(() => {
  if (!nextStep.value?.wait) return false
  switch (currentMethod.value?.callback) {
    case 'auto': return true
    case 'wait': return false
    case 'auto_wait': return isLocal
    default: return true
  }
})

// 打开编辑弹窗 / 在线测试抽屉：父级只记「对哪个账号」，其余状态在子组件里
function openEdit(row: Account) {
  editRow.value = row
  editVisible.value = true
}

function openTest(row: Account) {
  testRow.value = row
  testVisible.value = true
}

async function loadAll() {
  loading.value = true
  try {
    const [p, a, g, px] = await Promise.all([
      api.get<{ plugins: PluginInfo[] }>('/admin/plugins'),
      api.get<{ accounts: Account[] }>('/admin/accounts'),
      api.get<{ groups: GroupInfo[] }>('/admin/groups'),
      api.get<{ proxies: typeof proxies.value }>('/admin/proxies'),
    ])
    plugins.value = p.plugins ?? []
    accounts.value = a.accounts ?? []
    groups.value = g.groups ?? []
    proxies.value = px.proxies ?? []
  } finally {
    loading.value = false
  }
}

const proxyOptions = computed(() =>
  proxies.value.map((px) => ({ value: px.ID, label: `${px.Scheme}://${px.Host}:${px.Port}` })),
)
const wizardGroupOptions = computed(() =>
  pluginGroups.value.map((g) => ({ value: g.id, label: `${g.name} (${g.plugin_label || g.plugin})` })),
)






function openAdd() {
  addVisible.value = true
  wizardStep.value = 'select'
  nextStep.value = null
  form.value = {}
  stepForm.value = {}
  wizardModels.value = []
  stopPolling()
}

// 第一步点选客户端 → 进入授权
function choosePlugin(p: PluginInfo) {
  pluginName.value = p.name
  wizardStep.value = 'auth'
  loadMethods(p.name)
}

async function loadMethods(name: string) {
  const resp = await api.get<{ auth_methods: AuthMethod[] }>(`/admin/plugins/${name}/auth-methods`)
  methods.value = resp.auth_methods ?? []
  methodId.value = methods.value[0]?.id ?? ''
  nextStep.value = null
  stepForm.value = {}
  stopPolling()
}

watch(methodId, () => {
  form.value = {}
  nextStep.value = null
  stepForm.value = {}
  stopPolling()
})

function onDrop(e: DragEvent, field: string) {
  const file = e.dataTransfer?.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = () => {
    form.value[field] = String(reader.result ?? '')
    MessagePlugin.success(t('accounts.loadedN', { name: file.name }))
  }
  reader.readAsText(file)
}

function copyText(text: string) {
  navigator.clipboard.writeText(text)
  MessagePlugin.success(t('common.copied'))
}

function openURL(url: string) {
  window.open(url, '_blank')
}

async function submit() {
  if (!pluginName.value || !methodId.value) return
  submitting.value = true
  stopPolling() // 手动提交优先于轮询，避免并发打插件
  try {
    const payload = nextStep.value
      ? { plugin: pluginName.value, method_id: methodId.value, form: stepForm.value, state: nextStep.value.state ?? '' }
      : { plugin: pluginName.value, method_id: methodId.value, form: form.value, state: '' }
    const resp = await api.post<LoginResp>('/admin/accounts/login', payload)
    if (resp.done) {
      enterDoneStep(resp.account_id ?? 0)
    } else {
      nextStep.value = resp.next ?? null
      stepForm.value = {}
      // 浏览器授权：自动打开页面；能自动回调才轮询，否则等用户粘贴提交
      const step = nextStep.value
      if (step?.wait) {
        if (step.url) openURL(step.url)
        if (autoPolling.value) startPolling()
      }
    }
  } catch (e: any) {
    MessagePlugin.error(String(e.message))
    stopPolling()
  } finally {
    submitting.value = false
  }
}

// enterDoneStep 授权成功：进入向导第三步（名称 / 模型列表 / 分组）
function enterDoneStep(accountID: number) {
  stopPolling()
  newAccountId.value = accountID
  newAccountName.value = ''
  newAccountGroups.value = []
  wizardModels.value = []
  wizardProfileName.value = ''
  wizardStep.value = 'done'
  // 默认展示名称（详情接口取 profile）
  api.get<AccountDetail>(`/admin/accounts/${accountID}/detail`).then((d) => {
    wizardProfileName.value = d.display_name || (d.profile?.displayName ?? '')
  }).catch(() => {})
  syncModels()
}

// syncModels 同步客户端模型目录（账号凭据）
async function syncModels() {
  if (!newAccountId.value || modelsSyncing.value) return
  modelsSyncing.value = true
  try {
    const resp = await api.get<{ models: { id: string }[] | null }>(`/admin/accounts/${newAccountId.value}/models?refresh=1`)
    wizardModels.value = resp.models ?? []
  } catch (e: any) {
    MessagePlugin.warning(t('accounts.syncFailed', { msg: e.message }))
  } finally {
    modelsSyncing.value = false
  }
}

// finishWizard 完成向导：保存名称与分组
async function finishWizard() {
  if (newAccountId.value) {
    savingConfig.value = true
    try {
      const body: Record<string, unknown> = {}
      if (newAccountName.value) body.display_name = newAccountName.value
      body.group_ids = newAccountGroups.value
      await api.put(`/admin/accounts/${newAccountId.value}`, body)
    } finally {
      savingConfig.value = false
    }
  }
  addVisible.value = false
  MessagePlugin.success(t('accounts.added'))
  await loadAll()
}

// auto 回调：2 秒一次无表单提交，插件侧等浏览器跳回调
function startPolling() {
  stopPolling()
  pollTimer = setTimeout(async () => {
    if (!autoPolling.value) return
    try {
      const resp = await api.post<LoginResp>('/admin/accounts/login', {
        plugin: pluginName.value, method_id: methodId.value,
        form: {}, state: nextStep.value?.state ?? '',
      })
      if (resp.done) {
        enterDoneStep(resp.account_id ?? 0)
      } else if (resp.next?.wait && autoPolling.value) {
        nextStep.value = resp.next
        startPolling()
      }
    } catch {
      // 网络抖动继续等
      startPolling()
    }
  }, 2000)
}

function stopPolling() {
  if (pollTimer) {
    clearTimeout(pollTimer)
    pollTimer = null
  }
}

// 行内分组：账号插件对应的分组
function groupsOf(pluginID: number): GroupInfo[] {
  return groups.value.filter((g) => g.plugin_id === pluginID)
}

function groupName(id: number): string {
  return groups.value.find((g) => g.id === id)?.name ?? `#${id}`
}

// 单个分组增减（tag 关闭 / 弹层勾选）
async function toggleGroup(row: Account, groupID: number, add: boolean) {
  const cur = row.group_ids ?? []
  const next = add ? [...cur, groupID] : cur.filter((id) => id !== groupID)
  row.group_ids = next // 乐观更新
  try {
    await api.put(`/admin/accounts/${row.id}`, { group_ids: next })
  } catch (e: any) {
    row.group_ids = cur
    MessagePlugin.error(e.message)
  }
}

async function refresh(id: number) {
  try {
    await api.post(`/admin/accounts/${id}/refresh`)
    MessagePlugin.success(t('common.refreshed'))
    await loadAll()
  } catch (e: any) {
    MessagePlugin.error(e.message)
  }
}

// 一键刷新：服务端并发刷所有账号（含积分快照），返回逐条失败明细
async function refreshAll() {
  if (!accounts.value.length) return
  refreshingAll.value = true
  try {
    const r = await api.post<{ total: number; refreshed: number; failed: number; items: { name: string; message: string }[] }>(
      '/admin/accounts/refresh-all',
    )
    MessagePlugin.success(t('accounts.refreshAllDone', { ok: r.refreshed, total: r.total }))
    if (r.failed > 0) {
      const head = r.items.slice(0, 3).map((i) => `${i.name || '#'}: ${i.message}`).join('；')
      MessagePlugin.warning(`${r.failed} 个账号刷新失败：${head}`)
    }
    await loadAll()
  } catch (e: any) {
    MessagePlugin.error(e.message)
  } finally {
    refreshingAll.value = false
  }
}

// 快过期积分：7 天内有到期额度时高亮，避免白白作废
function expiringSoon(c?: { expiring?: number }): boolean {
  return !!c?.expiring && c.expiring > 0
}

function expiryLabel(c?: { next_expiry?: string; next_left?: number }): string {
  if (!c?.next_expiry) return '-'
  const at = new Date(c.next_expiry.replace(' ', 'T'))
  const days = Math.ceil((at.getTime() - Date.now()) / 86400000)
  const left = c.next_left ? ` · ${Math.round(c.next_left)}` : ''
  if (days <= 0) return `已到期${left}`
  return `${days} 天后${left}`
}

// 今日用量列：大数压缩展示（12.3K / 1.2M）
function fmtCompact(n?: number): string {
  const v = n || 0
  if (v < 1000) return String(v)
  if (v < 1000000) return `${(v / 1000).toFixed(1).replace(/\.0$/, '')}K`
  return `${(v / 1000000).toFixed(2)}M`
}

function fmtCredit(n?: number): string {
  const v = n || 0
  if (!v) return '0'
  return Number.isInteger(v) ? String(v) : v.toFixed(2).replace(/0+$/, '').replace(/\.$/, '')
}

async function remove(id: number) {
  await api.del(`/admin/accounts/${id}`)
  await loadAll()
}

onBeforeUnmount(stopPolling)
onMounted(loadAll)
</script>

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}
.expiry { font-variant-numeric: tabular-nums; color: var(--cph-text-2); }
.expiry.soon { color: var(--cph-warning); font-weight: 600; }
.today {
  color: var(--cph-text-1);
  font-weight: 600;
}
.today .est {
  margin-left: 4px;
  font-style: normal;
  font-size: 10px;
  color: var(--cph-text-3);
  border: 1px solid var(--cph-border-strong);
  border-radius: 6px;
  padding: 0 4px;
}
.tok-detail { min-width: 200px }
.tok-detail-title { font-weight: 700; margin-bottom: 8px }
.tok-detail-row {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  padding: 2px 0;
}
/* 账号名称：点击开详情，移入高亮 */
.acct-name {
  cursor: pointer;
  transition: color 0.15s ease;
}
.acct-name:hover {
  color: var(--td-brand-color);
  text-decoration: underline;
}
.tab-body {
  padding: 12px 4px;
}
.login-url {
  margin-top: 6px;
  display: flex;
  gap: 10px;
  align-items: flex-start;
}
.login-url-text {
  font-family: monospace;
  font-size: 12px;
  word-break: break-all;
  flex: 1;
}
.mode-hint {
  margin-top: 4px;
  font-size: 12px;
  opacity: 0.85;
}
.section-title {
  font-weight: 600;
  margin-bottom: 8px;
}
.credit-summary {
  display: flex;
  gap: 16px;
  margin-bottom: 8px;
  font-size: 13px;
  color: var(--td-text-color-secondary);
  font-variant-numeric: tabular-nums;
}
.credit-cell {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}
.model-list {
  max-height: 160px;
  overflow-y: auto;
}
.hint {
  color: var(--td-text-color-placeholder);
  font-size: 12px;
}
.test-answer {
  padding: 12px;
  white-space: pre-wrap;
  word-break: break-word;
  background: var(--td-bg-color-secondarycontainer);
  border-radius: 6px;
}
.test-logs {
  padding: 8px 12px;
  font-family: monospace;
  font-size: 12px;
  color: var(--td-text-color-secondary);
  background: var(--td-bg-color-container-hover);
  border-radius: 6px;
}
.test-log-line {
  word-break: break-all;
  line-height: 1.7;
}
.wizard-back {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.wizard-client {
  font-weight: 600;
}
.client-card {
  border: 1px solid var(--td-component-border);
  border-radius: var(--td-radius-medium);
  padding: 16px 12px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}
.client-card:hover {
  border-color: var(--td-brand-color);
  box-shadow: var(--td-shadow-1);
}
.client-icon {
  width: 44px;
  height: 44px;
  margin: 0 auto 8px;
  border-radius: 50%;
  background: var(--td-brand-color-light);
  color: var(--td-brand-color);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 700;
  overflow: hidden;
}
.client-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.client-name {
  font-weight: 600;
  margin-bottom: 6px;
}
.client-caps {
  min-height: 22px;
}
.drop-zone {
  position: relative;
  width: 100%;
}
.drop-hint {
  position: absolute;
  right: 8px;
  bottom: 6px;
  font-size: 11px;
  color: var(--td-text-color-placeholder);
  pointer-events: none;
}
.scroll-textarea :deep(.t-textarea__inner) {
  max-height: 320px;
  overflow-y: auto;
}

/* 行内分组 tag：动态增减 */
.group-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  align-items: center;
}
.group-add {
  cursor: pointer;
  min-width: 22px;
  text-align: center;
}
.group-picker {
  min-width: 160px;
  max-height: 240px;
  overflow-y: auto;
}
.group-picker-empty {
  padding: 8px 12px;
  font-size: 12px;
  color: var(--td-text-color-placeholder);
}
.group-picker-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 6px 12px;
  font-size: 13px;
  color: var(--td-text-color-primary);
  cursor: pointer;
  white-space: nowrap;
  transition: background-color 0.15s ease;
}
.group-picker-item:hover {
  background: var(--td-bg-color-secondarycontainer);
}
.group-picker-item.active {
  color: var(--td-brand-color);
}
</style>
