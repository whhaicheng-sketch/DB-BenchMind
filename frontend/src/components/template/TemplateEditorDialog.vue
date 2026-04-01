<template>
  <div
    v-if="templateStore.isEditorOpen && activeTemplate"
    class="modal-overlay"
  >
    <section class="editor-dialog">
      <header class="dialog-header">
        <div class="header-main">
          <div class="badge-row">
            <span class="mode-badge">{{ editorLabel }}</span>
            <span class="mode-badge badge-db">{{ templateStore.dbFamilyLabels[activeTemplate.dbFamily] }}</span>
            <span class="mode-badge badge-tool">{{ templateStore.toolLabels[activeTemplate.tool] }}</span>
            <span v-if="activeTemplate.is_builtin" class="mode-badge muted">Built-in</span>
            <span v-if="templateStore.isDirty" class="mode-badge dirty">Unsaved changes</span>
          </div>
          <h2 class="dialog-title">{{ activeTemplate.name }}</h2>
        </div>
      </header>

      <div class="dialog-body">
        <div v-if="isReadonlyTemplate && templateStore.editorState === 'view'" class="readonly-banner">
          <div>
            <strong>Built-in template</strong>
            <span> Use Copy in the template list to create an editable duplicate.</span>
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
        <button
          class="btn btn-primary"
          :disabled="isSaveBlocked"
          @click="handleSave"
        >
          Save
        </button>
        <button class="btn btn-secondary" @click="templateStore.closeEditor()">Close</button>
      </footer>
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import TemplateBasicSection from './TemplateBasicSection.vue'
import TemplatePreviewSection from './TemplatePreviewSection.vue'
import TemplateRuntimeSection from './TemplateRuntimeSection.vue'
import { useTemplateStore } from '../../stores/template'

const templateStore = useTemplateStore()

const activeTemplate = computed(() => templateStore.activeTemplate)
const isReadOnly = computed(() => templateStore.editorState === 'view')
const isReadonlyTemplate = computed(() => !!templateStore.selectedTemplate?.is_builtin)
const isSaveBlocked = computed(() => isReadOnly.value && isReadonlyTemplate.value)
const editorLabel = computed(() => {
  if (templateStore.editorState === 'creating') return 'New Template'
  if (templateStore.editorState === 'editing') return 'Editing'
  return 'Overview'
})

const handleSave = () => {
  if (isReadOnly.value) {
    templateStore.showNotice('Built-in templates are read-only. Use Copy in the template list to create an editable duplicate.', 'info')
    return
  }

  templateStore.saveTemplate()
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 80;
  background: rgba(15, 23, 42, 0.24);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.editor-dialog {
  width: min(76vw, 1240px);
  height: min(82vh, 900px);
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  border-radius: 16px;
  border: 1px solid var(--border-color);
  background: var(--bg-primary);
  box-shadow: var(--shadow-modal);
  overflow: hidden;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 14px 10px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-primary);
}

.header-main {
  min-width: 0;
}

.badge-row {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
  margin-bottom: 6px;
}

.mode-badge {
  padding: 3px 8px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 700;
  background: var(--primary-light);
  color: var(--primary);
}

.mode-badge.badge-db {
  background: rgba(14, 165, 233, 0.12);
  color: #0369a1;
}

.mode-badge.badge-tool {
  background: var(--success-bg);
  color: var(--success);
}

.mode-badge.muted {
  background: var(--bg-secondary);
  color: var(--text-secondary);
}

.mode-badge.dirty {
  background: var(--warning-bg);
  color: var(--warning);
}

.dialog-title {
  font-size: 18px;
  line-height: 1.2;
  color: var(--text-primary);
}

.dialog-body {
  min-height: 0;
  overflow-y: auto;
  padding: 10px 14px 12px;
  background: var(--bg-secondary);
}

.content-stack {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.readonly-banner,
.validation-summary {
  margin-bottom: 12px;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  padding: 10px 12px;
}

.readonly-banner {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
  background: var(--info-bg);
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.5;
}

.validation-summary {
  background: var(--danger-bg);
  border-color: var(--danger-border);
}

.validation-title {
  color: var(--danger);
  font-size: 12px;
  font-weight: 700;
}

.validation-list {
  margin: 8px 0 0 16px;
  color: var(--danger);
  font-size: 11px;
  line-height: 1.5;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  align-items: center;
  padding: 8px 14px;
  border-top: 1px solid var(--border-color);
  background: var(--bg-primary);
}

.btn {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 7px 10px;
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  background: var(--bg-primary);
  color: var(--text-primary);
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
  background: var(--primary-hover);
}

.btn-secondary {
  background: var(--bg-primary);
  color: var(--text-primary);
}

.btn-secondary:hover:not(:disabled) {
  border-color: var(--primary);
  background: var(--bg-secondary);
  color: var(--text-primary);
}

@media (max-width: 1180px) {
  .editor-dialog {
    width: min(100vw - 24px, 1240px);
    height: min(88vh, 900px);
  }

  .dialog-header,
  .dialog-footer {
    flex-direction: column;
    align-items: stretch;
  }
}

@media (max-width: 760px) {
  .dialog-body {
    padding: 12px;
  }

  .dialog-header,
  .dialog-footer {
    padding: 12px;
  }
}
</style>
