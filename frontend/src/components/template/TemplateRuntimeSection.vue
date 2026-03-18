<template>
  <section class="section-card">
    <div class="section-header">
      <div>
        <h3 class="section-title">Workload / Benchmark Settings</h3>
        <p class="section-subtitle">Phase switches, runtime controls and tool-specific benchmark parameters.</p>
      </div>
    </div>

    <div class="phase-grid">
      <label
        v-for="phase in visiblePhaseKeys"
        :key="phase"
        class="phase-pill"
        :class="{
          active: templateModel.phases[phase]?.enabled,
          unavailable: !isPhaseAllowed(phase),
          required: isPhaseRequired(phase)
        }"
      >
        <input
          type="checkbox"
          :checked="templateModel.phases[phase]?.enabled"
          :disabled="readonly || !isPhaseAllowed(phase) || isPhaseRequired(phase)"
          @change="handlePhaseToggle(phase, $event.target.checked)"
        >
        <span>{{ phase }}</span>
        <small v-if="!isPhaseAllowed(phase)">Unavailable</small>
        <small v-else-if="isPhaseRequired(phase)">Required</small>
      </label>
    </div>

    <div class="runtime-grid">
      <label class="field">
        <span class="field-label">Concurrency Mode</span>
        <select
          class="field-input"
          :class="{ invalid: errors.concurrencyMode }"
          :value="templateModel.runtime.concurrency.mode"
          :disabled="readonly"
          @change="templateStore.updateDraftConcurrencyMode($event.target.value)"
        >
          <option v-for="mode in availableModes" :key="mode" :value="mode">{{ concurrencyLabels[mode] }}</option>
        </select>
        <span v-if="errors.concurrencyMode" class="field-error">{{ errors.concurrencyMode }}</span>
      </label>

      <label class="field">
        <span class="field-label">Concurrency Value</span>
        <input
          :value="templateModel.runtime.concurrency.value"
          class="field-input"
          :class="{ invalid: errors.concurrencyValue }"
          type="number"
          min="1"
          :disabled="readonly"
          @input="templateStore.updateDraftConcurrencyValue(Number($event.target.value))"
        >
        <span v-if="errors.concurrencyValue" class="field-error">{{ errors.concurrencyValue }}</span>
      </label>

      <label class="field">
        <span class="field-label">Duration (s)</span>
        <input v-model.number="templateModel.runtime.durationSeconds" class="field-input" :class="{ invalid: errors.durationSeconds }" type="number" min="1" :disabled="readonly" @input="handleRuntimeInput">
        <span v-if="errors.durationSeconds" class="field-error">{{ errors.durationSeconds }}</span>
      </label>

      <label class="field">
        <span class="field-label">Warm-up (s)</span>
        <input v-model.number="templateModel.runtime.warmupSeconds" class="field-input" type="number" min="0" :disabled="readonly" @input="handleRuntimeInput">
      </label>

      <label class="field">
        <span class="field-label">Ramp-up (s)</span>
        <input v-model.number="templateModel.runtime.rampUpSeconds" class="field-input" type="number" min="0" :disabled="readonly" @input="handleRuntimeInput">
      </label>

      <label class="field">
        <span class="field-label">Report Interval (s)</span>
        <input v-model.number="templateModel.runtime.reportIntervalSeconds" class="field-input" type="number" min="1" :disabled="readonly" @input="handleRuntimeInput">
      </label>

      <label class="field">
        <span class="field-label">Percentile</span>
        <input v-model.number="templateModel.runtime.percentile" class="field-input" type="number" min="1" max="100" :disabled="readonly" @input="handleRuntimeInput">
      </label>

      <label class="field">
        <span class="field-label">Iterations</span>
        <input v-model.number="templateModel.runtime.iterations" class="field-input" type="number" min="0" :disabled="readonly" @input="handleRuntimeInput">
      </label>

      <label class="field">
        <span class="field-label">Rate Limit</span>
        <input v-model.number="templateModel.runtime.rateLimit" class="field-input" type="number" min="0" :disabled="readonly" @input="handleRuntimeInput">
      </label>
      <label v-if="errors.phaseCombination || errors.phaseRun" class="field field-wide">
        <span class="field-error">{{ errors.phaseCombination || errors.phaseRun }}</span>
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
              :class="{ invalid: getToolFieldError(field) }"
              :disabled="readonly || isFieldPinned(field)"
              @change="handleToolFieldChange(field)"
            >
              <option v-for="option in getToolFieldOptions(field)" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>

            <textarea
              v-else-if="field.type === 'textarea'"
              v-model="toolModel[field.key]"
              class="field-input textarea"
              :class="{ invalid: getToolFieldError(field) }"
              :disabled="readonly"
              @input="handleRuntimeInput"
            />

            <input
              v-else
              v-model.number="toolModel[field.key]"
              class="field-input"
              :class="{ invalid: getToolFieldError(field) }"
              type="number"
              :min="field.min || 0"
              :disabled="readonly"
              @input="handleRuntimeInput"
            >

            <span v-if="getToolFieldError(field)" class="field-error">{{ getToolFieldError(field) }}</span>
          </label>
        </template>

        <label class="field field-wide">
          <span class="field-label">Runtime Notes</span>
          <textarea
            v-model="templateModel.runtime.notes"
            class="field-input textarea"
            :disabled="readonly"
            placeholder="Add operator guidance or future backend mapping notes"
            @input="handleRuntimeInput"
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
const errors = computed(() => templateStore.validationErrors)

const phaseKeys = PHASE_KEYS
const concurrencyLabels = CONCURRENCY_MODE_LABELS

const capability = computed(() => TEMPLATE_CAPABILITIES[props.templateModel.tool] || TEMPLATE_CAPABILITIES.sysbench)
const availableModes = computed(() => capability.value.concurrencyModes)
const toolModel = computed(() => props.templateModel.toolConfig[props.templateModel.tool])
const visiblePhaseKeys = computed(() => {
  if (templateStore.editorMode === 'standard') {
    return phaseKeys.filter((phase) => capability.value.allowedPhases.includes(phase))
  }

  return phaseKeys
})
const visibleToolFields = computed(() => capability.value.toolFields.filter((field) => {
  if (!field.visibleWhen) return true
  return field.visibleWhen(props.templateModel)
}))

const handlePhaseToggle = (phase, enabled) => {
  templateStore.updateDraftPhase(phase, enabled)
}

const isPhaseAllowed = (phase) => capability.value.allowedPhases.includes(phase)
const isPhaseRequired = (phase) => capability.value.requiredPhases.includes(phase)

const handleRuntimeInput = () => {
  templateStore.markDirty()
  templateStore.validateTemplate(props.templateModel)
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
  handleRuntimeInput()
}

const getToolFieldError = (field) => {
  const keyMap = {
    scriptType: 'sysbenchScriptType',
    benchmark: props.templateModel.tool === 'swingbench' ? 'swingbenchBenchmark' : 'hammerdbBenchmark',
    warehouses: 'hammerdbWarehouses',
    scaleFactor: 'hammerdbScaleFactor'
  }

  return errors.value[keyMap[field.key] || field.key] || ''
}

const getToolFieldOptions = (field) => {
  if (field.key === 'scriptType') {
    const mappedValue = capability.value.workloadFieldMap?.[props.templateModel.workloadFamily]?.scriptType
    return field.options.filter((option) => option.value === mappedValue)
  }

  if (field.key === 'benchmark') {
    const mappedValue = capability.value.workloadFieldMap?.[props.templateModel.workloadFamily]?.benchmark
    return mappedValue ? field.options.filter((option) => option.value === mappedValue) : field.options
  }

  return field.options || []
}

const isFieldPinned = (field) => {
  return ['scriptType', 'benchmark'].includes(field.key) && getToolFieldOptions(field).length === 1
}
</script>

<style scoped>
.section-card {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-primary);
  padding: 16px;
}

.section-header {
  margin-bottom: 16px;
}

.section-title {
  font-size: 16px;
  color: var(--text-primary);
}

.section-subtitle,
.tool-note {
  margin-top: 4px;
  color: var(--text-muted);
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
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  background: var(--bg-secondary);
  color: var(--text-secondary);
  text-transform: capitalize;
  font-size: 12px;
}

.phase-pill.active {
  border-color: var(--primary);
  background: var(--primary-light);
}

.phase-pill.unavailable {
  opacity: 0.45;
  background: var(--bg-secondary);
}

.phase-pill.required {
  border-color: var(--primary);
}

.phase-pill small {
  font-size: 10px;
  color: var(--text-muted);
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
  color: var(--text-muted);
}

.field-input {
  min-height: 40px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  background: var(--bg-primary);
  color: var(--text-primary);
  padding: 10px 12px;
}

select.field-input {
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  background-image:
    linear-gradient(45deg, transparent 50%, var(--primary) 50%),
    linear-gradient(135deg, var(--primary) 50%, transparent 50%);
  background-position:
    calc(100% - 18px) calc(50% - 3px),
    calc(100% - 12px) calc(50% - 3px);
  background-size: 6px 6px, 6px 6px;
  background-repeat: no-repeat;
  padding-right: 34px;
}

select.field-input option {
  background: var(--bg-primary);
  color: var(--text-primary);
}

.field-input.invalid {
  border-color: var(--danger);
  box-shadow: 0 0 0 2px var(--danger-bg);
}

.field-error {
  font-size: 11px;
  color: var(--danger);
}

.textarea {
  min-height: 88px;
  resize: vertical;
}

.tool-fields {
  margin-top: 18px;
  padding-top: 18px;
  border-top: 1px solid var(--border-light);
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
