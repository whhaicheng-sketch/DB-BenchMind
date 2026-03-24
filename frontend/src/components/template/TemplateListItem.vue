<template>
  <article class="template-item" :class="{ selected }" @click="$emit('open')">
    <div class="item-main">
      <div class="item-title-row">
        <span class="item-title">{{ template.name }}</span>
      </div>
      <p class="item-description">{{ template.description }}</p>
    </div>

    <div class="item-meta">
      <span class="tag db">{{ dbLabels[template.dbFamily] || template.dbFamily }}</span>
      <span class="tag tool">{{ toolLabels[template.tool] || template.tool }}</span>
      <span class="tag neutral">{{ workloadLabels[template.workloadFamily] || template.workloadFamily }}</span>
    </div>

    <div class="item-actions" @click.stop>
      <button class="btn-action btn-primary" title="Open template" @click="$emit('open')">{{ primaryActionLabel }}</button>
      <button class="btn-action" title="Duplicate" @click="$emit('duplicate')">Copy</button>
      <button
        v-if="!template.is_builtin"
        class="btn-action btn-danger"
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
import { useTemplateStore } from '../../stores/template'

const props = defineProps({
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
const workloadLabels = computed(() => templateStore.workloadLabels)
const primaryActionLabel = computed(() => (props.template.is_builtin ? 'View' : 'Edit'))
</script>

<style scoped>
.template-item {
  width: 100%;
  display: grid;
  grid-template-columns: minmax(280px, 1.6fr) minmax(220px, 1fr) 150px;
  gap: 10px;
  align-items: center;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 10px 12px;
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
.item-description {
  margin-top: 4px;
  font-size: var(--font-size-sm);
  color: var(--text-muted);
  line-height: 1.4;
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

.tag.neutral {
  background-color: var(--bg-secondary);
  color: var(--text-muted);
  border: 1px solid var(--border-light);
}

/* Actions */
.item-actions {
  display: flex;
  gap: 6px;
  justify-content: flex-end;
  flex-wrap: wrap;
}

.btn-action {
  padding: 4px 8px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background-color: var(--bg-primary);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-action:hover:not(:disabled) {
  border-color: var(--border-dark);
  background-color: var(--bg-secondary);
  color: var(--text-primary);
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
