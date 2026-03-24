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
    </div>

    <div v-if="templateModel.runtime.notes" class="preview-notes">
      <span class="summary-label">Runtime Notes</span>
      <strong>{{ templateModel.runtime.notes }}</strong>
    </div>
  </section>
</template>

<script setup>
import { CONCURRENCY_MODE_LABELS } from '../../models/template'
import { useTemplateStore } from '../../stores/template'

const props = defineProps({
  templateModel: {
    type: Object,
    required: true
  }
})

const templateStore = useTemplateStore()
const concurrencyLabels = CONCURRENCY_MODE_LABELS
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

@media (max-width: 980px) {
  .summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>
