<template>
  <section class="section-card preview-card">
    <div class="section-header">
      <div>
        <h3 class="section-title">Preview Summary</h3>
        <p class="section-subtitle">A compact summary of the scenario that will later be bound to a connection in Tasks & Monitor.</p>
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

    <div class="preview-notes">
      <div class="preview-pill-group">
        <span v-for="phase in enabledPhases" :key="phase" class="preview-pill">{{ phase }}</span>
      </div>
      <p class="preview-text">
        {{ summaryText }}
      </p>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue'
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

const enabledPhases = computed(() => Object.entries(props.templateModel.phases)
  .filter(([, config]) => config.enabled)
  .map(([phase]) => phase))

const summaryText = computed(() => {
  return `${templateStore.toolLabels[props.templateModel.tool]} will run ${templateStore.workloadLabels[props.templateModel.workloadFamily] || props.templateModel.workloadFamily} against ${templateStore.dbFamilyLabels[props.templateModel.dbFamily]} with ${props.templateModel.runtime.concurrency.value} ${concurrencyLabels[props.templateModel.runtime.concurrency.mode].toLowerCase()} for ${props.templateModel.runtime.durationSeconds} seconds. Connection binding, persistence and execution mapping remain reserved for later phases.`
})
</script>

<style scoped>
.section-card {
  border: 1px solid #1f2937;
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.95), rgba(15, 23, 42, 0.95));
  padding: 16px;
}

.section-title {
  font-size: 16px;
  color: #f8fafc;
}

.section-subtitle {
  margin-top: 4px;
  color: #64748b;
  font-size: 12px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-top: 16px;
}

.summary-item {
  padding: 12px;
  border: 1px solid #1e293b;
  border-radius: 10px;
  background: rgba(15, 23, 42, 0.6);
}

.summary-label {
  display: block;
  margin-bottom: 4px;
  color: #64748b;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.preview-notes {
  margin-top: 16px;
  padding: 14px;
  border-radius: 10px;
  background: rgba(15, 23, 42, 0.7);
  border: 1px solid #1e293b;
}

.preview-pill-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 10px;
}

.preview-pill {
  padding: 4px 8px;
  border-radius: 999px;
  background: rgba(96, 165, 250, 0.14);
  color: #93c5fd;
  font-size: 11px;
  text-transform: capitalize;
  font-weight: 700;
}

.preview-text {
  color: #cbd5e1;
  font-size: 13px;
  line-height: 1.6;
}

@media (max-width: 980px) {
  .summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>
