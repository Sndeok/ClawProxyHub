<template>
    <!-- 编辑账号：改名 / 绑分组 / 绑代理 / 同步模型 -->
    <t-dialog :visible="visible" @update:visible="(v: boolean) => emit('update:visible', v)" :header="$t('accounts.editTitle')" :confirm-btn="{ loading: editSaving }" width="640px" @confirm="submitEdit">
      <t-form v-if="row" label-width="90px">
        <t-form-item :label="$t('accounts.name')">
          <t-input v-model="editName" :placeholder="$t('accounts.namePh')" clearable />
        </t-form-item>
        <t-form-item :label="$t('accounts.groupsTitle')">
          <bind-select v-model="editGroups" :options="editGroupOptions" :placeholder="$t('accounts.groupsPh')" />
        </t-form-item>
        <t-form-item :label="$t('accounts.proxyTitle')">
          <bind-select v-model="editProxies" :options="proxyOptions" :placeholder="$t('accounts.proxyPh')" />
        </t-form-item>
        <t-form-item :label="$t('accounts.modelsTitle')">
          <div style="width: 100%">
            <t-link theme="primary" @click="editSyncModels">{{ editSyncing ? $t('accounts.syncing') : $t('accounts.sync') }}</t-link>
            <div v-if="editModels.length" class="model-list" style="margin-top: 8px">
              <t-tag v-for="m in editModels" :key="m.id" closable variant="light-outline" style="margin: 0 6px 6px 0" @close="editModels = editModels.filter((x) => x.id !== m.id)">{{ m.id }}</t-tag>
            </div>
            <span v-else class="hint">{{ $t('accounts.noModels') }}</span>
          </div>
        </t-form-item>
      </t-form>
    </t-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { api } from '../api/client'
import BindSelect from './BindSelect.vue'
import type { Account, AccountDetail, GroupInfo, ModelInfo } from '../api/types'

const props = defineProps<{
  visible: boolean
  row: Account | null
  groups: GroupInfo[]
  proxies: { ID: number; Scheme: string; Host: string; Port: number }[]
}>()
const emit = defineEmits<{ 'update:visible': [boolean]; saved: [] }>()
const { t } = useI18n()

// 代理下拉：直接用父级传入的代理列表（与列表页同一份数据，不再各自请求）
const proxyOptions = computed(() =>
  props.proxies.map((px) => ({ value: px.ID, label: `${px.Scheme}://${px.Host}:${px.Port}` })),
)

const editName = ref('')
const editGroups = ref<number[]>([])
const editProxies = ref<number[]>([])
const editModels = ref<{ id: string }[]>([])
const editSaving = ref(false)
const editSyncing = ref(false)
const editGroupOptions = computed(() => {
  const pid = props.row?.plugin_id
  return props.groups.filter((g) => g.plugin_id === pid).map((g) => ({ value: g.id, label: `${g.name} (${g.plugin_label || g.plugin})` }))
})

// 打开时回填：名称/分组取列表行，代理与模型再拉一次详情
watch(() => props.visible, async (open) => {
  if (!open || !props.row) return
  const row = props.row
  editName.value = row.display_name
  editGroups.value = [...(row.group_ids ?? [])]
  editModels.value = []
  editProxies.value = []
  const [px, detail] = await Promise.all([
    api.get<{ proxy_ids: number[] }>(`/admin/accounts/${row.id}/proxies`).catch(() => ({ proxy_ids: [] })),
    api.get<AccountDetail>(`/admin/accounts/${row.id}/detail`).catch(() => null),
  ])
  editProxies.value = px.proxy_ids ?? []
  editModels.value = (detail?.models ?? []).map((m) => ({ id: m.id }))
})

// editSyncModels 拉上游模型目录（?refresh=1 落库）
async function editSyncModels() {
  if (!props.row || editSyncing.value) return
  editSyncing.value = true
  try {
    const resp = await api.get<{ models: ModelInfo[] | null }>(`/admin/accounts/${props.row.id}/models?refresh=1`)
    editModels.value = (resp.models ?? []).map((m) => ({ id: m.id }))
  } catch (e: any) {
    MessagePlugin.warning(t('accounts.syncFailed', { msg: e.message }))
  } finally {
    editSyncing.value = false
  }
}

// submitEdit 保存名称/分组/代理/模型（模型以用户勾选为准）
async function submitEdit() {
  if (!props.row) return
  editSaving.value = true
  try {
    const id = props.row.id
    await api.put(`/admin/accounts/${id}`, { display_name: editName.value, group_ids: editGroups.value })
    await api.put(`/admin/accounts/${id}/proxies`, { proxy_ids: editProxies.value })
    await api.put(`/admin/accounts/${id}/models`, { models: editModels.value })
    MessagePlugin.success(t('common.saved'))
    emit('update:visible', false)
    emit('saved')
  } catch (e: any) {
    MessagePlugin.error(e.message)
  } finally {
    editSaving.value = false
  }
}
</script>
