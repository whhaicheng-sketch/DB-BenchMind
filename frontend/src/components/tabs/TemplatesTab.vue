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

    <div class="templates-workbench">
      <TemplateList
        :templates="templateStore.filteredTemplates"
        :selected-id="templateStore.selectedTemplateId"
        :all-count="templateStore.displayTemplates.length"
        :has-filters="templateStore.hasActiveFilters"
        :has-search-only="templateStore.hasSearchOnly"
        @select="templateStore.selectTemplate"
        @duplicate="templateStore.duplicateTemplate"
        @delete="templateStore.deleteTemplate"
        @reset-filters="templateStore.resetFilters()"
      />

      <TemplateDetailPanel />
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { DB_OPTIONS, TOOL_OPTIONS } from '../../constants/templateCapabilities'
import TemplateDetailPanel from '../template/TemplateDetailPanel.vue'
import TemplateFilterBar from '../template/TemplateFilterBar.vue'
import TemplateHeader from '../template/TemplateHeader.vue'
import TemplateList from '../template/TemplateList.vue'
import { useTemplateStore } from '../../stores/template'

const templateStore = useTemplateStore()

const toolOptions = TOOL_OPTIONS
const dbOptions = DB_OPTIONS

const scopeOptions = [
  { value: 'builtin', label: 'Built-in' },
  { value: 'user', label: 'User' }
]

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

.templates-workbench {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(320px, 380px) minmax(0, 1fr);
  gap: 16px;
}

@media (max-width: 1180px) {
  .templates-workbench {
    grid-template-columns: 1fr;
  }
}
</style>
