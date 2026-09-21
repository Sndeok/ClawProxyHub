<template>
  <div class="page">
    <div class="page-header">
      
    </div>

    <t-card :title="$t('settings.gateway')" class="card" :bordered="false">
      <t-form label-width="140px">
        <t-form-item :label="$t('settings.firstTokenTimeout')" :help="$t('settings.firstTokenTimeoutHelp')">
          <t-input-number v-model="gwForm.first_event_timeout" :min="5" :max="3600" theme="column" style="width: 160px" />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" :loading="savingGw" @click="saveGw">{{ $t('common.save') }}</t-button>
        </t-form-item>
      </t-form>
    </t-card>

    <t-card :title="$t('settings.logs')" class="card" :bordered="false">
      <t-form label-width="140px">
        <t-form-item :label="$t('settings.logRetention')" :help="$t('settings.logRetentionHelp')">
          <t-input-number v-model="logForm.log_retention_days" :min="0" :max="3650" theme="column" style="width: 160px" />
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
          <t-input v-model="netForm.marketplace_url" :placeholder="$t('settings.marketplacePh')" style="width: 520px" />
        </t-form-item>
        <t-form-item :label="$t('settings.marketProxy')" :help="$t('settings.marketProxyHelp')">
          <t-input v-model="netForm.market_proxy" :placeholder="$t('settings.marketProxyPh')" style="width: 520px" />
        </t-form-item>
        <t-form-item :label="$t('settings.githubProxy')" :help="$t('settings.githubProxyHelp')">
          <t-input v-model="netForm.github_proxy" placeholder="https://ghproxy.com" style="width: 520px" />
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
      <t-form label-width="140px">
        <t-form-item :label="$t('settings.outboundUA')" :help="$t('settings.outboundUAHelp')">
          <t-input v-model="outForm.outbound_user_agent" :placeholder="outPlaceholder.ua" style="width: 520px" />
        </t-form-item>
        <t-form-item :label="$t('settings.clientName')" :help="$t('settings.clientNameHelp')">
          <t-input v-model="outForm.outbound_client_name" :placeholder="outPlaceholder.name" style="width: 240px" />
        </t-form-item>
        <t-form-item :label="$t('settings.clientVersion')" :help="$t('settings.clientVersionHelp')">
          <t-input v-model="outForm.outbound_client_version" :placeholder="outPlaceholder.version" style="width: 240px" />
        </t-form-item>
        <t-form-item :label="$t('settings.cliVersion')" :help="$t('settings.cliVersionHelp')">
          <t-input v-model="outForm.outbound_cli_version" :placeholder="outPlaceholder.cli" style="width: 240px" />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" :loading="savingOut" @click="saveOutbound">{{ $t('common.save') }}</t-button>
          <span class="form-hint">{{ $t('settings.outboundHint') }}</span>
        </t-form-item>
      </t-form>
    </t-card>

    <t-card :title="$t('settings.sticky')" class="card" :bordered="false">
      <t-form label-width="140px">
        <t-form-item :label="$t('settings.stickySwitch')" :help="$t('settings.stickyHelp')">
          <t-switch v-model="stickyOn" />
          <span class="form-hint">{{ stickyOn ? $t('settings.stickyOn') : $t('settings.stickyOff') }}</span>
        </t-form-item>
        <t-form-item :label="$t('settings.stickyTTL')" :help="$t('settings.stickyTTLHelp')">
          <t-input v-model="stickyForm.sticky_ttl" placeholder="30m" style="width: 160px" :disabled="!stickyOn" />
          <span class="form-hint">{{ $t('settings.stickyTTLHint') }}</span>
        </t-form-item>
        <t-form-item :label="$t('settings.stickyClean')" :help="$t('settings.stickyCleanHelp')">
          <t-input v-model="stickyForm.sticky_cleanup_period" placeholder="5m" style="width: 160px" :disabled="!stickyOn" />
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
          <t-input v-model="pwForm.password" type="password" :placeholder="$t('settings.passwordPh')" style="width: 280px" />
        </t-form-item>
        <t-form-item :label="$t('settings.confirmPassword')" mark>
          <t-input v-model="pwForm.confirm" type="password" style="width: 280px" />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" :loading="savingPw" @click="savePw">{{ $t('settings.changePassword') }}</t-button>
        </t-form-item>
      </t-form>
    </t-card>
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
const outForm = reactive({
  outbound_user_agent: '', outbound_client_name: '',
  outbound_client_version: '', outbound_cli_version: '',
})
// 占位符展示插件内置默认值（留空即用它们）
const outPlaceholder = {
  ua: 'WorkBuddy/5.5.4 WorkBuddy/5.5.4 CLI/2.137.1',
  name: 'WorkBuddy', version: '5.5.4', cli: '2.137.1',
}
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
  outForm.outbound_user_agent = r.settings?.outbound_user_agent ?? ''
  outForm.outbound_client_name = r.settings?.outbound_client_name ?? ''
  outForm.outbound_client_version = r.settings?.outbound_client_version ?? ''
  outForm.outbound_cli_version = r.settings?.outbound_cli_version ?? ''
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

async function saveOutbound() {
  savingOut.value = true
  try {
    await putSettings({
      outbound_user_agent: outForm.outbound_user_agent.trim(),
      outbound_client_name: outForm.outbound_client_name.trim(),
      outbound_client_version: outForm.outbound_client_version.trim(),
      outbound_cli_version: outForm.outbound_cli_version.trim(),
    })
    MessagePlugin.success(t('settings.saved'))
    await load()
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

onMounted(load)
</script>

<style scoped>
.card { max-width: 720px; margin-bottom: 16px }
.form-hint {
  margin-left: 12px;
  font-size: 12px;
  color: var(--td-text-color-secondary);
}
</style>
