<template>
  <section class="detail-panel">
    <div v-if="templateStore.deleteCandidate" class="confirm-overlay" @click.self="templateStore.cancelDeleteTemplate()">
      <div class="confirm-modal">
        <div class="confirm-title">Delete Template</div>
        <p class="confirm-body">
          Delete editable template "{{ templateStore.deleteCandidate.name }}" from local mock state? This cannot be undone in the current session.
        </p>
        <div class="confirm-actions">
          <button class="btn btn-ghost" @click="templateStore.cancelDeleteTemplate()">Cancel</button>
          <button class="btn btn-danger" @click="templateStore.confirmDeleteTemplate()">Delete</button>
        </div>
      </div>
    </div>

    <div v-if="!activeTemplate" class="empty-wrap">
      <TemplateEmptyState
        title="No template selected"
        description="Choose a template from the left list to inspect it, or create a new user template draft."
        primary-label="Create Template"
        @primary="templateStore.createTemplate()"
      />
    </div>

    <template v-else>
      <div class="detail-header">
        <div class="header-main">
          <div class="mode-line">
            <span class="mode-badge">{{ editorLabel }}</span>
            <span class="mode-badge muted">{{ templateStore.scopeLabels[activeTemplate.scope] }}</span>
            <span class="mode-badge muted">{{ templateStore.statusLabels[activeTemplate.status] }}</span>
            <span v-if="templateStore.isDirty" class="mode-badge dirty">Unsaved changes</span>
          </div>
          <h2 class="detail-title">{{ activeTemplate.name }}</h2>
          <p class="detail-subtitle">{{ activeTemplate.description }}</p>
        </div>

        <div class="header-actions">
          <div class="mode-switch">
            <button
              v-for="mode in editorModes"
              :key="mode.value"
              class="mode-btn"
              :class="{ active: templateStore.editorMode === mode.value }"
              @click="templateStore.editorMode = mode.value"
            >
              {{ mode.label }}
            </button>
          </div>

          <div class="action-row">
            <button
              v-if="templateStore.editorState === 'view' && canDirectEdit"
              class="btn btn-secondary"
              @click="templateStore.startEditing()"
            >
              Edit Template
            </button>
            <button
              v-else-if="templateStore.editorState === 'view'"
              class="btn btn-secondary"
              @click="handleSaveAs"
            >
              Save As User Template
            </button>
            <button
              v-if="canDeleteSelected && templateStore.editorState === 'view'"
              class="btn btn-danger"
              @click="templateStore.requestDeleteTemplate()"
            >
              Delete
            </button>
          </div>
        </div>
      </div>

      <div class="detail-scroll">
        <div v-if="isReadonlyScope && templateStore.editorState === 'view'" class="readonly-banner">
          <div>
            <strong>{{ readonlyBannerTitle }}</strong>
            <span> {{ readonlyBannerBody }}</span>
          </div>
          <button class="btn btn-secondary" @click="handleSaveAsReadonly">Save As User Template</button>
        </div>

        <div v-if="templateStore.hasValidationErrors" class="validation-summary">
          <div class="validation-title">Validation issues</div>
          <ul class="validation-list">
            <li v-for="(message, key) in templateStore.validationErrors" :key="key">{{ message }}</li>
          </ul>
        </div>

        <TemplateBasicSection :template-model="activeTemplate" :readonly="isReadOnly" />
        <TemplateRuntimeSection :template-model="activeTemplate" :readonly="isReadOnly" />
        <TemplatePreviewSection :template-model="activeTemplate" />
      </div>

      <div class="footer-actions">
        <div class="left-actions">
          <button
            v-if="templateStore.editorState === 'editing' || templateStore.editorState === 'creating'"
            class="btn btn-ghost"
            @click="templateStore.cancelEditing()"
          >
            Cancel
          </button>
        </div>

        <div class="right-actions">
          <button
            class="btn btn-primary"
            @click="handleSave"
          >
            Save
          </button>
          <button class="btn btn-secondary" @click="handleSaveAs">Save As</button>
          <button class="btn btn-secondary" @click="handleCreateTask">Create Task from Template</button>
        </div>
      </div>
    </template>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import TemplateBasicSection from './TemplateBasicSection.vue'
import TemplateEmptyState from './TemplateEmptyState.vue'
import TemplatePreviewSection from './TemplatePreviewSection.vue'
import TemplateRuntimeSection from './TemplateRuntimeSection.vue'
import { useAppStore } from '../../stores/app'
import { useTemplateStore } from '../../stores/template'

const templateStore = useTemplateStore()
const appStore = useAppStore()

const editorModes = [
  { value: 'standard', label: 'Standard' },
  { value: 'advanced', label: 'Advanced' },
  { value: 'expert', label: 'Expert' }
]

const activeTemplate = computed(() => templateStore.activeTemplate)
const isReadOnly = computed(() => templateStore.editorState === 'view')
const canDirectEdit = computed(() => ['user', 'project', 'test'].includes(templateStore.selectedTemplate?.scope))
const canDeleteSelected = computed(() => ['user', 'test'].includes(templateStore.selectedTemplate?.scope))
const isReadonlyScope = computed(() => ['builtin', 'readonlyShared'].includes(templateStore.selectedTemplate?.scope))
const editorLabel = computed(() => {
  if (templateStore.editorState === 'creating') return 'New Template'
  if (templateStore.editorState === 'editing') return 'Editing'
  return 'Overview'
})
const readonlyBannerTitle = computed(() => {
  if (templateStore.selectedTemplate?.scope === 'readonlyShared') {
    return 'Readonly Shared templates are read-only.'
  }
  return 'Built-in templates are read-only.'
})
const readonlyBannerBody = computed(() => {
  if (templateStore.selectedTemplate?.scope === 'readonlyShared') {
    return 'Use Save As to create your own editable copy, or create a task directly from the shared template.'
  }
  return 'Use Save As to create a user-editable copy before making changes.'
})

const handleSave = () => {
  if (isReadOnly.value) {
    if (['user', 'project', 'test'].includes(templateStore.selectedTemplate?.scope)) {
      templateStore.showNotice('Switch to edit mode before saving updates.', 'info')
    } else {
      templateStore.placeholderAction('unsupportedEdit')
    }
    return
  }

  templateStore.saveTemplate()
}

const handleSaveAsReadonly = () => {
  templateStore.placeholderAction('readonlySaveAs')
  templateStore.saveAsTemplate()
}

const handleSaveAs = () => {
  if (['builtin', 'readonlyShared'].includes(templateStore.selectedTemplate?.scope) && templateStore.editorState === 'view') {
    handleSaveAsReadonly()
    return
  }

  templateStore.saveAsTemplate()
}

const handleCreateTask = () => {
  const payload = templateStore.createTaskFromTemplate()
  if (!payload) return

  appStore.queueTemplateForTask(payload)
  templateStore.showNotice(`Task shell opened in Tasks & Monitor for template "${payload.templateName}".`, 'success')
}
</script>

<style scoped>
.detail-panel {
  position: relative;
  min-height: 0;
  display: flex;
  flex-direction: column;
  border: 1px solid #2d3748;
  border-radius: 14px;
  background: linear-gradient(180deg, #111827 0%, #0f172a 100%);
  overflow: hidden;
}

.empty-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.confirm-overlay {
  position: absolute;
  inset: 0;
  background: rgba(2, 6, 23, 0.72);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  padding: 24px;
}

.confirm-modal {
  width: min(460px, 100%);
  border-radius: 14px;
  border: 1px solid #334155;
  background: #111827;
  box-shadow: 0 18px 48px rgba(0, 0, 0, 0.4);
  padding: 20px;
}

.confirm-title {
  font-size: 18px;
  font-weight: 700;
  color: #f8fafc;
}

.confirm-body {
  margin-top: 10px;
  color: #cbd5e1;
  line-height: 1.6;
  font-size: 13px;
}

.confirm-actions {
  margin-top: 18px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  padding: 18px 20px;
  border-bottom: 1px solid #1f2937;
  flex-wrap: wrap;
}

.mode-line {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 10px;
}

.mode-badge {
  padding: 4px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
  background: rgba(59, 130, 246, 0.15);
  color: #93c5fd;
}

.mode-badge.muted {
  background: rgba(148, 163, 184, 0.16);
  color: #cbd5e1;
}

.mode-badge.dirty {
  background: rgba(245, 158, 11, 0.16);
  color: #fcd34d;
}

.detail-title {
  font-size: 22px;
  color: #f8fafc;
}

.detail-subtitle {
  margin-top: 6px;
  color: #94a3b8;
  max-width: 740px;
  font-size: 13px;
}

.header-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-end;
}

.mode-switch {
  display: flex;
  gap: 4px;
  background: #0f172a;
  padding: 4px;
  border-radius: 10px;
}

.mode-btn {
  border: none;
  border-radius: 8px;
  padding: 8px 12px;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
}

.mode-btn.active {
  background: #1e293b;
  color: #f8fafc;
}

.action-row {
  display: flex;
  gap: 8px;
}

.detail-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.readonly-banner,
.validation-summary {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
  padding: 14px 16px;
  border-radius: 10px;
  border: 1px solid #334155;
  background: rgba(15, 23, 42, 0.86);
}

.readonly-banner {
  border-color: rgba(96, 165, 250, 0.3);
  background: rgba(30, 64, 175, 0.12);
  color: #dbeafe;
}

.validation-summary {
  flex-direction: column;
  border-color: rgba(239, 68, 68, 0.35);
  background: rgba(127, 29, 29, 0.16);
}

.validation-title {
  font-size: 13px;
  font-weight: 700;
  color: #fecaca;
}

.validation-list {
  margin: 0;
  padding-left: 18px;
  color: #fecaca;
  font-size: 12px;
  line-height: 1.7;
}

.footer-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #1f2937;
  background: rgba(15, 23, 42, 0.9);
  flex-wrap: wrap;
}

.right-actions,
.left-actions {
  display: flex;
  gap: 8px;
}

.btn {
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 10px 14px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.btn-primary {
  background: #3182ce;
  border-color: #3182ce;
  color: #fff;
}

.btn-primary:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.btn-secondary,
.btn-ghost {
  background: #111827;
  color: #cbd5e0;
}

.btn-secondary:hover,
.btn-ghost:hover {
  border-color: #4299e1;
  color: #fff;
}

.btn-danger {
  background: rgba(220, 38, 38, 0.12);
  color: #fca5a5;
}

@media (max-width: 900px) {
  .detail-header,
  .footer-actions {
    flex-direction: column;
    align-items: stretch;
  }

  .header-actions {
    align-items: stretch;
  }

  .action-row,
  .right-actions,
  .left-actions,
  .readonly-banner {
    flex-wrap: wrap;
  }
}
</style>
