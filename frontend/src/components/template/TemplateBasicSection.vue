<template>
  <section class="section-card">
    <div class="section-header">
      <div>
        <h3 class="section-title">Basic Information</h3>
        <p class="section-subtitle">Core identity, scope and applicability for the selected template.</p>
      </div>
    </div>

    <div class="form-grid">
      <label class="field">
        <span class="field-label">Template Name</span>
        <input v-model="templateModel.name" class="field-input" :class="{ invalid: errors.name }" :disabled="readonly" @input="handleFieldInput('name')">
        <span v-if="errors.name" class="field-error">{{ errors.name }}</span>
      </label>

      <label class="field">
        <span class="field-label">Database Type</span>
        <select
          class="field-input"
          :class="{ invalid: errors.dbFamily }"
          :disabled="readonly"
          :value="templateModel.dbFamily"
          @change="handleDbChange($event.target.value)"
        >
          <option
            v-for="option in availableDbOptions"
            :key="option"
            :value="option"
          >
            {{ templateStore.dbFamilyLabels[option] || option }}
          </option>
        </select>
        <span v-if="errors.dbFamily" class="field-error">{{ errors.dbFamily }}</span>
      </label>

      <label class="field">
        <span class="field-label">Benchmark Tool</span>
        <select
          class="field-input"
          :class="{ invalid: errors.tool }"
          :disabled="readonly"
          :value="templateModel.tool"
          @change="handleToolChange($event.target.value)"
        >
          <option v-for="tool in availableToolOptions" :key="tool" :value="tool">{{ templateStore.toolLabels[tool] || tool }}</option>
        </select>
        <span v-if="errors.tool" class="field-error">{{ errors.tool }}</span>
      </label>

      <label class="field">
        <span class="field-label">Workload Family</span>
        <select
          class="field-input"
          :class="{ invalid: errors.workloadFamily }"
          :disabled="readonly"
          :value="templateModel.workloadFamily"
          @change="handleWorkloadChange($event.target.value)"
        >
          <option v-for="workload in availableWorkloads" :key="workload" :value="workload">
            {{ templateStore.workloadLabels[workload] || workload }}
          </option>
        </select>
        <span v-if="errors.workloadFamily" class="field-error">{{ errors.workloadFamily }}</span>
      </label>

      <label class="field field-wide">
        <span class="field-label">Description</span>
        <textarea
          v-model="templateModel.description"
          class="field-input textarea"
          :disabled="readonly"
          @input="templateStore.markDirty()"
        />
      </label>

      <label class="field">
        <span class="field-label">Tags</span>
        <input
          :value="(templateModel.tags || []).join(', ')"
          class="field-input"
          :disabled="readonly"
          placeholder="baseline, smoke, custom"
          @input="handleTagsInput($event.target.value)"
        >
      </label>

      <label class="field">
        <span class="field-label">Scope</span>
        <input class="field-input" :value="templateStore.scopeLabels[templateModel.scope]" disabled>
      </label>

      <label class="field">
        <span class="field-label">Status</span>
        <select v-model="templateModel.status" class="field-input" :disabled="readonly" @change="templateStore.markDirty()">
          <option value="draft">Draft</option>
          <option value="ready">Ready</option>
          <option value="deprecated">Deprecated</option>
        </select>
      </label>

      <label class="field">
        <span class="field-label">Last Updated</span>
        <input class="field-input" :value="formatTemplateDate(templateModel.updatedAt)" disabled>
      </label>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import { DB_OPTIONS, TEMPLATE_CAPABILITIES, getToolsForDbFamily } from '../../constants/templateCapabilities'
import { formatTemplateDate } from '../../models/template'
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

const availableDbOptions = computed(() => DB_OPTIONS.map((option) => option.value))
const availableToolOptions = computed(() => getToolsForDbFamily(props.templateModel.dbFamily))
const availableWorkloads = computed(() => TEMPLATE_CAPABILITIES[props.templateModel.tool]?.workloads || [])

const handleToolChange = (value) => {
  templateStore.updateDraftForTool(value)
}

const handleDbChange = (value) => {
  templateStore.updateDraftDbFamily(value)
}

const handleWorkloadChange = (value) => {
  props.templateModel.workloadFamily = value
  templateStore.updateDraftWorkload(value)
}

const handleTagsInput = (value) => {
  props.templateModel.tags = value.split(',').map((tag) => tag.trim()).filter(Boolean)
  templateStore.markDirty()
}

const handleFieldInput = () => {
  templateStore.markDirty()
  templateStore.validateTemplate(props.templateModel)
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
  display: flex;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.section-title {
  font-size: 16px;
  color: var(--text-primary);
}

.section-subtitle {
  margin-top: 4px;
  color: var(--text-muted);
  font-size: 12px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-wide {
  grid-column: span 2;
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

.field-input:disabled {
  opacity: 0.72;
  cursor: not-allowed;
  background-color: var(--bg-secondary);
}

.field-error {
  font-size: 11px;
  color: var(--danger);
}

.textarea {
  min-height: 88px;
  resize: vertical;
}

@media (max-width: 920px) {
  .form-grid {
    grid-template-columns: 1fr;
  }

  .field-wide {
    grid-column: span 1;
  }
}
</style>
