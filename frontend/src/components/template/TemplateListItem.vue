<template>
  <button class="template-item" :class="{ selected }" @click="$emit('select')">
    <div class="item-top">
      <div class="item-title-wrap">
        <span class="item-title">{{ template.name }}</span>
        <span class="item-updated">{{ formatTemplateDate(template.updatedAt) }}</span>
      </div>

      <div class="item-actions" @click.stop>
        <button class="action-btn" title="Duplicate" @click="$emit('duplicate')">Copy</button>
        <button
          class="action-btn danger"
          :disabled="!['user', 'test'].includes(template.scope)"
          title="Delete"
          @click="$emit('delete')"
        >
          Delete
        </button>
      </div>
    </div>

    <p class="item-description">{{ template.description }}</p>

    <div class="tag-row">
      <span class="tag db">{{ dbLabels[template.dbFamily] || template.dbFamily }}</span>
      <span class="tag tool">{{ toolLabels[template.tool] || template.tool }}</span>
      <span class="tag scope" :class="template.scope">{{ scopeLabels[template.scope] }}</span>
      <span v-for="tag in template.tags.slice(0, 2)" :key="tag" class="tag neutral">{{ tag }}</span>
    </div>
  </button>
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

defineEmits(['select', 'duplicate', 'delete'])

const templateStore = useTemplateStore()
const toolLabels = computed(() => templateStore.toolLabels)
const dbLabels = computed(() => templateStore.dbFamilyLabels)
const scopeLabels = computed(() => templateStore.scopeLabels)
</script>

<style scoped>
.template-item {
  width: 100%;
  text-align: left;
  border: 1px solid #1f2937;
  border-radius: 12px;
  padding: 14px;
  background: #111827;
  color: inherit;
  cursor: pointer;
  transition: all 0.15s ease;
}

.template-item:hover {
  border-color: #334155;
  background: #172032;
}

.template-item.selected {
  border-color: #4299e1;
  background: rgba(30, 64, 175, 0.22);
  box-shadow: inset 0 0 0 1px rgba(96, 165, 250, 0.24);
}

.item-top {
  display: flex;
  gap: 12px;
  justify-content: space-between;
  align-items: flex-start;
}

.item-title-wrap {
  min-width: 0;
}

.item-title {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: #f8fafc;
}

.item-updated {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  color: #64748b;
}

.item-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.action-btn {
  border: none;
  border-radius: 6px;
  padding: 4px 8px;
  background: #1f2937;
  color: #cbd5e0;
  font-size: 11px;
  cursor: pointer;
}

.action-btn:hover:not(:disabled) {
  background: #334155;
}

.action-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.action-btn.danger:hover:not(:disabled) {
  background: rgba(220, 38, 38, 0.2);
  color: #fca5a5;
}

.item-description {
  margin-top: 10px;
  font-size: 12px;
  color: #94a3b8;
  line-height: 1.5;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 12px;
}

.tag {
  display: inline-flex;
  align-items: center;
  padding: 3px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
}

.tag.db {
  background: rgba(59, 130, 246, 0.15);
  color: #93c5fd;
}

.tag.tool {
  background: rgba(16, 185, 129, 0.15);
  color: #6ee7b7;
}

.tag.scope {
  background: rgba(148, 163, 184, 0.15);
  color: #e2e8f0;
}

.tag.scope.user {
  background: rgba(245, 158, 11, 0.15);
  color: #fcd34d;
}

.tag.scope.project {
  background: rgba(56, 189, 248, 0.16);
  color: #7dd3fc;
}

.tag.scope.readonlyShared {
  background: rgba(148, 163, 184, 0.18);
  color: #e2e8f0;
}

.tag.scope.test {
  background: rgba(251, 191, 36, 0.16);
  color: #fde68a;
}

.tag.neutral {
  background: rgba(100, 116, 139, 0.18);
  color: #cbd5e1;
}
</style>
