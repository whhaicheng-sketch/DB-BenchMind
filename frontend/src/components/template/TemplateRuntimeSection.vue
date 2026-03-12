<template>
  <section class="section-card">
    <div class="section-header">
      <div>
        <h3 class="section-title">Workload / Benchmark Settings</h3>
        <p class="section-subtitle">Phase switches, runtime controls and tool-specific benchmark parameters.</p>
      </div>
    </div>

    <div class="phase-grid">
      <label v-for="phase in phaseKeys" :key="phase" class="phase-pill" :class="{ active: templateModel.phases[phase]?.enabled }">
        <input
          type="checkbox"
          :checked="templateModel.phases[phase]?.enabled"
          :disabled="readonly"
          @change="handlePhaseToggle(phase, $event.target.checked)"
        >
        <span>{{ phase }}</span>
      </label>
    </div>

    <div class="runtime-grid">
      <label class="field">
        <span class="field-label">Concurrency Mode</span>
        <select
          class="field-input"
          v-model="templateModel.runtime.concurrency.mode"
          :disabled="readonly"
          @change="templateStore.markDirty()"
        >
          <option v-for="mode in availableModes" :key="mode" :value="mode">{{ concurrencyLabels[mode] }}</option>
        </select>
      </label>

      <label class="field">
        <span class="field-label">Concurrency Value</span>
        <input
          v-model.number="templateModel.runtime.concurrency.value"
          class="field-input"
          type="number"
          min="1"
          :disabled="readonly"
          @input="templateStore.markDirty()"
        >
      </label>

      <label class="field">
        <span class="field-label">Duration (s)</span>
        <input v-model.number="templateModel.runtime.durationSeconds" class="field-input" type="number" min="1" :disabled="readonly" @input="templateStore.markDirty()">
      </label>

      <label class="field">
        <span class="field-label">Warm-up (s)</span>
        <input v-model.number="templateModel.runtime.warmupSeconds" class="field-input" type="number" min="0" :disabled="readonly" @input="templateStore.markDirty()">
      </label>

      <label class="field">
        <span class="field-label">Ramp-up (s)</span>
        <input v-model.number="templateModel.runtime.rampUpSeconds" class="field-input" type="number" min="0" :disabled="readonly" @input="templateStore.markDirty()">
      </label>

      <label class="field">
        <span class="field-label">Report Interval (s)</span>
        <input v-model.number="templateModel.runtime.reportIntervalSeconds" class="field-input" type="number" min="1" :disabled="readonly" @input="templateStore.markDirty()">
      </label>

      <label class="field">
        <span class="field-label">Percentile</span>
        <input v-model.number="templateModel.runtime.percentile" class="field-input" type="number" min="1" max="100" :disabled="readonly" @input="templateStore.markDirty()">
      </label>

      <label class="field">
        <span class="field-label">Iterations</span>
        <input v-model.number="templateModel.runtime.iterations" class="field-input" type="number" min="0" :disabled="readonly" @input="templateStore.markDirty()">
      </label>

      <label class="field">
        <span class="field-label">Rate Limit</span>
        <input v-model.number="templateModel.runtime.rateLimit" class="field-input" type="number" min="0" :disabled="readonly" @input="templateStore.markDirty()">
      </label>
    </div>

    <div class="tool-fields">
      <div class="tool-header">
        <h4>{{ templateStore.toolLabels[templateModel.tool] }} Specific</h4>
        <span class="tool-note">{{ templateStore.editorMode === 'standard' ? 'High-frequency fields only' : 'Expanded field set placeholder for later phases' }}</span>
      </div>

      <div class="runtime-grid">
        <template v-for="field in visibleToolFields" :key="field.key">
          <label class="field" :class="{ 'field-wide': field.type === 'textarea' }">
            <span class="field-label">{{ field.label }}</span>
            <select
              v-if="field.type === 'select'"
              v-model="toolModel[field.key]"
              class="field-input"
              :disabled="readonly"
              @change="handleToolFieldChange(field)"
            >
              <option v-for="option in field.options" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>

            <textarea
              v-else-if="field.type === 'textarea'"
              v-model="toolModel[field.key]"
              class="field-input textarea"
              :disabled="readonly"
              @input="templateStore.markDirty()"
            />

            <input
              v-else
              v-model.number="toolModel[field.key]"
              class="field-input"
              type="number"
              :min="field.min || 0"
              :disabled="readonly"
              @input="templateStore.markDirty()"
            >
          </label>
        </template>

        <label class="field field-wide">
          <span class="field-label">Runtime Notes</span>
          <textarea
            v-model="templateModel.runtime.notes"
            class="field-input textarea"
            :disabled="readonly"
            placeholder="Add operator guidance or future backend mapping notes"
            @input="templateStore.markDirty()"
          />
        </label>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import { TEMPLATE_CAPABILITIES } from '../../constants/templateCapabilities'
import { CONCURRENCY_MODE_LABELS, PHASE_KEYS } from '../../models/template'
import { useTemplateStore } from '../../stores/template'

const props = defineProps({
  templateModel: {
    type: Object,
    required: true
  },
  readonly: {
    type: Boolean,
    default: false
  }
})

const templateStore = useTemplateStore()

const phaseKeys = PHASE_KEYS
const concurrencyLabels = CONCURRENCY_MODE_LABELS

const capability = computed(() => TEMPLATE_CAPABILITIES[props.templateModel.tool] || TEMPLATE_CAPABILITIES.sysbench)
const availableModes = computed(() => capability.value.concurrencyModes)
const toolModel = computed(() => props.templateModel.toolConfig[props.templateModel.tool])
const visibleToolFields = computed(() => capability.value.toolFields.filter((field) => {
  if (!field.visibleWhen) return true
  return field.visibleWhen(props.templateModel)
}))

const handlePhaseToggle = (phase, enabled) => {
  props.templateModel.phases[phase].enabled = enabled
  templateStore.markDirty()
}

const handleToolFieldChange = (field) => {
  if (field.key === 'benchmark') {
    if (props.templateModel.tool === 'hammerdb') {
      props.templateModel.workloadFamily = toolModel.value[field.key]
    }

    if (props.templateModel.tool === 'swingbench') {
      const swingbenchWorkloads = {
        orderEntry: 'order-entry',
        salesHistory: 'sales-history',
        stressTest: 'stress-test'
      }
      props.templateModel.workloadFamily = swingbenchWorkloads[toolModel.value[field.key]] || props.templateModel.workloadFamily
    }
  }
  templateStore.markDirty()
}
</script>

<style scoped>
.section-card {
  border: 1px solid #1f2937;
  border-radius: 12px;
  background: rgba(15, 23, 42, 0.9);
  padding: 16px;
}

.section-header {
  margin-bottom: 16px;
}

.section-title {
  font-size: 16px;
  color: #f8fafc;
}

.section-subtitle,
.tool-note {
  margin-top: 4px;
  color: #64748b;
  font-size: 12px;
}

.phase-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
  margin-bottom: 18px;
}

.phase-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid #334155;
  background: #111827;
  color: #cbd5e0;
  text-transform: capitalize;
  font-size: 12px;
}

.phase-pill.active {
  border-color: #4299e1;
  background: rgba(66, 153, 225, 0.12);
}

.runtime-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-wide {
  grid-column: span 3;
}

.field-label {
  font-size: 12px;
  color: #94a3b8;
}

.field-input {
  min-height: 40px;
  border-radius: 8px;
  border: 1px solid #334155;
  background: #1e293b;
  color: #e2e8f0;
  padding: 10px 12px;
}

.textarea {
  min-height: 88px;
  resize: vertical;
}

.tool-fields {
  margin-top: 18px;
  padding-top: 18px;
  border-top: 1px solid #1f2937;
}

.tool-header {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 14px;
}

@media (max-width: 980px) {
  .phase-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .runtime-grid {
    grid-template-columns: 1fr;
  }

  .field-wide {
    grid-column: span 1;
  }
}
</style>
