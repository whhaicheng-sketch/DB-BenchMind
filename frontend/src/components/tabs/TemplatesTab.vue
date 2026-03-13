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
      :tag-options="tagOptions"
      :scope-options="scopeOptions"
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

    <div v-if="templateStore.deleteCandidate" class="confirm-overlay" @click.self="templateStore.cancelDeleteTemplate()">
      <div class="confirm-modal">
        <div class="confirm-title">Delete Template</div>
        <p class="confirm-body">
          Delete editable template "{{ templateStore.deleteCandidate.name }}"? This cannot be undone.
        </p>
        <div class="confirm-actions">
          <button class="btn btn-ghost" @click="templateStore.cancelDeleteTemplate()">Cancel</button>
          <button class="btn btn-danger" @click="templateStore.confirmDeleteTemplate()">Delete</button>
        </div>
      </div>
    </div>

    <TemplateEditorDialog />
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { DB_OPTIONS, TOOL_OPTIONS } from '../../constants/templateCapabilities'
import TemplateEditorDialog from '../template/TemplateEditorDialog.vue'
import TemplateFilterBar from '../template/TemplateFilterBar.vue'
import TemplateHeader from '../template/TemplateHeader.vue'
import TemplateList from '../template/TemplateList.vue'
import { useTemplateStore } from '../../stores/template'

const templateStore = useTemplateStore()

const toolOptions = TOOL_OPTIONS
const dbOptions = DB_OPTIONS

const scopeOptions = Object.entries(templateStore.scopeLabels).map(([value, label]) => ({ value, label }))

const tagOptions = computed(() => templateStore.allTags.map((tag) => ({ value: tag, label: tag })))

const handleFilterChange = ({ key, value }) => {
  templateStore.setFilter(key, value)
}

const handleCreate = () => {
  templateStore.createTemplate()
}

onMounted(async () => {
  await templateStore.initializeTemplates()
})
</script>

<style scoped>
.templates-tab {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 0;
}

.notice-banner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border-radius: 8px;
  border: 1px solid transparent;
  font-size: 13px;
}

.notice-info {
  background: rgba(66, 153, 225, 0.12);
  border-color: rgba(66, 153, 225, 0.3);
  color: #90cdf4;
}

.notice-success {
  background: rgba(72, 187, 120, 0.12);
  border-color: rgba(72, 187, 120, 0.28);
  color: #9ae6b4;
}

.notice-warning {
  background: rgba(237, 137, 54, 0.12);
  border-color: rgba(237, 137, 54, 0.28);
  color: #f6ad55;
}

.notice-dismiss {
  border: none;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font-size: 12px;
}

.confirm-overlay {
  position: fixed;
  inset: 0;
  z-index: 90;
  background: rgba(2, 6, 23, 0.72);
  display: flex;
  align-items: center;
  justify-content: center;
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

.btn {
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 10px 14px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
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
</style>
