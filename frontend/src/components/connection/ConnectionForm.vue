<script setup>
/**
 * ConnectionForm.vue
 * Navicat-style database connection editor with tabs: General / Remote / AI Assistant
 * Supports MySQL, PostgreSQL, Oracle, and SQL Server with database-specific fields.
 *
 * Structural note: All reactive state, computed properties, watchers, validation,
 * and action handlers live in useConnectionFormState.mjs.  The three tab sections
 * are rendered by ConnectionGeneralSection, ConnectionRemoteSection, and
 * ConnectionAISection respectively.
 */
import { useConnectionFormState } from './useConnectionFormState.mjs'
import ConnectionGeneralSection from './ConnectionGeneralSection.vue'
import ConnectionRemoteSection from './ConnectionRemoteSection.vue'
import ConnectionAISection from './ConnectionAISection.vue'

// Props
const props = defineProps({
  connectionId: {
    type: String,
    default: null
  },
  mode: {
    type: String,
    default: 'create',
    validator: (value) => ['create', 'edit'].includes(value)
  },
  defaultType: {
    type: String,
    default: ''
  }
})

// Emits
const emit = defineEmits(['saved', 'cancelled', 'tested'])

const s = useConnectionFormState(props, emit)
</script>

<template>
  <div class="conn-editor">
    <!-- Header -->
    <div class="conn-editor__header">
      <h2 class="conn-editor__title">{{ s.title }}</h2>
    </div>

    <!-- Type Selection (only for new connection) -->
    <div v-if="s.showTypeSelection && !s.isEditing" class="conn-editor__type-select">
      <div class="conn-editor__type-label">选择数据库类型</div>
      <div class="conn-editor__type-grid">
        <button
          v-for="opt in s.typeOptions"
          :key="opt.value"
          class="conn-editor__type-btn"
          :class="{ 'conn-editor__type-btn--active': s.selectedType === opt.value }"
          @click="s.selectedType = opt.value"
        >
          <span class="conn-editor__type-icon">{{ opt.icon }}</span>
          <span class="conn-editor__type-name">{{ opt.label }}</span>
        </button>
      </div>
    </div>

    <!-- Tabbed Editor (shown after type selection or in edit mode) -->
    <div v-else class="conn-editor__content">
      <!-- Error Banner -->
      <div v-if="s.formError" class="conn-editor__error">
        {{ s.formError }}
      </div>

      <!-- Tabs -->
      <div class="conn-editor__tabs">
        <button
          v-for="tab in s.tabs"
          :key="tab.id"
          class="conn-editor__tab"
          :class="{ 'conn-editor__tab--active': s.activeTab === tab.id }"
          @click="s.activeTab = tab.id"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Tab Content -->
      <div class="conn-editor__tab-content">
        <!-- ==================== General Tab ==================== -->
        <div v-show="s.activeTab === 'general'" class="conn-editor__panel">
          <ConnectionGeneralSection
            :form-data="s.formData"
            :field-errors="s.fieldErrors"
            :current-schema="s.currentSchema"
            :should-show-host-port="s.shouldShowHostPort"
            :should-show-database-field="s.shouldShowDatabaseField"
            :is-oracle-basic-mode="s.isOracleBasicMode"
            :is-oracle-t-n-s-mode="s.isOracleTNSMode"
            :show-password="s.showPassword"
            :db-test-result="s.dbTestResult"
            :validate-field="s.validateField"
            :get-oracle-mode-field-error="s.getOracleModeFieldError"
            @update:show-password="s.showPassword = $event"
          />
        </div>

        <!-- ==================== Remote Tab ==================== -->
        <div v-show="s.activeTab === 'remote'" class="conn-editor__panel">
          <ConnectionRemoteSection
            :form-data="s.formData"
            :field-errors="s.fieldErrors"
            :is-remote-type-none-selected="s.isRemoteTypeNoneSelected"
            :is-remote-type-s-s-h-selected="s.isRemoteTypeSSHSelected"
            :is-remote-type-win-r-m-selected="s.isRemoteTypeWinRMSelected"
            :show-ssh-password="s.showSshPassword"
            :show-winrm-password="s.showWinrmPassword"
            :ssh-test-result="s.sshTestResult"
            :winrm-test-result="s.winrmTestResult"
            @set-remote-type="s.setRemoteType"
            @on-ssh-host-change="s.onSshHostChange"
            @on-winrm-host-change="s.onWinRMHostChange"
            @on-winrm-port-change="s.onWinRMPortChange"
            @sync-remote-host-from-general="s.syncCurrentRemoteHostFromGeneral"
            @update:show-ssh-password="s.showSshPassword = $event"
            @update:show-winrm-password="s.showWinrmPassword = $event"
          />
        </div>

        <!-- ==================== AI Assistant Tab ==================== -->
        <div v-show="s.activeTab === 'ai'" class="conn-editor__panel conn-editor__panel--ai">
          <ConnectionAISection
            :form-data="s.formData"
            :selected-assistant="s.selectedAssistant"
            :ai-providers="s.aiProviders"
            :show-provider-dropdown="s.showProviderDropdown"
            :dropdown-position="s.dropdownPosition"
            :show-api-key="s.showApiKey"
            :ai-testing="s.aiTesting"
            :ai-test-result="s.aiTestResult"
            :ai-test-status="s.aiTestStatus"
            :show-ai-test-dialog="s.showAiTestDialog"
            :model-querying="s.modelQuerying"
            :model-query-error="s.modelQueryError"
            :available-models="s.availableModels"
            :show-model-selector="s.showModelSelector"
            :pending-model-selection="s.pendingModelSelection"
            :select-assistant="s.selectAssistant"
            :toggle-provider-dropdown="s.toggleProviderDropdown"
            :add-assistant="s.addAssistant"
            :remove-assistant="s.removeAssistant"
            :get-provider-info="s.getProviderInfo"
            :should-show-assistant-api-key-field="s.shouldShowAssistantApiKeyField"
            :toggle-api-key-visibility="s.toggleApiKeyVisibility"
            :handle-query-models="s.handleQueryModels"
            :handle-test-a-i="s.handleTestAI"
            :select-model="s.selectModel"
            :confirm-model-selection="s.confirmModelSelection"
            :close-model-selector="s.closeModelSelector"
            :open-ai-test-dialog="s.openAiTestDialog"
            :close-ai-test-dialog="s.closeAiTestDialog"
            :get-a-i-field-error="s.getAIFieldError"
            :close-provider-dropdown="s.closeProviderDropdown"
          />
        </div>
      </div>
    </div>

    <!-- Footer Actions -->
    <div class="conn-editor__footer">
      <div class="conn-editor__footer-left">
        <!-- General tab: Test DB button (hidden during type selection) -->
        <button
          v-if="s.activeTab === 'general' && !s.showTypeSelection"
          type="button"
          class="conn-editor__btn conn-editor__btn--test"
          :class="{
            'conn-editor__btn--testing': s.dbTesting,
            'conn-editor__btn--success': s.dbTestStatus === 'success',
            'conn-editor__btn--error': s.dbTestStatus === 'error'
          }"
          :disabled="s.dbTesting || s.saving"
          @click="s.handleTestDB"
        >
          <span v-if="s.dbTesting" class="spinner"></span>
          <span v-else-if="s.dbTestStatus === 'success'" class="icon-success">✓</span>
          <span v-else-if="s.dbTestStatus === 'error'" class="icon-error">✗</span>
          {{ s.dbTesting ? '测试中...' : 'Test DB' }}
        </button>

        <!-- Remote tab: Test SSH / Test WinRM button -->
        <button
          v-if="s.activeTab === 'remote' && s.isRemoteTypeSSHSelected"
          type="button"
          class="conn-editor__btn conn-editor__btn--test"
          :class="{
            'conn-editor__btn--testing': s.sshTesting,
            'conn-editor__btn--success': s.sshTestStatus === 'success',
            'conn-editor__btn--error': s.sshTestStatus === 'error'
          }"
          :disabled="s.sshTesting || s.saving"
          @click="s.handleTestSSH"
        >
          <span v-if="s.sshTesting" class="spinner"></span>
          <span v-else-if="s.sshTestStatus === 'success'" class="icon-success">✓</span>
          <span v-else-if="s.sshTestStatus === 'error'" class="icon-error">✗</span>
          {{ s.sshTesting ? '测试中...' : 'Test SSH' }}
        </button>
        <button
          v-if="s.activeTab === 'remote' && s.isRemoteTypeWinRMSelected"
          type="button"
          class="conn-editor__btn conn-editor__btn--test"
          :class="{
            'conn-editor__btn--testing': s.winrmTesting,
            'conn-editor__btn--success': s.winrmTestStatus === 'success',
            'conn-editor__btn--error': s.winrmTestStatus === 'error'
          }"
          :disabled="s.winrmTesting || s.saving"
          @click="s.handleTestWinRM"
        >
          <span v-if="s.winrmTesting" class="spinner"></span>
          <span v-else-if="s.winrmTestStatus === 'success'" class="icon-success">✓</span>
          <span v-else-if="s.winrmTestStatus === 'error'" class="icon-error">✗</span>
          {{ s.winrmTesting ? '测试中...' : 'Test WinRM' }}
        </button>
      </div>

      <div class="conn-editor__footer-right">
        <button type="button" class="conn-editor__btn conn-editor__btn--cancel" @click="s.handleCancel" :disabled="s.saving">
          Cancel
        </button>
        <button
          v-if="!s.showTypeSelection || s.isEditing"
          type="button"
          class="conn-editor__btn conn-editor__btn--save"
          @click="s.handleSave"
          :disabled="s.saving"
        >
          {{ s.saving ? 'Saving...' : 'Save' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style>
/* ============================================================
   Navicat-style Light Theme Connection Editor
   ============================================================ */

.conn-editor {
  --bg-primary: #ffffff;
  --bg-secondary: #f5f7fa;
  --bg-input: #ffffff;
  --border-color: #dcdfe6;
  --border-focus: #409eff;
  --text-primary: #303133;
  --text-secondary: #606266;
  --text-muted: #909399;
  --success: #67c23a;
  --success-bg: #f0f9eb;
  --error: #f56c6c;
  --error-bg: #fef0f0;
  --primary: #409eff;
  --primary-hover: #66b1ff;
  --primary-light: #ecf5ff;

  background-color: var(--bg-primary);
  border-radius: 4px;
  width: 100%;
  min-width: 600px;
  max-width: 720px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

/* Header */
.conn-editor__header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  background-color: var(--bg-primary);
}

.conn-editor__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

/* Type Selection */
.conn-editor__type-select {
  padding: 24px;
}

.conn-editor__type-label {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 16px;
}

.conn-editor__type-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.conn-editor__type-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px 16px;
  background-color: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.conn-editor__type-btn:hover {
  border-color: var(--primary);
  background-color: #ecf5ff;
}

.conn-editor__type-btn--active {
  border-color: var(--primary);
  background-color: #ecf5ff;
}

.conn-editor__type-icon {
  font-size: 28px;
  margin-bottom: 8px;
}

.conn-editor__type-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

/* Content */
.conn-editor__content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

/* Error Banner */
.conn-editor__error {
  padding: 12px 20px;
  background-color: var(--error-bg);
  color: var(--error);
  font-size: 13px;
  border-bottom: 1px solid var(--error);
}

/* Tabs */
.conn-editor__tabs {
  display: flex;
  border-bottom: 1px solid var(--border-color);
  background-color: var(--bg-secondary);
}

.conn-editor__tab {
  padding: 12px 24px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.conn-editor__tab:hover {
  color: var(--primary);
}

.conn-editor__tab--active {
  color: var(--primary);
  border-bottom-color: var(--primary);
  background-color: var(--bg-primary);
}

/* Tab Content */
.conn-editor__tab-content {
  flex: 1;
  overflow-y: auto;
  min-height: 300px;
}

.conn-editor__panel {
  padding: 20px;
}

.conn-editor__panel--ai {
  padding: 0;
}

/* Form */
.conn-form__row {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}

.conn-form__row--inline {
  gap: 12px;
}

.conn-form__row--inline .conn-form__field {
  flex: 1;
}

.conn-form__row--checkbox {
  margin-top: 8px;
}

.conn-form__row--test {
  margin-top: 20px;
}

.conn-form__label {
  width: 140px;
  min-width: 140px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  text-align: right;
  padding-right: 12px;
}

.conn-form__label .required {
  color: var(--error);
  margin-left: 2px;
}

.conn-form__field {
  flex: 1;
  display: flex;
  align-items: center;
}

.conn-form__field--port {
  width: 100px;
  flex: none;
}

.conn-form__radio-group {
  flex: 1;
  display: flex;
  gap: 16px;
}

.conn-form__radio {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-primary);
}

.conn-form__input,
.conn-form__select,
.conn-form__textarea {
  flex: 1;
  padding: 8px 12px;
  background-color: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 13px;
  color: var(--text-primary);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.conn-form__input:focus,
.conn-form__select:focus,
.conn-form__textarea:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.conn-form__input--error {
  border-color: var(--error);
}

.conn-form__input--error:focus {
  box-shadow: 0 0 0 2px rgba(245, 108, 108, 0.2);
}

.conn-form__select {
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23606266' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  padding-right: 30px;
}

.conn-form__textarea {
  resize: vertical;
  min-height: 60px;
}

.conn-form__password {
  flex: 1;
  display: flex;
  align-items: center;
  position: relative;
}

.conn-form__password .conn-form__input {
  padding-right: 60px;
}

.conn-form__password-toggle {
  position: absolute;
  right: 8px;
  background: none;
  border: none;
  font-size: 12px;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px 8px;
}

.conn-form__password-toggle:hover {
  color: var(--primary);
}

.conn-form__sync-btn {
  padding: 6px 12px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
}

.conn-form__sync-btn:hover {
  background-color: #e9ecf0;
}

.conn-form__model-btn {
  padding: 8px 12px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 14px;
  color: var(--text-secondary);
  cursor: pointer;
}

.conn-form__model-btn:hover {
  background-color: #e9ecf0;
}

.conn-form__slider {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
}

.conn-form__range {
  flex: 1;
  height: 4px;
  -webkit-appearance: none;
  appearance: none;
  background: var(--border-color);
  border-radius: 2px;
  cursor: pointer;
}

.conn-form__range::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 16px;
  height: 16px;
  background: var(--primary);
  border-radius: 50%;
  cursor: pointer;
}

.conn-form__slider-value {
  min-width: 36px;
  font-size: 13px;
  color: var(--text-primary);
  text-align: center;
}

.conn-form__checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-primary);
}

.conn-form__checkbox input {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: var(--primary);
}

.conn-form__hint {
  margin-left: 152px;
  margin-top: -12px;
  margin-bottom: 16px;
  font-size: 11px;
  color: var(--text-muted);
}

.conn-form__error-text {
  margin-left: 152px;
  margin-top: -12px;
  margin-bottom: 16px;
  font-size: 11px;
  color: var(--error);
}

.conn-form__section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 20px 0 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}

.conn-form__test-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 13px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.conn-form__test-btn:hover:not(:disabled) {
  background-color: #e9ecf0;
}

.conn-form__test-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.conn-form__test-btn--testing {
  color: var(--primary);
  border-color: var(--primary);
}

.conn-form__test-btn--success {
  color: var(--success);
  border-color: var(--success);
  background-color: var(--success-bg);
}

.conn-form__test-btn--error {
  color: var(--error);
  border-color: var(--error);
  background-color: var(--error-bg);
}

.conn-form__test-result {
  margin-left: 152px;
  margin-top: 8px;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 12px;
}

.conn-form__test-result--success {
  background-color: var(--success-bg);
  color: var(--success);
  border: 1px solid var(--success);
}

.conn-form__test-result--error {
  background-color: var(--error-bg);
  color: var(--error);
  border: 1px solid var(--error);
}

/* ============================================================
   AI Config - Desktop Tool Style
   ============================================================ */

.ai-config {
  display: flex;
  min-height: 420px;
  background-color: var(--bg-primary);
}

/* Left Sidebar - Assistant List */
.ai-config__sidebar {
  width: 140px;
  border-right: 1px solid var(--border-color);
  background-color: var(--bg-secondary);
  display: flex;
  flex-direction: column;
}

.ai-config__sidebar-header {
  padding: 8px 10px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--border-color);
}

.ai-config__list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

.ai-config__list-item {
  padding: 6px 12px;
  font-size: 12px;
  color: var(--text-primary);
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-left: 2px solid transparent;
  transition: none;
}

.ai-config__list-item:hover {
  background-color: #e8eaed;
}

.ai-config__list-item--active {
  background-color: #fff;
  border-left-color: var(--primary);
  color: var(--primary);
}

.ai-config__sidebar-actions {
  display: flex;
  border-top: 1px solid var(--border-color);
  padding: 4px;
  gap: 4px;
}

.ai-config__action-btn {
  flex: 1;
  height: 26px;
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 2px;
  color: var(--text-secondary);
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ai-config__action-btn:hover:not(.ai-config__action-btn--disabled) {
  background-color: #e8eaed;
  color: var(--text-primary);
}

.ai-config__action-btn--disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Provider Dropdown */
.ai-config__add-wrapper {
  position: relative;
  flex: 1;
}

.ai-config__provider-dropdown {
  position: fixed;
  width: 190px;
  max-height: 204px;
  overflow-y: auto;
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  z-index: 1000;
}

.ai-config__provider-option {
  padding: 8px 12px;
  font-size: 12px;
  color: var(--text-primary);
  cursor: pointer;
  white-space: nowrap;
  overflow: visible;
}

.ai-config__provider-option:hover {
  background-color: #ecf5ff;
  color: var(--primary);
}

/* Readonly Field */
.ai-config__readonly {
  flex: 1;
  height: 26px;
  padding: 0 8px;
  font-size: 12px;
  line-height: 26px;
  color: var(--text-secondary);
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ai-config__readonly--hint {
  font-style: italic;
  color: var(--text-muted);
}

/* Model Field with Query Button */
.ai-config__model-field {
  flex: 1;
  display: flex;
  gap: 4px;
}

.ai-config__model-input {
  flex: 1;
}

.ai-config__model-btn {
  width: 32px;
  height: 26px;
  padding: 0;
  border: 1px solid var(--border-color);
  border-radius: 2px;
  background-color: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.ai-config__model-btn:hover:not(:disabled) {
  background-color: var(--primary);
  border-color: var(--primary);
  color: #fff;
}

.ai-config__model-btn--loading,
.ai-config__model-btn:disabled {
  opacity: 0.6;
  cursor: wait;
}

/* Model Query Error */
.ai-config__model-error {
  width: 100%;
  margin-top: 4px;
  font-size: 11px;
  color: var(--error);
}

/* Model Selector Dialog */
.ai-config__model-selector-overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(0, 0, 0, 0.22);
  z-index: 1200;
}

.ai-config__model-selector-dialog {
  width: min(420px, calc(100vw - 32px));
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.2);
  overflow: hidden;
}

.ai-config__model-selector-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-color);
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
}

.ai-config__model-selector-close {
  width: 20px;
  height: 20px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
}

.ai-config__model-selector-close:hover {
  color: var(--text-primary);
}

.ai-config__model-selector-list {
  max-height: 240px;
  overflow-y: auto;
  padding: 8px 0;
}

.ai-config__model-option {
  display: block;
  width: 100%;
  padding: 8px 12px;
  font-size: 12px;
  color: var(--text-secondary);
  text-align: left;
  border: none;
  background: transparent;
  cursor: pointer;
  transition: background-color 0.15s;
}

.ai-config__model-option:hover {
  background-color: var(--bg-secondary);
}

.ai-config__model-option--selected {
  background-color: var(--primary-light);
  color: var(--primary);
}

.ai-config__model-selector-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 16px 16px;
  border-top: 1px solid var(--border-color);
}

.ai-config__selector-btn {
  min-width: 72px;
  padding: 8px 14px;
  border-radius: 4px;
  border: 1px solid var(--border-color);
  font-size: 12px;
  cursor: pointer;
}

.ai-config__selector-btn--secondary {
  background-color: var(--bg-primary);
  color: var(--text-secondary);
}

.ai-config__selector-btn--primary {
  background-color: var(--primary);
  border-color: var(--primary);
  color: #fff;
}

.ai-config__selector-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Right Main Config Area */
.ai-config__main {
  flex: 1;
  padding: 12px 16px;
  overflow-y: auto;
}

/* Section */
.ai-config__section {
  margin-bottom: 12px;
}

.ai-config__section--bordered {
  padding-top: 10px;
  border-top: 1px solid var(--border-color);
}

.ai-config__section-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Form Row */
.ai-config__row {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  min-height: 26px;
}

.ai-config__row--temp {
  align-items: flex-start;
  flex-direction: column;
}

.ai-config__row--checkbox {
  padding-left: 100px;
}

/* Label */
.ai-config__label {
  width: 100px;
  min-width: 100px;
  font-size: 12px;
  color: var(--text-secondary);
  flex-shrink: 0;
  line-height: 26px;
}

.ai-config__help {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  margin-left: 4px;
  font-size: 10px;
  font-weight: 600;
  color: var(--text-muted);
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 50%;
  cursor: help;
}

/* Input */
.ai-config__input {
  flex: 1;
  height: 26px;
  padding: 0 8px;
  font-size: 12px;
  border: 1px solid var(--border-color);
  border-radius: 2px;
  background-color: var(--bg-input);
  color: var(--text-primary);
}

.ai-config__input:focus {
  outline: none;
  border-color: var(--primary);
}

.ai-config__input--error {
  border-color: var(--error);
}

.ai-config__input--error:focus {
  box-shadow: 0 0 0 2px rgba(245, 108, 108, 0.2);
}

.ai-config__input::placeholder {
  color: var(--text-muted);
}

/* Field Error Text */
.ai-config__error-text {
  width: 100%;
  margin-top: 2px;
  margin-bottom: 4px;
  font-size: 11px;
  color: var(--error);
  padding-left: 100px;
}

/* Select */
.ai-config__select {
  flex: 1;
  height: 26px;
  padding: 0 6px;
  font-size: 12px;
  border: 1px solid var(--border-color);
  border-radius: 2px;
  background-color: var(--bg-input);
  color: var(--text-primary);
  cursor: pointer;
}

.ai-config__select:focus {
  outline: none;
  border-color: var(--primary);
}

/* Textarea */
.ai-config__textarea {
  flex: 1;
  padding: 4px 8px;
  font-size: 12px;
  border: 1px solid var(--border-color);
  border-radius: 2px;
  background-color: var(--bg-input);
  color: var(--text-primary);
  resize: none;
  font-family: inherit;
  line-height: 1.4;
}

.ai-config__textarea:focus {
  outline: none;
  border-color: var(--primary);
}

/* API Key Field */
.ai-config__key-field {
  flex: 1;
  display: flex;
  align-items: center;
}

.ai-config__key-input {
  border-radius: 2px 0 0 2px;
  flex: 1;
}

.ai-config__key-toggle {
  width: 32px;
  height: 26px;
  padding: 0;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-left: none;
  border-radius: 0 2px 2px 0;
  color: var(--text-muted);
  font-size: 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ai-config__key-toggle:hover {
  background-color: #e8eaed;
  color: var(--text-primary);
}

/* Temperature Control */
.ai-config__temp-control {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ai-config__temp-labels {
  display: flex;
  justify-content: space-between;
  font-size: 10px;
  color: var(--text-muted);
  padding: 0 2px;
}

.ai-config__temp-slider {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ai-config__range {
  flex: 1;
  height: 4px;
  background: var(--border-color);
  border-radius: 2px;
  appearance: none;
  cursor: pointer;
}

.ai-config__range::-webkit-slider-thumb {
  appearance: none;
  width: 14px;
  height: 14px;
  background: var(--primary);
  border-radius: 50%;
  cursor: pointer;
}

.ai-config__range::-webkit-slider-thumb:hover {
  background: var(--primary-hover);
}

.ai-config__temp-value {
  min-width: 28px;
  font-size: 11px;
  color: var(--text-secondary);
  text-align: right;
  font-family: monospace;
}

/* Checkbox */
.ai-config__checkbox {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-primary);
  cursor: pointer;
}

.ai-config__checkbox input[type="checkbox"] {
  width: 14px;
  height: 14px;
  margin: 0;
  cursor: pointer;
}

/* Test Button Area */
.ai-config__test-area {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.ai-config__test-btn {
  padding: 6px 16px;
  font-size: 12px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 2px;
  color: var(--text-primary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: none;
}

.ai-config__test-btn:hover:not(:disabled) {
  background-color: #e8eaed;
}

.ai-config__test-btn--primary {
  background-color: var(--primary);
  border-color: var(--primary);
  color: #fff;
}

.ai-config__test-btn--primary:hover:not(:disabled) {
  background-color: var(--primary-hover);
}

.ai-config__test-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.ai-config__test-btn--testing {
  color: var(--primary);
}

.ai-config__test-btn--success {
  color: var(--success);
  border-color: var(--success);
}

.ai-config__test-btn--error {
  color: var(--error);
  border-color: var(--error);
}

.ai-config__test-result {
  font-size: 11px;
}

.ai-config__test-result--success {
  color: var(--success);
}

.ai-config__test-result--error {
  color: var(--error);
}

/* Footer */
.conn-editor__footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
  background-color: var(--bg-primary);
}

.conn-editor__footer-left,
.conn-editor__footer-right {
  display: flex;
  gap: 12px;
}

.conn-editor__btn {
  padding: 8px 20px;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.conn-editor__btn--test {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
}

.conn-editor__btn--test:hover:not(:disabled) {
  background-color: #e9ecf0;
}

.conn-editor__btn--test:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.conn-editor__btn--testing {
  color: var(--primary);
  border-color: var(--primary);
}

.conn-editor__btn--success {
  color: var(--success);
  border-color: var(--success);
  background-color: var(--success-bg);
}

.conn-editor__btn--error {
  color: var(--error);
  border-color: var(--error);
  background-color: var(--error-bg);
}

.conn-editor__btn--cancel {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
}

.conn-editor__btn--cancel:hover {
  background-color: #e9ecf0;
}

.conn-editor__btn--save {
  background-color: var(--primary);
  border: 1px solid var(--primary);
  color: white;
}

.conn-editor__btn--save:hover {
  background-color: var(--primary-hover);
}

.conn-editor__btn--save:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Spinner */
.spinner {
  width: 14px;
  height: 14px;
  border: 2px solid transparent;
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.icon-success {
  font-weight: 600;
}

.icon-error {
  font-weight: 600;
}
</style>
