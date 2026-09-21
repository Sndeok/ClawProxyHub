<template>
    <!-- 在线测试：选端点/模型/问题 → 响应日志 -->
    <t-drawer :visible="visible" @update:visible="(v: boolean) => emit('update:visible', v)" :header="$t('accounts.testTitle')" size="560px" :footer="false">
      <t-space v-if="row" direction="vertical" style="width: 100%" size="large">
        <t-form label-width="80px">
          <t-form-item :label="$t('accounts.testEndpoint')">
            <bind-select v-model="testEndpoint" :multiple="false" :options="endpointOptions" />
          </t-form-item>
          <t-form-item :label="$t('accounts.testModel')">
            <bind-select v-model="testModel" :multiple="false" :options="modelOptions" :placeholder="$t('accounts.testModelPh')" />
          </t-form-item>
          <t-form-item :label="$t('accounts.testQuestion')">
            <t-input v-model="testQuestion" :placeholder="$t('accounts.testQuestionPh')" />
          </t-form-item>
        </t-form>
        <t-button theme="primary" block :loading="testing" :disabled="!testModel" @click="runTest">{{ $t('accounts.testRun') }}</t-button>
        <div v-if="testText" class="test-answer">{{ testText }}</div>
        <div v-if="testLogs.length" class="test-logs">
          <div v-for="(l, i) in testLogs" :key="i" class="test-log-line">{{ l }}</div>
        </div>
      </t-space>
    </t-drawer>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { api } from '../api/client'
import BindSelect from './BindSelect.vue'
import type { Account, AccountDetail } from '../api/types'

const props = defineProps<{ visible: boolean; row: Account | null }>()
const emit = defineEmits<{ 'update:visible': [boolean] }>()
const { t } = useI18n()

const endpointOptions = [
  { value: 'chat_completions', label: 'chat/completions' },
  { value: 'messages', label: 'messages' },
  { value: 'responses', label: 'responses' },
]
const testEndpoint = ref('chat_completions')
const testModel = ref('')
const testQuestion = ref('')
const testText = ref('')
const testLogs = ref<string[]>([])
const testing = ref(false)

// 模型候选来自该账号自己的模型目录（修掉旧实现复用编辑弹窗残留状态的问题）
const models = ref<{ id: string }[]>([])
const modelOptions = computed(() => models.value.map((m) => ({ value: m.id, label: m.id })))

watch(() => props.visible, async (open) => {
  if (!open || !props.row) return
  testEndpoint.value = 'chat_completions'
  testQuestion.value = ''
  testText.value = ''
  testLogs.value = []
  testModel.value = ''
  const detail = await api.get<AccountDetail>(`/admin/accounts/${props.row.id}/detail`).catch(() => null)
  models.value = (detail?.models ?? []).map((m) => ({ id: m.id }))
  if (models.value.length) testModel.value = models.value[0].id
})

// runTest 直调插件 Chat（绕路由/key），输出响应与日志
async function runTest() {
  if (!props.row || !testModel.value) return
  testing.value = true
  testText.value = ''
  testLogs.value = []
  try {
    const resp = await api.post<{ text: string; logs: string[] }>(`/admin/accounts/${props.row.id}/test`, {
      endpoint: testEndpoint.value, model: testModel.value, question: testQuestion.value,
    })
    testText.value = resp.text ?? ''
    testLogs.value = resp.logs ?? []
  } catch (e: any) {
    testLogs.value = ['✗ ' + (e.message || 'error')]
  } finally {
    testing.value = false
  }
}
</script>

<style scoped>
.test-answer {
  white-space: pre-wrap;
  background: var(--cph-surface-2);
  border: 1px solid var(--cph-border);
  border-radius: var(--cph-radius);
  padding: 12px;
  font-size: 13px;
}
.test-logs {
  max-height: 240px;
  overflow: auto;
  font-family: ui-monospace, Menlo, Consolas, monospace;
  font-size: 12px;
  color: var(--cph-text-2);
}
.test-log-line { padding: 2px 0; }
</style>
