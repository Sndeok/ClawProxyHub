<template>
  <div class="page">
    <div class="page-actions">
    </div>

    <div class="settings-grid">

    <t-card :title="$t('settings.gateway')" class="card" :bordered="false">
      <t-form label-width="140px">
        <t-form-item :label="$t('settings.firstTokenTimeout')" :help="$t('settings.firstTokenTimeoutHelp')">
          <t-input-number v-model="gwForm.first_event_timeout" :min="5" :max="3600" theme="column" style="max-width: 160px; width: 100%" />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" :loading="savingGw" @click="saveGw">{{ $t('common.save') }}</t-button>
        </t-form-item>
      </t-form>
    </t-card>

    <t-card :title="$t('settings.logs')" class="card" :bordered="false">
      <t-form label-width="140px">
        <t-form-item :label="$t('settings.logRetention')" :help="$t('settings.logRetentionHelp')">
          <t-input-number v-model="logForm.log_retention_days" :min="0" :max="3650" theme="column" style="max-width: 160px; width: 100%" />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" :loading="savingLog" @click="saveLog">{{ $t('common.save') }}</t-button>
          <span class="form-hint">{{ $t('settings.logRetentionHint') }}</span>
        </t-form-item>
      </t-form>
    </t-card>

    <t-card :title="$t('settings.network')" class="card" :bordered="false">
      <t-form label-width="140px">
        <t-form-item :label="$t('settings.marketplaceUrl')" :help="$t('settings.marketplaceHelp')">
          <t-input v-model="netForm.marketplace_url" :placeholder="$t('settings.marketplacePh')" style="max-width: 520px; width: 100%" />
        </t-form-item>
        <t-form-item :label="$t('settings.marketProxy')" :help="$t('settings.marketProxyHelp')">
          <t-input v-model="netForm.market_proxy" :placeholder="$t('settings.marketProxyPh')" style="max-width: 520px; width: 100%" />
        </t-form-item>
        <t-form-item :label="$t('settings.githubProxy')" :help="$t('settings.githubProxyHelp')">
          <t-input v-model="netForm.github_proxy" placeholder="https://ghproxy.com" style="max-width: 520px; width: 100%" />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" :loading="savingNet" @click="saveNet">{{ $t('common.save') }}</t-button>
          <t-button variant="outline" :loading="testingConn" style="margin-left: 8px" @click="testConn">
            {{ $t('settings.testConn') }}
          </t-button>
        </t-form-item>
      </t-form>
    </t-card>

    <t-card :title="$t('settings.outbound')" class="card" :bordered="false">
      <div class="plugin-tabs">
        <t-radio-group v-model="pluginName" variant="default-filled" @change="loadPluginSettings">
          <t-radio-button v-for="p in plugins" :key="p.name" :value="p.name">
            {{ p.label || p.name }}
          </t-radio-button>
        </t-radio-group>
        <span class="form-hint">{{ $t('settings.outboundPerPlugin') }}</span>
      </div>
      <t-form v-if="pluginName" label-width="140px">
        <t-form-item :label="$t('settings.outboundUA')" :help="$t('settings.outboundUAHelp')">
          <t-input v-model="outForm.user_agent" :placeholder="outDefaults.user_agent" style="max-width: 520px; width: 100%" clearable />
        </t-form-item>
        <t-form-item :label="$t('settings.clientName')" :help="$t('settings.clientNameHelp')">
          <t-input v-model="outForm.client_name" :placeholder="outDefaults.client_name" style="max-width: 240px; width: 100%" clearable />
        </t-form-item>
        <t-form-item :label="$t('settings.clientVersion')" :help="$t('settings.clientVersionHelp')">
          <t-input v-model="outForm.client_version" :placeholder="outDefaults.client_version" style="max-width: 240px; width: 100%" clearable />
        </t-form-item>
        <t-form-item :label="$t('settings.cliVersion')" :help="$t('settings.cliVersionHelp')">
          <t-input v-model="outForm.cli_version" :placeholder="outDefaults.cli_version" style="max-width: 240px; width: 100%" clearable />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" :loading="savingOut" @click="saveOutbound">{{ $t('common.save') }}</t-button>
          <t-button variant="outline" style="margin-left: 8px" @click="loadPluginSettings">
            {{ $t('common.refresh') }}
          </t-button>
          <span class="form-hint">{{ $t('settings.outboundHint') }}</span>
        </t-form-item>
      </t-form>
      <t-alert v-else theme="info" :message="$t('settings.outboundNoPlugin')" />
    </t-card>

    <t-card :title="$t('settings.sticky')" class="card" :bordered="false">
      <t-form label-width="140px">
        <t-form-item :label="$t('settings.stickySwitch')" :help="$t('settings.stickyHelp')">
          <t-switch v-model="stickyOn" />
          <span class="form-hint">{{ stickyOn ? $t('settings.stickyOn') : $t('settings.stickyOff') }}</span>
        </t-form-item>
        <t-form-item :label="$t('settings.stickyTTL')" :help="$t('settings.stickyTTLHelp')">
          <t-input v-model="stickyForm.sticky_ttl" placeholder="30m" style="max-width: 160px; width: 100%" :disabled="!stickyOn" />
          <span class="form-hint">{{ $t('settings.stickyTTLHint') }}</span>
        </t-form-item>
        <t-form-item :label="$t('settings.stickyClean')" :help="$t('settings.stickyCleanHelp')">
          <t-input v-model="stickyForm.sticky_cleanup_period" placeholder="5m" style="max-width: 160px; width: 100%" :disabled="!stickyOn" />
          <span class="form-hint">{{ $t('settings.stickyCleanHint') }}</span>
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" :loading="savingSticky" @click="saveSticky">{{ $t('common.save') }}</t-button>
          <span class="form-hint">{{ $t('settings.stickyApply') }}</span>
        </t-form-item>
      </t-form>
    </t-card>

    <t-card :title="$t('settings.adminPassword')" class="card" :bordered="false">
      <t-form label-width="140px">
        <t-form-item :label="$t('settings.newPassword')" mark>
          <t-input v-model="pwForm.password" type="password" :placeholder="$t('settings.passwordPh')" style="max-width: 280px; width: 100%" />
        </t-form-item>
        <t-form-item :label="$t('settings.confirmPassword')" mark>
          <t-input v-model="pwForm.confirm" type="password" style="max-width: 280px; width: 100%" />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" :loading="savingPw" @click="savePw">{{ $t('settings.changePassword') }}</t-button>
        </t-form-item>
      </t-form>
    </t-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { api } from '../api/client'

const { t } = useI18n()

const gwForm = reactive({ first_event_timeout: 90 })
const logForm = reactive({ log_retention_days: 0 })
const netForm = reactive({ github_proxy: '', marketplace_url: '', market_proxy: '' })
// 出站标识：按插件单独配置（键名与插件设置 schema 一致）
const outForm = reactive({ user_agent: '', client_name: '', client_version: '', cli_version: '' })
// 插件内置默认值（来自 manifest 的 settings schema default），显示为占位符
const outDefaults = reactive({ user_agent: '', client_name: '', client_version: '', cli_version: '' })
const plugins = ref<{ name: string; label: string }[]>([])
const pluginName = ref('')
const stickyForm = reactive({ sticky_ttl: '30m', sticky_cleanup_period: '5m' })
const stickyOn = ref(true)
const savingOut = ref(false)
const savingSticky = ref(false)

const testingConn = ref(false)
const pwForm = reactive({ password: '', confirm: '' })
const savingGw = ref(false)
const savingLog = ref(false)
const savingNet = ref(false)
const savingPw = ref(false)

const outKeys = ['user_agent', 'client_name', 'client_version', 'cli_version'] as const

// 拉插件列表（默认选第一个）
async function loadPlugins() {
  const r = await api.get<{ plugins: { name: string; label?: string }[] }>('/admin/plugins')
  plugins.value = (r.plugins ?? []).map((p) => ({ name: p.name, label: p.label || p.name }))
  if (!pluginName.value && plugins.value.length) {
    pluginName.value = plugins.value[0].name
    await loadPluginSettings()
  }
}

// 读单个插件的设置：值 + schema 默认值（默认值作为输入框占位符）
async function loadPluginSettings() {
  if (!pluginName.value) return
  try {
    const r = await api.get<{ schema: any; values: Record<string, string> }>(
      `/admin/plugins/${pluginName.value}/settings`,
    )
    const values = r.values ?? {}
    const props = r.schema?.properties ?? {}
    for (const k of outKeys) {
      outForm[k] = values[k] ?? ''
      outDefaults[k] = props[k]?.default ?? ''
    }
  } catch (e: any) {
    MessagePlugin.error(e.message)
  }
}

async function load() {
  const r = await api.get<{
    settings: {
      first_event_timeout: number
      log_retention_days?: number
      github_proxy?: string
      marketplace_url?: string
      market_proxy?: string
      outbound_user_agent?: string
      outbound_client_name?: string
      outbound_client_version?: string
      outbound_cli_version?: string
      sticky_ttl?: string
      sticky_cleanup_period?: string
      [key: string]: unknown
    }
  }>('/admin/settings')
  gwForm.first_event_timeout = r.settings?.first_event_timeout ?? 90
  logForm.log_retention_days = r.settings?.log_retention_days ?? 0
  netForm.github_proxy = r.settings?.github_proxy ?? ''
  netForm.marketplace_url = r.settings?.marketplace_url ?? ''
  netForm.market_proxy = r.settings?.market_proxy ?? ''
  stickyForm.sticky_ttl = r.settings?.sticky_ttl || '30m'
  stickyForm.sticky_cleanup_period = r.settings?.sticky_cleanup_period || '5m'
}

// 用当前输入框（未保存）的值试拉一次市场索引，确认代理是否通
async function testConn() {
  testingConn.value = true
  try {
    const r = await api.post<{ ok: boolean; error?: string; count?: number; elapsed_ms?: number; url?: string }>(
      '/admin/settings/test-market',
      { marketplace_url: netForm.marketplace_url.trim(), market_proxy: netForm.market_proxy.trim() },
    )
    if (r.ok) {
      MessagePlugin.success(t('settings.testOk', { n: r.count ?? 0, ms: r.elapsed_ms ?? 0 }))
    } else {
      MessagePlugin.error(`${t('settings.testFail')}: ${r.error ?? ''}`)
    }
  } catch (e: any) {
    MessagePlugin.error(e.message)
  } finally {
    testingConn.value = false
  }
}

// 后端 PUT /admin/settings 是整体覆盖：所有卡片共用同一份 payload，
// 避免某个卡片漏字段导致另一卡片的配置被清空。
function settingsPayload() {
  return {
    first_event_timeout: gwForm.first_event_timeout,
    log_retention_days: logForm.log_retention_days,
    github_proxy: netForm.github_proxy.trim(),
    marketplace_url: netForm.marketplace_url.trim(),
    market_proxy: netForm.market_proxy.trim(),
  }
}

async function saveGw() {
  savingGw.value = true
  try {
    await api.put('/admin/settings', settingsPayload())
    MessagePlugin.success(t('settings.saved'))
  } catch (e: any) {
    MessagePlugin.error(e.message)
  } finally {
    savingGw.value = false
  }
}

async function saveLog() {
  savingLog.value = true
  try {
    await api.put('/admin/settings', settingsPayload())
    MessagePlugin.success(t('settings.saved'))
  } catch (e: any) {
    MessagePlugin.error(e.message)
  } finally {
    savingLog.value = false
  }
}

// 出站标识与粘性策略走同一个 PUT /admin/settings，未改的字段按当前值回传避免被清空
async function putSettings(patch: Record<string, unknown>) {
  const r = await api.get<{ settings: Record<string, unknown> }>('/admin/settings')
  const merged = { ...r.settings, ...patch }
  await api.put('/admin/settings', merged)
}

// 保存到该插件自己的设置里；键留空 = 用插件内置默认
async function saveOutbound() {
  if (!pluginName.value) return
  savingOut.value = true
  try {
    const r = await api.get<{ values: Record<string, string> }>(
      `/admin/plugins/${pluginName.value}/settings`,
    )
    const merged: Record<string, string> = { ...(r.values ?? {}) }
    for (const k of outKeys) merged[k] = outForm[k].trim()
    await api.put(`/admin/plugins/${pluginName.value}/settings`, { values: merged })
    MessagePlugin.success(t('settings.saved'))
    await loadPluginSettings()
  } catch (e: any) {
    MessagePlugin.error(e.message)
  } finally {
    savingOut.value = false
  }
}

async function saveSticky() {
  savingSticky.value = true
  try {
    await putSettings({
      sticky_ttl: stickyOn.value ? stickyForm.sticky_ttl.trim() : '1m',
      sticky_cleanup_period: stickyForm.sticky_cleanup_period.trim() || '5m',
    })
    MessagePlugin.success(t('settings.saved'))
    await load()
  } catch (e: any) {
    MessagePlugin.error(e.message)
  } finally {
    savingSticky.value = false
  }
}

async function saveNet() {
  savingNet.value = true
  try {
    await api.put('/admin/settings', settingsPayload())
    MessagePlugin.success(t('settings.saved'))
  } catch (e: any) {
    MessagePlugin.error(e.message)
  } finally {
    savingNet.value = false
  }
}

async function savePw() {
  if (pwForm.password.length < 6) {
    MessagePlugin.warning(t('settings.errPassword'))
    return
  }
  if (pwForm.password !== pwForm.confirm) {
    MessagePlugin.warning(t('settings.errConfirm'))
    return
  }
  savingPw.value = true
  try {
    await api.post('/admin/password', { password: pwForm.password })
    MessagePlugin.success(t('settings.passwordChanged'))
    pwForm.password = ''
    pwForm.confirm = ''
  } catch (e: any) {
    MessagePlugin.error(e.message)
  } finally {
    savingPw.value = false
  }
}

onMounted(async () => {
  await Promise.all([load(), loadPlugins()])
})
</script>

<style scoped>
/* 双列自适应：窄卡片并排铺满整行，宽卡片（网络/出站标识）独占一行 */
/* 所有卡片等宽自适应：一行放得下就并排，放不下自动换行，右侧不留空 */
.settings-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: flex-start;
}
.settings-grid > * {
  flex: 1 1 420px;
  min-width: 0;
  margin-bottom: 0 !important;
}
.plugin-tabs {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 14px;
}
.card { margin-bottom: 0; }
.card { max-width: 720px; margin-bottom: 16px }
.form-hint {
  margin-left: 12px;
  font-size: 12px;
  color: var(--td-text-color-secondary);
}
</style>
