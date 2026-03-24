<template>
  <section class="section-card preview-card">
    <div class="section-header">
      <div>
        <h3 class="section-title">Preview</h3>
      </div>
    </div>

    <div class="summary-grid">
      <div class="summary-item">
        <span class="summary-label">DB Type</span>
        <strong>{{ templateStore.dbFamilyLabels[templateModel.dbFamily] }}</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">Tool</span>
        <strong>{{ templateStore.toolLabels[templateModel.tool] }}</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">Workload</span>
        <strong>{{ templateStore.workloadLabels[templateModel.workloadFamily] || templateModel.workloadFamily }}</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">Profile</span>
        <strong>{{ profileLabel }}</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">Source Alignment</span>
        <strong>{{ sourceAlignmentLabel }}</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">Concurrency</span>
        <strong>{{ templateModel.runtime.concurrency.value }} {{ concurrencyLabels[templateModel.runtime.concurrency.mode] }}</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">Duration</span>
        <strong>{{ templateModel.runtime.durationSeconds }}s</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">Warm-up / Ramp-up</span>
        <strong>{{ templateModel.runtime.warmupSeconds }}s / {{ templateModel.runtime.rampUpSeconds }}s</strong>
      </div>
      <div class="summary-item">
        <span class="summary-label">Metrics</span>
        <strong>{{ metricsText }}</strong>
      </div>
    </div>

    <div v-if="templateModel.goal" class="preview-notes">
      <span class="summary-label">Goal</span>
      <strong>{{ templateModel.goal }}</strong>
    </div>

    <div v-if="templateModel.runtime.notes" class="preview-notes">
      <span class="summary-label">Runtime Notes</span>
      <strong>{{ templateModel.runtime.notes }}</strong>
    </div>

    <div class="config-grid">
      <div class="config-card">
        <span class="summary-label">Prepare</span>
        <pre>{{ formatConfig(templateModel.prepare_config) }}</pre>
      </div>
      <div class="config-card">
        <span class="summary-label">Run</span>
        <pre>{{ formatConfig(templateModel.run_config) }}</pre>
      </div>
      <div class="config-card">
        <span class="summary-label">Cleanup</span>
        <pre>{{ formatConfig(templateModel.cleanup_config) }}</pre>
      </div>
    </div>

    <div v-if="templateModel.tags?.length" class="tag-row">
      <span v-for="tag in decoratedTags" :key="tag" class="tag-chip">{{ tag }}</span>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import { CONCURRENCY_MODE_LABELS, PROFILE_TYPE_LABELS, SOURCE_ALIGNMENT_LABELS } from '../../models/template'
import { useTemplateStore } from '../../stores/template'

const props = defineProps({
  templateModel: {
    type: Object,
    required: true
  }
})

const templateStore = useTemplateStore()
const concurrencyLabels = CONCURRENCY_MODE_LABELS

const profileLabel = computed(() => PROFILE_TYPE_LABELS[props.templateModel.profile_type] || props.templateModel.profile_type || 'Custom')
const sourceAlignmentLabel = computed(() => SOURCE_ALIGNMENT_LABELS[props.templateModel.source_alignment] || props.templateModel.source_alignment || 'Custom')
const metricsText = computed(() => props.templateModel.metrics?.join(', ') || 'N/A')
const decoratedTags = computed(() => {
  const tags = [...(props.templateModel.tags || [])]
  if (props.templateModel.test_position === 'smoke' && !tags.includes('Smoke')) {
    tags.unshift('Smoke')
  }
  if (props.templateModel.source_alignment?.includes('engineered') && !tags.includes('Engineered')) {
    tags.unshift('Engineered')
  }
  return tags
})

function formatConfig(config) {
  if (!config || Object.keys(config).length === 0) return 'N/A'
  return JSON.stringify(config, null, 2)
}
</script>

<style scoped>
.section-card {
  border: 1px solid var(--border-color);
  border-radius: 12px;
  background: var(--bg-primary);
  padding: 10px;
}

.section-title {
  font-size: 14px;
  color: var(--text-primary);
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.summary-item {
  padding: 10px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background: var(--bg-secondary);
}

.config-card {
  padding: 10px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background: var(--bg-secondary);
}

.summary-label {
  display: block;
  margin-bottom: 4px;
  color: var(--text-muted);
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.preview-notes {
  display: grid;
  gap: 4px;
  margin-top: 8px;
  padding: 8px;
  border-radius: 10px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
}

.config-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  margin-top: 8px;
}

.config-card pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 11px;
  color: var(--text-secondary);
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.tag-chip {
  display: inline-flex;
  align-items: center;
  padding: 3px 8px;
  border-radius: 999px;
  background: var(--primary-light);
  color: var(--primary);
  font-size: 11px;
  font-weight: 600;
}

@media (max-width: 980px) {
  .summary-grid,
  .config-grid {
    grid-template-columns: 1fr;
  }
}
</style>
