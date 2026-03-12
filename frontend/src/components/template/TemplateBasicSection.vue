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
        <input v-model="templateModel.name" class="field-input" :disabled="readonly" @input="templateStore.markDirty()">
      </label>

      <label class="field">
        <span class="field-label">Database Type</span>
        <select
          class="field-input"
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
      </label>

      <label class="field">
        <span class="field-label">Benchmark Tool</span>
        <select
          class="field-input"
          :disabled="readonly"
          :value="templateModel.tool"
          @change="handleToolChange($event.target.value)"
        >
          <option v-for="(label, value) in templateStore.toolLabels" :key="value" :value="value">{{ label }}</option>
        </select>
      </label>

      <label class="field">
        <span class="field-label">Workload Family</span>
        <select
          class="field-input"
          :disabled="readonly"
          :value="templateModel.workloadFamily"
          @change="handleWorkloadChange($event.target.value)"
        >
          <option v-for="workload in availableWorkloads" :key="workload" :value="workload">
            {{ templateStore.workloadLabels[workload] || workload }}
          </option>
        </select>
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
import { TEMPLATE_CAPABILITIES } from '../../constants/templateCapabilities'
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

const availableDbOptions = computed(() => TEMPLATE_CAPABILITIES[props.templateModel.tool]?.dbFamilies || [])
const availableWorkloads = computed(() => TEMPLATE_CAPABILITIES[props.templateModel.tool]?.workloads || [])

const handleToolChange = (value) => {
  templateStore.updateDraftForTool(value)
}

const handleDbChange = (value) => {
  props.templateModel.dbFamily = value
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
</script>

<style scoped>
.section-card {
  border: 1px solid #1f2937;
  border-radius: 12px;
  background: rgba(15, 23, 42, 0.9);
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
  color: #f8fafc;
}

.section-subtitle {
  margin-top: 4px;
  color: #64748b;
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

.field-input:disabled {
  opacity: 0.72;
  cursor: not-allowed;
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
