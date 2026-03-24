<template>
  <section class="section-card">
    <div class="section-header">
      <div>
        <h3 class="section-title">Runtime</h3>
      </div>
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
    </div>

    <div class="tool-fields">
      <div class="tool-header">
        <h4>{{ templateStore.toolLabels[templateModel.tool] }}</h4>
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
import { CONCURRENCY_MODE_LABELS } from '../../models/template'
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

const concurrencyLabels = CONCURRENCY_MODE_LABELS

const capability = computed(() => TEMPLATE_CAPABILITIES[props.templateModel.tool] || TEMPLATE_CAPABILITIES.sysbench)
const availableModes = computed(() => capability.value.concurrencyModes)
const toolModel = computed(() => props.templateModel.toolConfig[props.templateModel.tool])
const visibleToolFields = computed(() => capability.value.toolFields.filter((field) => {
  if (!field.visibleWhen) return true
  return field.visibleWhen(props.templateModel)
}))

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
  padding: 10px;
}

.section-header {
  margin-bottom: 8px;
}

.section-title {
  font-size: 14px;
  color: var(--text-primary);
}

.runtime-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px 10px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.field-wide {
  grid-column: span 3;
}

.field-label {
  font-size: 12px;
  color: var(--text-muted);
}

.field-input {
  min-height: 34px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  background: var(--bg-primary);
  color: var(--text-primary);
  padding: 7px 9px;
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
  min-height: 64px;
  resize: vertical;
}

.tool-fields {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--border-color);
}

.tool-header {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 8px;
}

@media (max-width: 980px) {
  .runtime-grid {
    grid-template-columns: 1fr;
  }

  .field-wide {
    grid-column: span 1;
  }
}
</style>
