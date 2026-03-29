<script setup>
/**
 * ConnectionAISection.vue
 * AI Assistant tab content for AI provider/model configuration.
 * Receives all state via props from parent ConnectionForm.
 */
import AiTestDialog from './AiTestDialog.vue'

const props = defineProps({
  formData: { type: Object, required: true },
  selectedAssistant: { type: Object, default: null },
  aiProviders: { type: Array, required: true },
  showProviderDropdown: { type: Boolean, required: true },
  dropdownPosition: { type: Object, required: true },
  showApiKey: { type: Object, required: true },
  aiTesting: { type: Boolean, required: true },
  aiTestResult: { type: Object, default: null },
  aiTestStatus: { type: String, required: true },
  showAiTestDialog: { type: Boolean, required: true },
  modelQuerying: { type: Boolean, required: true },
  modelQueryError: { type: String, required: true },
  availableModels: { type: Array, required: true },
  showModelSelector: { type: Boolean, required: true },
  pendingModelSelection: { type: String, required: true },

  // Functions
  selectAssistant: { type: Function, required: true },
  toggleProviderDropdown: { type: Function, required: true },
  addAssistant: { type: Function, required: true },
  removeAssistant: { type: Function, required: true },
  getProviderInfo: { type: Function, required: true },
  shouldShowAssistantApiKeyField: { type: Function, required: true },
  toggleApiKeyVisibility: { type: Function, required: true },
  handleQueryModels: { type: Function, required: true },
  handleTestAI: { type: Function, required: true },
  selectModel: { type: Function, required: true },
  confirmModelSelection: { type: Function, required: true },
  closeModelSelector: { type: Function, required: true },
  openAiTestDialog: { type: Function, required: true },
  closeAiTestDialog: { type: Function, required: true },
  getAIFieldError: { type: Function, required: true },
  closeProviderDropdown: { type: Function, required: true }
})

const handleMainAreaClick = () => {
  if (props.showProviderDropdown) {
    props.closeProviderDropdown()
  }
}
</script>

<template>
  <div class="ai-config">
    <!-- Left: Assistant List -->
    <div class="ai-config__sidebar">
      <div class="ai-config__sidebar-header">AI 助手</div>
      <div class="ai-config__list">
        <div
          v-for="assistant in formData.ai_assistants"
          :key="assistant.id"
          class="ai-config__list-item"
          :class="{ 'ai-config__list-item--active': formData.selectedAssistantId === assistant.id }"
          @click="selectAssistant(assistant.id)"
          :title="assistant.name"
        >
          {{ assistant.name }}
        </div>
      </div>
      <div class="ai-config__sidebar-actions">
        <button class="ai-config__action-btn" @click="toggleProviderDropdown($event)" title="添加助手">+</button>
        <button
          class="ai-config__action-btn"
          :class="{ 'ai-config__action-btn--disabled': !selectedAssistant }"
          @click="removeAssistant(selectedAssistant?.id)"
          :disabled="!selectedAssistant"
          title="删除助手"
        >−</button>
      </div>
    </div>

    <!-- Provider Dropdown - Fixed position using Teleport -->
    <Teleport to="body">
      <div
        v-if="showProviderDropdown"
        class="ai-config__provider-dropdown"
        :style="{ top: dropdownPosition.top + 'px', left: dropdownPosition.left + 'px' }"
      >
        <div
          v-for="p in aiProviders"
          :key="p.value"
          class="ai-config__provider-option"
          @click="addAssistant(p.value)"
        >
          {{ p.label }}
        </div>
      </div>
    </Teleport>

    <!-- Right: Config Form -->
    <div class="ai-config__main" @click="handleMainAreaClick">
      <!-- Empty state when no assistants -->
      <div v-if="formData.ai_assistants.length === 0" class="ai-config__empty">
        <div class="ai-config__empty-icon">🤖</div>
        <div class="ai-config__empty-title">暂无 AI 助手</div>
        <div class="ai-config__empty-hint">请点击左下角 + 添加 AI 助手</div>
      </div>

      <template v-else-if="selectedAssistant">
        <!-- Basic Config Section -->
        <div class="ai-config__section">
          <div class="ai-config__row">
            <label class="ai-config__label">AI 助手名称</label>
            <input v-model="selectedAssistant.name" type="text" class="ai-config__input" />
          </div>

          <div class="ai-config__row">
            <label class="ai-config__label">AI 提供商</label>
            <div class="ai-config__readonly">{{ getProviderInfo(selectedAssistant.provider).label }}</div>
          </div>

          <div class="ai-config__row">
            <label class="ai-config__label">API 主机</label>
            <input
              v-model="selectedAssistant.api_host"
              type="text"
              class="ai-config__input"
              :class="{ 'ai-config__input--error': getAIFieldError(selectedAssistant.id, 'api_host') }"
              placeholder="https://api.example.com"
            />
          </div>
          <div v-if="getAIFieldError(selectedAssistant.id, 'api_host')" class="ai-config__error-text">
            {{ getAIFieldError(selectedAssistant.id, 'api_host') }}
          </div>

          <!-- API Endpoint is readonly and linked to provider -->
          <div class="ai-config__row">
            <label class="ai-config__label">API 端点</label>
            <div class="ai-config__readonly">{{ getProviderInfo(selectedAssistant.provider).endpoint }}</div>
          </div>

          <div v-if="shouldShowAssistantApiKeyField(selectedAssistant)" class="ai-config__row">
            <label class="ai-config__label">
              API 密钥
              <span class="ai-config__help" title="API 密钥用于身份验证，请妥善保管">?</span>
            </label>
            <div class="ai-config__key-field">
              <input
                v-model="selectedAssistant.api_key"
                :type="showApiKey[selectedAssistant.id] ? 'text' : 'password'"
                class="ai-config__input ai-config__key-input"
                :class="{ 'ai-config__input--error': getAIFieldError(selectedAssistant.id, 'api_key') }"
                placeholder="sk-..."
              />
              <button
                type="button"
                class="ai-config__key-toggle"
                @click="toggleApiKeyVisibility(selectedAssistant.id)"
                :title="showApiKey[selectedAssistant.id] ? '隐藏密钥' : '显示密钥'"
              >
                <span v-if="showApiKey[selectedAssistant.id]">🔒</span>
                <span v-else>👁</span>
              </button>
            </div>
          </div>
          <div v-if="shouldShowAssistantApiKeyField(selectedAssistant) && getAIFieldError(selectedAssistant.id, 'api_key')" class="ai-config__error-text">
            {{ getAIFieldError(selectedAssistant.id, 'api_key') }}
          </div>

          <div class="ai-config__row">
            <label class="ai-config__label">模型</label>
            <div class="ai-config__model-field">
              <input
                v-model="selectedAssistant.model"
                type="text"
                class="ai-config__input ai-config__model-input"
                :class="{ 'ai-config__input--error': getAIFieldError(selectedAssistant.id, 'model') }"
                placeholder="模型名称"
              />
              <button
                type="button"
                class="ai-config__model-btn"
                :class="{ 'ai-config__model-btn--loading': modelQuerying }"
                :disabled="modelQuerying"
                @click="handleQueryModels"
                title="查询可用模型"
              >{{ modelQuerying ? '查询中' : '查询' }}</button>
            </div>
          </div>
          <div v-if="getAIFieldError(selectedAssistant.id, 'model')" class="ai-config__error-text">
            {{ getAIFieldError(selectedAssistant.id, 'model') }}
          </div>
          <!-- Model query error -->
          <div v-if="modelQueryError" class="ai-config__error-text">{{ modelQueryError }}</div>

          <div class="ai-config__row ai-config__row--temp">
            <label class="ai-config__label">温度</label>
            <div class="ai-config__temp-control">
              <div class="ai-config__temp-labels">
                <span>0.0 更有确定性</span>
                <span>1.0 平衡</span>
                <span>2.0 更有创造性</span>
              </div>
              <div class="ai-config__temp-slider">
                <input
                  v-model.number="selectedAssistant.temperature"
                  type="range"
                  min="0"
                  max="2"
                  step="0.1"
                  class="ai-config__range"
                />
                <span class="ai-config__temp-value">{{ selectedAssistant.temperature.toFixed(1) }}</span>
              </div>
            </div>
          </div>

          <div class="ai-config__row">
            <label class="ai-config__label">说明</label>
            <textarea
              v-model="selectedAssistant.description"
              class="ai-config__textarea"
              rows="2"
              placeholder="（可选）"
            ></textarea>
          </div>
        </div>

        <!-- Test Button -->
        <div class="ai-config__test-area">
          <button
            type="button"
            class="ai-config__test-btn"
            :class="{
              'ai-config__test-btn--testing': aiTestStatus === 'testing',
              'ai-config__test-btn--success': aiTestStatus === 'success',
              'ai-config__test-btn--error': aiTestStatus === 'error'
            }"
            :disabled="aiTesting"
            @click="handleTestAI"
          >
            <span v-if="aiTestStatus === 'testing'" class="spinner"></span>
            <span v-else-if="aiTestStatus === 'success'">✓</span>
            <span v-else-if="aiTestStatus === 'error'">✗</span>
            测试连接
          </button>
          <button
            type="button"
            class="ai-config__test-btn ai-config__test-btn--primary"
            @click="openAiTestDialog"
          >
            测试模型
          </button>
          <span v-if="aiTestResult" class="ai-config__test-result" :class="aiTestResult.success ? 'ai-config__test-result--success' : 'ai-config__test-result--error'">
            {{ aiTestResult.success ? (aiTestResult.message || '连接成功') : aiTestResult.error }}
          </span>
        </div>
      </template>
    </div>
  </div>

  <Teleport to="body">
    <div v-if="showModelSelector" class="ai-config__model-selector-overlay" @click.self="closeModelSelector">
      <div class="ai-config__model-selector-dialog" role="dialog" aria-modal="true" aria-label="选择模型">
        <div class="ai-config__model-selector-header">
          <span>选择模型</span>
          <button type="button" class="ai-config__model-selector-close" @click="closeModelSelector">×</button>
        </div>
        <div class="ai-config__model-selector-list">
          <button
            v-for="model in availableModels"
            :key="model.id"
            type="button"
            class="ai-config__model-option"
            :class="{ 'ai-config__model-option--selected': pendingModelSelection === model.id }"
            @click="selectModel(model.id)"
          >
            {{ model.name }}
          </button>
        </div>
        <div class="ai-config__model-selector-actions">
          <button type="button" class="ai-config__selector-btn ai-config__selector-btn--secondary" @click="closeModelSelector">
            取消
          </button>
          <button
            type="button"
            class="ai-config__selector-btn ai-config__selector-btn--primary"
            :disabled="!pendingModelSelection"
            @click="confirmModelSelection"
          >
            确定
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <AiTestDialog
    :visible="showAiTestDialog"
    :assistant="selectedAssistant"
    :provider-label="selectedAssistant ? getProviderInfo(selectedAssistant.provider).label : ''"
    @close="closeAiTestDialog"
  />
</template>
