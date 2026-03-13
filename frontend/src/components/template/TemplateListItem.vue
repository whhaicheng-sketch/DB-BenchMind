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
      <button class="action-btn action-btn-primary" title="View or Edit" @click="$emit('open')">View</button>
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
  gap: 14px;
  align-items: center;
  border: 1px solid #1f2937;
  border-radius: 14px;
  padding: 14px 16px;
  background: linear-gradient(180deg, rgba(17, 24, 39, 0.98), rgba(15, 23, 42, 0.98));
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

.item-main {
  min-width: 0;
}

.item-title-row {
  display: flex;
  gap: 10px;
  align-items: center;
}

.item-title {
  display: inline-block;
  font-size: 15px;
  font-weight: 700;
  color: #f8fafc;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.12);
  color: #cbd5e1;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.status-pill.ready {
  background: rgba(34, 197, 94, 0.14);
  color: #86efac;
}

.status-pill.draft {
  background: rgba(245, 158, 11, 0.15);
  color: #fcd34d;
}

.status-pill.deprecated {
  background: rgba(248, 113, 113, 0.16);
  color: #fca5a5;
}

.item-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.item-updated {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  color: #cbd5e1;
}

.updated-label {
  color: #64748b;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.item-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  flex-wrap: wrap;
}

.action-btn {
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 8px 10px;
  background: #0f172a;
  color: #cbd5e0;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.action-btn-primary {
  background: rgba(49, 130, 206, 0.18);
  border-color: rgba(96, 165, 250, 0.28);
  color: #bfdbfe;
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
  margin-top: 8px;
  font-size: 13px;
  color: #94a3b8;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
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

@media (max-width: 1100px) {
  .template-item {
    grid-template-columns: 1fr;
  }

  .item-actions {
    justify-content: flex-start;
  }
}
</style>
