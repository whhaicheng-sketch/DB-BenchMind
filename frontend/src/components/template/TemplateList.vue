<template>
  <section class="template-list-panel">
    <div class="panel-header">
      <div>
        <h2 class="panel-title">Template Library</h2>
        <p class="panel-subtitle">{{ templates.length }} visible / {{ allCount }} total</p>
      </div>
    </div>

    <div v-if="allCount === 0" class="state-wrap">
      <TemplateEmptyState
        title="No templates yet"
        description="Create your first benchmark scenario template to start building reusable workload definitions."
        primary-label="Create Template"
        secondary-label="Import Template"
        @primary="templateStore.createTemplate()"
        @secondary="templateStore.placeholderAction('import')"
      />
    </div>

    <div v-else-if="templates.length === 0" class="state-wrap">
      <TemplateEmptyState
        :title="hasSearchOnly ? 'No search results' : 'No templates match current filters'"
        :description="hasSearchOnly ? 'Try another keyword or clear the search box.' : 'Adjust database, tool, scope or tag filters to broaden the result set.'"
        primary-label="Reset Filters"
        @primary="$emit('reset-filters')"
      />
    </div>

    <div v-else class="list-scroll">
      <TemplateListItem
        v-for="template in templates"
        :key="template.id"
        :template="template"
        :selected="template.id === selectedId"
        @select="$emit('select', template.id)"
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

defineEmits(['select', 'duplicate', 'delete', 'reset-filters'])

const templateStore = useTemplateStore()
</script>

<style scoped>
.template-list-panel {
  min-height: 0;
  display: flex;
  flex-direction: column;
  border: 1px solid #2d3748;
  border-radius: 14px;
  background: linear-gradient(180deg, #111827 0%, #0f172a 100%);
  overflow: hidden;
}

.panel-header {
  padding: 16px;
  border-bottom: 1px solid #1f2937;
}

.panel-title {
  font-size: 18px;
  font-weight: 600;
  color: #f8fafc;
}

.panel-subtitle {
  margin-top: 4px;
  color: #64748b;
  font-size: 12px;
}

.list-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.state-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}
</style>
