<template>
  <section class="template-list-panel">
    <div class="panel-header">
      <div class="header-copy">
        <h2 class="panel-title">Template Library</h2>
        <p class="panel-subtitle">{{ templates.length }} visible / {{ allCount }} total</p>
      </div>
      <div v-if="templates.length > 0" class="table-head">
        <span>Name</span>
        <span>Config</span>
        <span>Actions</span>
      </div>
    </div>

    <div v-if="allCount === 0" class="state-wrap">
      <TemplateEmptyState
        title="No templates yet"
        description="Create a template to get started."
        primary-label="Create Template"
        @primary="templateStore.createTemplate()"
      />
    </div>

    <div v-else-if="templates.length === 0" class="state-wrap">
      <TemplateEmptyState
        :title="hasSearchOnly ? 'No search results' : 'No templates match current filters'"
        :description="hasSearchOnly ? 'Try another keyword.' : 'Change the database or tool filter.'"
        primary-label="Reset Filters"
        @primary="$emit('reset-filters')"
      />
    </div>

    <div v-else class="list-scroll">
      <TemplateListItem
        v-for="template in templates"
        :key="template.id"
        :template="template"
        :selected="editorOpen && template.id === selectedId"
        @open="$emit('open', template.id)"
        @duplicate="$emit('duplicate', template.id)"
        @delete="$emit('delete', template.id)"
      />
    </div>
  </section>
</template>

<script setup>
import TemplateEmptyState from './TemplateEmptyState.vue'
import TemplateListItem from './TemplateListItem.vue'
import { useTemplateStore } from '../../stores/template'

defineProps({
  templates: {
    type: Array,
    default: () => []
  },
  selectedId: {
    type: String,
    default: ''
  },
  editorOpen: {
    type: Boolean,
    default: false
  },
  allCount: {
    type: Number,
    default: 0
  },
  hasFilters: {
    type: Boolean,
    default: false
  },
  hasSearchOnly: {
    type: Boolean,
    default: false
  }
})

defineEmits(['open', 'duplicate', 'delete', 'reset-filters'])

const templateStore = useTemplateStore()
</script>

<style scoped>
.template-list-panel {
  min-height: 0;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  background-color: var(--bg-primary);
  overflow: hidden;
}

.panel-header {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-light);
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 10px;
  flex-wrap: wrap;
  background-color: var(--bg-secondary);
}

.panel-title {
  font-size: var(--font-size-base);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.panel-subtitle {
  margin-top: 2px;
  color: var(--text-muted);
  font-size: var(--font-size-xs);
}

.table-head {
  min-width: 420px;
  display: grid;
  grid-template-columns: minmax(280px, 1.6fr) minmax(240px, 1fr) 150px;
  gap: 10px;
  font-size: var(--font-size-xs);
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-muted);
}

.list-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 6px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.state-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

@media (max-width: 920px) {
  .table-head {
    display: none;
    min-width: 0;
  }
}
</style>
