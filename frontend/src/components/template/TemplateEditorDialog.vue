<template>
  <div
    v-if="templateStore.isEditorOpen && activeTemplate"
    class="modal-overlay"
    @click.self="templateStore.closeEditor()"
  >
    <section class="editor-dialog">
      <header class="dialog-header">
        <div class="header-main">
          <div class="badge-row">
            <span class="mode-badge">{{ editorLabel }}</span>
            <span class="mode-badge badge-db">{{ templateStore.dbFamilyLabels[activeTemplate.dbFamily] }}</span>
            <span class="mode-badge badge-tool">{{ templateStore.toolLabels[activeTemplate.tool] }}</span>
            <span class="mode-badge muted">{{ templateStore.scopeLabels[activeTemplate.scope] }}</span>
            <span class="mode-badge muted">{{ templateStore.statusLabels[activeTemplate.status] }}</span>
            <span v-if="templateStore.isDirty" class="mode-badge dirty">Unsaved changes</span>
          </div>
          <h2 class="dialog-title">{{ activeTemplate.name }}</h2>
          <p class="dialog-subtitle">{{ activeTemplate.description }}</p>
        </div>

        <div class="header-side">
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

          <button class="close-btn" aria-label="Close template editor" @click="templateStore.closeEditor()">Close</button>
        </div>
      </header>

      <div class="dialog-body">
        <div v-if="isReadonlyScope && templateStore.editorState === 'view'" class="readonly-banner">
          <div>
            <strong>{{ readonlyBannerTitle }}</strong>
            <span> {{ readonlyBannerBody }}</span>
          </div>
        </div>

        <div v-if="templateStore.hasValidationErrors" class="validation-summary">
          <div class="validation-title">Validation issues</div>
          <ul class="validation-list">
            <li v-for="(message, key) in templateStore.validationErrors" :key="key">{{ message }}</li>
          </ul>
        </div>

        <div class="content-stack">
          <TemplateBasicSection :template-model="activeTemplate" :readonly="isReadOnly" />
          <TemplateRuntimeSection :template-model="activeTemplate" :readonly="isReadOnly" />
          <TemplatePreviewSection :template-model="activeTemplate" />
        </div>
      </div>

      <footer class="dialog-footer">
        <div class="footer-left">
          <button
            v-if="templateStore.editorState === 'view' && canDirectEdit"
            class="btn btn-secondary"
            @click="templateStore.startEditing()"
          >
            Edit Template
          </button>
          <button
            v-if="canDeleteSelected"
            class="btn btn-danger"
            @click="templateStore.requestDeleteTemplate()"
          >
            Delete
          </button>
          <button
            v-if="templateStore.editorState === 'editing' || templateStore.editorState === 'creating'"
            class="btn btn-ghost"
            @click="handleCancel"
          >
            Cancel
          </button>
        </div>

        <div class="footer-right">
          <button
            class="btn btn-primary"
            :disabled="isSaveBlocked"
            @click="handleSave"
          >
            Save
          </button>
          <button class="btn btn-secondary" @click="handleSaveAs">Save As</button>
          <button class="btn btn-secondary" @click="handleCreateTask">Create Task from Template</button>
          <button class="btn btn-ghost" @click="templateStore.closeEditor()">Close</button>
        </div>
      </footer>
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import TemplateBasicSection from './TemplateBasicSection.vue'
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
const isSaveBlocked = computed(() => isReadOnly.value && isReadonlyScope.value)
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

const handleCancel = () => {
  if (templateStore.editorState === 'creating') {
    templateStore.closeEditor()
    return
  }

  templateStore.cancelEditing()
}

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

const handleSaveAs = () => {
  if (['builtin', 'readonlyShared'].includes(templateStore.selectedTemplate?.scope) && templateStore.editorState === 'view') {
    templateStore.placeholderAction('readonlySaveAs')
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
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 80;
  background: rgba(2, 6, 23, 0.76);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.editor-dialog {
  width: min(80vw, 1360px);
  height: min(86vh, 980px);
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  border-radius: 18px;
  border: 1px solid #334155;
  background: linear-gradient(180deg, #111827 0%, #0b1220 100%);
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.45);
  overflow: hidden;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 18px;
  padding: 20px 22px 18px;
  border-bottom: 1px solid #1f2937;
}

.header-main {
  min-width: 0;
}

.badge-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.mode-badge {
  padding: 4px 9px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
  background: rgba(59, 130, 246, 0.16);
  color: #93c5fd;
}

.mode-badge.badge-db {
  background: rgba(14, 165, 233, 0.16);
  color: #7dd3fc;
}

.mode-badge.badge-tool {
  background: rgba(16, 185, 129, 0.16);
  color: #6ee7b7;
}

.mode-badge.muted {
  background: rgba(148, 163, 184, 0.14);
  color: #cbd5e1;
}

.mode-badge.dirty {
  background: rgba(245, 158, 11, 0.18);
  color: #fcd34d;
}

.dialog-title {
  font-size: 24px;
  line-height: 1.2;
  color: #f8fafc;
}

.dialog-subtitle {
  margin-top: 8px;
  max-width: 860px;
  color: #94a3b8;
  font-size: 13px;
  line-height: 1.6;
}

.header-side {
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

.close-btn {
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 8px 12px;
  background: #0f172a;
  color: #cbd5e1;
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
}

.close-btn:hover {
  border-color: #4299e1;
  color: #fff;
}

.dialog-body {
  min-height: 0;
  overflow-y: auto;
  padding: 18px 22px 20px;
}

.content-stack {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.readonly-banner,
.validation-summary {
  margin-bottom: 16px;
  border-radius: 12px;
  border: 1px solid #334155;
  padding: 14px 16px;
}

.readonly-banner {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  background: rgba(15, 23, 42, 0.72);
  color: #cbd5e1;
  font-size: 13px;
  line-height: 1.6;
}

.validation-summary {
  background: rgba(127, 29, 29, 0.24);
  border-color: rgba(248, 113, 113, 0.3);
}

.validation-title {
  color: #fecaca;
  font-size: 13px;
  font-weight: 700;
}

.validation-list {
  margin: 10px 0 0 18px;
  color: #fecaca;
  font-size: 12px;
  line-height: 1.7;
}

.dialog-footer {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  padding: 16px 22px 18px;
  border-top: 1px solid #1f2937;
  background: rgba(10, 15, 27, 0.94);
}

.footer-left,
.footer-right {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.btn {
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 10px 14px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.btn-primary {
  background: #3182ce;
  border-color: #3182ce;
  color: #fff;
}

.btn-primary:hover:not(:disabled) {
  background: #2b6cb0;
}

.btn-secondary {
  background: #111827;
  color: #cbd5e0;
}

.btn-secondary:hover:not(:disabled),
.btn-ghost:hover:not(:disabled) {
  border-color: #4299e1;
  color: #fff;
}

.btn-ghost {
  background: #0f172a;
  color: #cbd5e1;
}

.btn-danger {
  background: rgba(127, 29, 29, 0.2);
  border-color: rgba(248, 113, 113, 0.28);
  color: #fca5a5;
}

.btn-danger:hover:not(:disabled) {
  background: rgba(153, 27, 27, 0.28);
}

@media (max-width: 1180px) {
  .modal-overlay {
    padding: 16px;
  }

  .editor-dialog {
    width: min(100vw - 32px, 1360px);
    height: min(90vh, 980px);
  }

  .dialog-header,
  .dialog-footer {
    flex-direction: column;
    align-items: stretch;
  }

  .header-side {
    align-items: stretch;
  }

  .footer-right,
  .footer-left {
    width: 100%;
  }
}

@media (max-width: 760px) {
  .mode-switch {
    width: 100%;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .dialog-body {
    padding: 16px;
  }

  .dialog-header,
  .dialog-footer {
    padding: 16px;
  }
}
</style>
