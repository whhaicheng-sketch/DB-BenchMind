<template>
  <article class="template-item" :class="{ selected }" @click="$emit('open')">
    <div class="item-main">
      <div class="item-title-row">
        <span class="item-title">{{ template.name }}</span>
        <span class="status-pill" :class="template.status">{{ statusLabels[template.status] }}</span>
      </div>
      <p class="item-description">{{ template.description }}</p>
    </div>

    <div class="item-meta">
      <span class="tag db">{{ dbLabels[template.dbFamily] || template.dbFamily }}</span>
      <span class="tag tool">{{ toolLabels[template.tool] || template.tool }}</span>
      <span class="tag scope" :class="template.scope">{{ scopeLabels[template.scope] }}</span>
      <span v-for="tag in template.tags.slice(0, 2)" :key="tag" class="tag neutral">{{ tag }}</span>
    </div>

    <div class="item-updated">
      <span class="updated-label">Updated</span>
      <span>{{ formatTemplateDate(template.updatedAt) }}</span>
    </div>

    <div class="item-actions" @click.stop>
      <button class="btn-action btn-primary" title="View or Edit" @click="$emit('open')">View</button>
      <button class="btn-action" title="Duplicate" @click="$emit('duplicate')">Copy</button>
      <button
        class="btn-action btn-danger"
        :disabled="!['user', 'test'].includes(template.scope)"
        title="Delete"
        @click="$emit('delete')"
      >
        Delete
      </button>
    </div>
  </article>
</template>

<script setup>
import { computed } from 'vue'
import { formatTemplateDate } from '../../models/template'
import { useTemplateStore } from '../../stores/template'

defineProps({
  template: {
    type: Object,
    required: true
  },
  selected: {
    type: Boolean,
    default: false
  }
})

defineEmits(['open', 'duplicate', 'delete'])

const templateStore = useTemplateStore()
const toolLabels = computed(() => templateStore.toolLabels)
const dbLabels = computed(() => templateStore.dbFamilyLabels)
const scopeLabels = computed(() => templateStore.scopeLabels)
const statusLabels = computed(() => templateStore.statusLabels)
</script>

<style scoped>
.template-item {
  width: 100%;
  display: grid;
  grid-template-columns: minmax(280px, 1.4fr) minmax(280px, 1fr) 140px 150px;
  gap: var(--spacing-md);
  align-items: center;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  background-color: var(--bg-primary);
  color: var(--text-primary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.template-item:hover {
  border-color: var(--border-dark);
  background-color: var(--bg-hover);
}

.template-item.selected {
  border-color: var(--primary);
  background-color: var(--bg-selected);
}

.item-main {
  min-width: 0;
}

.item-title-row {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

.item-title {
  display: inline-block;
  font-size: var(--font-size-base);
  font-weight: 600;
  color: var(--text-primary);
}

/* Status Pills */
.status-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 10px;
  background-color: var(--bg-secondary);
  color: var(--text-muted);
  font-size: var(--font-size-xs);
  font-weight: 600;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.status-pill.ready {
  background-color: var(--success-bg);
  color: var(--success);
}

.status-pill.draft {
  background-color: var(--warning-bg);
  color: var(--warning);
}

.status-pill.deprecated {
  background-color: var(--danger-bg);
  color: var(--danger);
}

.item-description {
  margin-top: 4px;
  font-size: var(--font-size-sm);
  color: var(--text-muted);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* Meta Tags */
.item-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.tag {
  display: inline-flex;
  align-items: center;
  padding: 3px 8px;
  border-radius: 10px;
  font-size: var(--font-size-xs);
  font-weight: 500;
}

.tag.db {
  background-color: var(--primary-light);
  color: var(--primary);
}

.tag.tool {
  background-color: var(--success-bg);
  color: var(--success);
}

.tag.scope {
  background-color: var(--bg-secondary);
  color: var(--text-secondary);
}

.tag.scope.user {
  background-color: var(--warning-bg);
  color: var(--warning);
}

.tag.scope.project {
  background-color: var(--primary-light);
  color: var(--primary);
}

.tag.neutral {
  background-color: var(--bg-secondary);
  color: var(--text-muted);
  border: 1px solid var(--border-light);
}

/* Updated Column */
.item-updated {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.updated-label {
  color: var(--text-muted);
  font-size: var(--font-size-xs);
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

/* Actions */
.item-actions {
  display: flex;
  gap: var(--spacing-xs);
  justify-content: flex-end;
  flex-wrap: wrap;
}

.btn-action {
  padding: 5px 10px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background-color: var(--bg-primary);
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-action:hover:not(:disabled) {
  border-color: var(--border-dark);
  background-color: var(--bg-secondary);
  color: var(--text-primary);
}

.btn-action:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-action.btn-primary {
  background-color: var(--primary-light);
  border-color: #b3d8ff;
  color: var(--primary);
}

.btn-action.btn-primary:hover:not(:disabled) {
  background-color: var(--primary);
  border-color: var(--primary);
  color: white;
}

.btn-action.btn-danger:hover:not(:disabled) {
  background-color: var(--danger-bg);
  border-color: var(--danger);
  color: var(--danger);
}

@media (max-width: 1000px) {
  .template-item {
    grid-template-columns: 1fr;
  }

  .item-actions {
    justify-content: flex-start;
  }
}
</style>
