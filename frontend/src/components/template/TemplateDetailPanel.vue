<template>
  <section class="detail-panel">
    <div v-if="!activeTemplate" class="empty-wrap">
      <TemplateEmptyState
        title="No template selected"
        description="Choose a template from the left list to inspect it, or create a new user template draft."
        primary-label="Create Template"
        secondary-label="Import Template"
        @primary="templateStore.createTemplate()"
        @secondary="templateStore.placeholderAction('import')"
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
              v-if="templateStore.editorState === 'view' && templateStore.selectedTemplate?.scope === 'user'"
              class="btn btn-secondary"
              @click="templateStore.startEditing()"
            >
              Edit Template
            </button>
            <button
              v-else-if="templateStore.editorState === 'view'"
              class="btn btn-secondary"
              @click="templateStore.saveAsTemplate()"
            >
              Save As User Template
            </button>
            <button
              v-if="templateStore.selectedTemplate?.scope === 'user' && templateStore.editorState === 'view'"
              class="btn btn-danger"
              @click="templateStore.deleteTemplate()"
            >
              Delete
            </button>
          </div>
        </div>
      </div>

      <div class="detail-scroll">
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
          <button class="btn btn-secondary" @click="templateStore.saveAsTemplate()">Save As</button>
          <button class="btn btn-secondary" @click="templateStore.placeholderAction('createTask')">Create Task from Template</button>
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
import { useTemplateStore } from '../../stores/template'

const templateStore = useTemplateStore()

const editorModes = [
  { value: 'standard', label: 'Standard' },
  { value: 'advanced', label: 'Advanced' },
  { value: 'expert', label: 'Expert' }
]

const activeTemplate = computed(() => templateStore.activeTemplate)
const isReadOnly = computed(() => templateStore.editorState === 'view')
const editorLabel = computed(() => {
  if (templateStore.editorState === 'creating') return 'New Template'
  if (templateStore.editorState === 'editing') return 'Editing'
  return 'Overview'
})

const handleSave = () => {
  if (isReadOnly.value) {
    if (templateStore.selectedTemplate?.scope === 'user') {
      templateStore.showNotice('Switch to edit mode before saving updates.', 'info')
    } else {
      templateStore.placeholderAction('unsupportedEdit')
    }
    return
  }

  templateStore.saveTemplate()
}
</script>

<style scoped>
.detail-panel {
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
  .left-actions {
    flex-wrap: wrap;
  }
}
</style>
