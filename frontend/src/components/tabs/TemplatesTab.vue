<template>
  <div class="templates-tab">
    <TemplateHeader @create="handleCreate" />

    <div v-if="templateStore.notice" class="notice-banner" :class="`notice-${templateStore.notice.tone}`">
      <span>{{ templateStore.notice.message }}</span>
      <button class="notice-dismiss" @click="templateStore.clearNotice()">Dismiss</button>
    </div>

    <TemplateFilterBar
      :filters="templateStore.filters"
      :tool-options="toolOptions"
      :db-options="dbOptions"
      @filter-change="handleFilterChange"
      @reset="templateStore.resetFilters()"
    />

    <TemplateList
      :templates="templateStore.filteredTemplates"
      :selected-id="templateStore.selectedTemplateId"
      :editor-open="templateStore.isEditorOpen"
      :all-count="templateStore.displayTemplates.length"
      :has-filters="templateStore.hasActiveFilters"
      :has-search-only="templateStore.hasSearchOnly"
      @open="templateStore.openTemplate"
      @duplicate="templateStore.duplicateTemplate"
      @delete="templateStore.requestDeleteTemplate"
      @reset-filters="templateStore.resetFilters()"
    />

    <div v-if="templateStore.deleteCandidate" class="confirm-overlay">
      <div class="confirm-modal">
        <div class="confirm-title">Delete Template</div>
        <p class="confirm-body">
          Delete template "{{ templateStore.deleteCandidate.name }}"? This action cannot be undone.
        </p>
        <div class="confirm-actions">
          <button class="btn" @click="templateStore.cancelDeleteTemplate()">Cancel</button>
          <button class="btn btn-danger" @click="templateStore.confirmDeleteTemplate()">Delete</button>
        </div>
      </div>
    </div>

    <TemplateEditorDialog />
  </div>
</template>

<script setup>
import { onMounted, onActivated } from 'vue'
import { DB_OPTIONS, TOOL_OPTIONS } from '../../constants/templateCapabilities'
import TemplateEditorDialog from '../template/TemplateEditorDialog.vue'
import TemplateFilterBar from '../template/TemplateFilterBar.vue'
import TemplateHeader from '../template/TemplateHeader.vue'
import TemplateList from '../template/TemplateList.vue'
import { useTemplateStore } from '../../stores/template'

const templateStore = useTemplateStore()

const toolOptions = TOOL_OPTIONS
const dbOptions = DB_OPTIONS

const handleFilterChange = ({ key, value }) => {
  templateStore.setFilter(key, value)
}

const handleCreate = () => {
  templateStore.createTemplate()
}

onMounted(async () => {
  await templateStore.initializeTemplates()
})

// Refresh templates when switching back to this tab
onActivated(async () => {
  await templateStore.initializeTemplates()
})
</script>

<style scoped>
.templates-tab {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  min-height: 0;
}

/* Notice Banner */
.notice-banner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  font-size: var(--font-size-sm);
}

.notice-info {
  background: var(--info-bg);
  border-color: var(--border-color);
  color: var(--text-secondary);
}

.notice-success {
  background: var(--success-bg);
  border-color: var(--success-border);
  color: var(--success);
}

.notice-warning {
  background: var(--warning-bg);
  border-color: var(--warning-border);
  color: var(--warning);
}

.notice-dismiss {
  border: none;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font-size: var(--font-size-xs);
  opacity: 0.7;
}

.notice-dismiss:hover {
  opacity: 1;
}

/* Confirm Modal */
.confirm-overlay {
  position: fixed;
  inset: 0;
  z-index: 90;
  background-color: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-lg);
}

.confirm-modal {
  width: min(460px, 100%);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  background: var(--bg-primary);
  box-shadow: var(--shadow-modal);
  padding: var(--spacing-lg);
}

.confirm-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-primary);
}

.confirm-body {
  margin-top: var(--spacing-sm);
  color: var(--text-secondary);
  line-height: 1.6;
  font-size: var(--font-size-sm);
}

.confirm-actions {
  margin-top: var(--spacing-lg);
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-sm);
}

.btn {
  padding: 8px 16px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  background-color: var(--bg-primary);
  color: var(--text-primary);
  transition: all var(--transition-fast);
}

.btn:hover {
  background-color: var(--bg-secondary);
  border-color: var(--border-dark);
}

.btn-danger {
  background-color: var(--danger-bg);
  border-color: var(--danger-border);
  color: var(--danger);
}

.btn-danger:hover {
  background-color: var(--danger);
  border-color: var(--danger);
  color: white;
}
</style>
