<script setup>
import { computed, ref, watch } from 'vue'
import { TestAIChat } from '../../../wailsjs/go/bindings/ConnectionBinding'
import {
  buildAiChatTestRequest,
  createAiTestDialogState
} from './connectionFormAiTestState.mjs'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  assistant: {
    type: Object,
    default: null
  },
  providerLabel: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['close'])

const prompt = ref('')
const status = ref('idle')
const responseText = ref('')
const errorText = ref('')
const latencyMs = ref(null)

const resetDialogState = () => {
  const state = createAiTestDialogState()
  prompt.value = state.prompt
  status.value = state.status
  responseText.value = state.responseText
  errorText.value = state.errorText
  latencyMs.value = state.latencyMs
}

watch(() => props.visible, (visible) => {
  if (visible) {
    resetDialogState()
  }
}, { immediate: true })

const isSending = computed(() => status.value === 'sending')
const assistantName = computed(() => props.assistant?.name || '')

const handleClose = () => {
  emit('close')
}

const handleSend = async () => {
  if (!props.assistant) {
    return
  }

  status.value = 'sending'
  responseText.value = ''
  errorText.value = ''
  latencyMs.value = null

  try {
    const result = await TestAIChat(buildAiChatTestRequest(props.assistant, prompt.value))
    latencyMs.value = result.latency_ms ?? null

    if (result.success) {
      status.value = 'success'
      responseText.value = result.content || ''
      return
    }

    status.value = 'error'
    errorText.value = result.error || '模型测试失败'
  } catch (error) {
    status.value = 'error'
    errorText.value = error.message || '模型测试失败'
  }
}
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="ai-test-dialog__overlay">
      <div class="ai-test-dialog" role="dialog" aria-modal="true" aria-label="模型测试">
        <div class="ai-test-dialog__header">
          <div>
            <div class="ai-test-dialog__title">测试 AI 助手</div>
            <div class="ai-test-dialog__meta">{{ assistantName }} · {{ providerLabel }}</div>
          </div>
          <button type="button" class="ai-test-dialog__close" @click="handleClose">×</button>
        </div>

        <div class="ai-test-dialog__body">
          <label class="ai-test-dialog__label">测试问题</label>
          <textarea
            v-model="prompt"
            class="ai-test-dialog__textarea"
            rows="4"
            placeholder="请输入要发送给模型的问题"
          />

          <div class="ai-test-dialog__actions">
            <button type="button" class="ai-test-dialog__btn ai-test-dialog__btn--secondary" @click="handleClose">
              关闭
            </button>
            <button
              type="button"
              class="ai-test-dialog__btn ai-test-dialog__btn--primary"
              :disabled="isSending"
              @click="handleSend"
            >
              {{ isSending ? '发送中...' : '发送' }}
            </button>
          </div>

          <div class="ai-test-dialog__result">
            <div class="ai-test-dialog__result-header">
              <span>模型响应</span>
              <span v-if="latencyMs !== null" class="ai-test-dialog__latency">{{ latencyMs }}ms</span>
            </div>
            <div v-if="status === 'idle'" class="ai-test-dialog__placeholder">
              发送后将在这里显示模型返回内容。
            </div>
            <div v-else-if="status === 'sending'" class="ai-test-dialog__placeholder">
              正在等待模型响应...
            </div>
            <div v-else-if="status === 'error'" class="ai-test-dialog__error">
              {{ errorText }}
            </div>
            <pre v-else class="ai-test-dialog__response">{{ responseText }}</pre>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.ai-test-dialog__overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(0, 0, 0, 0.28);
  z-index: 1300;
}

.ai-test-dialog {
  width: min(760px, calc(100vw - 32px));
  max-height: min(720px, calc(100vh - 32px));
  display: flex;
  flex-direction: column;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 10px;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.2);
  overflow: hidden;
}

.ai-test-dialog__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 16px 18px 12px;
  border-bottom: 1px solid #ebeef5;
}

.ai-test-dialog__title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

.ai-test-dialog__meta {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}

.ai-test-dialog__close {
  border: none;
  background: transparent;
  color: #909399;
  font-size: 18px;
  cursor: pointer;
}

.ai-test-dialog__body {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px 18px 18px;
  overflow: auto;
}

.ai-test-dialog__label {
  font-size: 12px;
  font-weight: 600;
  color: #606266;
}

.ai-test-dialog__textarea {
  width: 100%;
  min-height: 92px;
  padding: 10px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  font-size: 13px;
  resize: vertical;
}

.ai-test-dialog__textarea:focus {
  outline: none;
  border-color: #409eff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.15);
}

.ai-test-dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.ai-test-dialog__btn {
  min-width: 84px;
  padding: 8px 14px;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
}

.ai-test-dialog__btn--secondary {
  background: #fff;
  color: #606266;
}

.ai-test-dialog__btn--primary {
  background: #409eff;
  border-color: #409eff;
  color: #fff;
}

.ai-test-dialog__btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.ai-test-dialog__result {
  min-height: 280px;
  display: flex;
  flex-direction: column;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  background: #f8fafc;
  overflow: hidden;
}

.ai-test-dialog__result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  border-bottom: 1px solid #ebeef5;
  background: #f5f7fa;
  font-size: 12px;
  font-weight: 600;
  color: #606266;
}

.ai-test-dialog__latency {
  color: #909399;
  font-weight: 500;
}

.ai-test-dialog__placeholder,
.ai-test-dialog__error,
.ai-test-dialog__response {
  flex: 1;
  margin: 0;
  padding: 14px 12px;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  overflow: auto;
}

.ai-test-dialog__placeholder {
  color: #909399;
}

.ai-test-dialog__error {
  color: #f56c6c;
}

.ai-test-dialog__response {
  color: #303133;
}
</style>
