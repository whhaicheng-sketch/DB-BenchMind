<template>
  <div class="filter-bar">
    <label class="search-box">
      <span class="field-label">Search</span>
      <input
        :value="filters.search"
        class="field-input"
        type="text"
        placeholder="Search templates, tools, workload, tags"
        @input="$emit('filter-change', { key: 'search', value: $event.target.value })"
      >
    </label>

    <label class="filter-field">
      <span class="field-label">Database Type</span>
      <select
        class="field-select"
        :value="filters.dbFamily"
        @change="$emit('filter-change', { key: 'dbFamily', value: $event.target.value })"
      >
        <option value="">All Databases</option>
        <option v-for="option in dbOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
      </select>
    </label>

    <label class="filter-field">
      <span class="field-label">Benchmark Tool</span>
      <select
        class="field-select"
        :value="filters.tool"
        @change="$emit('filter-change', { key: 'tool', value: $event.target.value })"
      >
        <option value="">All Tools</option>
        <option v-for="option in toolOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
      </select>
    </label>

    <label class="filter-field">
      <span class="field-label">Scope / Tag</span>
      <div class="filter-pair">
        <select
          class="field-select"
          :value="filters.scope"
          @change="$emit('filter-change', { key: 'scope', value: $event.target.value })"
        >
          <option value="">All Scopes</option>
          <option v-for="option in scopeOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
        </select>

        <select
          class="field-select"
          :value="filters.tag"
          @change="$emit('filter-change', { key: 'tag', value: $event.target.value })"
        >
          <option value="">All Tags</option>
          <option v-for="option in tagOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
        </select>
      </div>
    </label>

    <button class="reset-btn" @click="$emit('reset')">Reset Filters</button>
  </div>
</template>

<script setup>
defineProps({
  filters: {
    type: Object,
    required: true
  },
  toolOptions: {
    type: Array,
    default: () => []
  },
  dbOptions: {
    type: Array,
    default: () => []
  },
  tagOptions: {
    type: Array,
    default: () => []
  },
  scopeOptions: {
    type: Array,
    default: () => []
  }
})

defineEmits(['filter-change', 'reset'])
</script>

<style scoped>
.filter-bar {
  display: grid;
  grid-template-columns: minmax(260px, 1.4fr) repeat(3, minmax(180px, 1fr)) auto;
  gap: 12px;
  align-items: end;
  padding: 14px;
  border: 1px solid #2d3748;
  border-radius: 12px;
  background: #111827;
}

.search-box,
.filter-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-size: 12px;
  color: #94a3b8;
}

.field-input,
.field-select {
  width: 100%;
  min-height: 40px;
  padding: 10px 12px;
  border: 1px solid #334155;
  border-radius: 8px;
  background: #1e293b;
  color: #e2e8f0;
}

.field-select {
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  background-image:
    linear-gradient(45deg, transparent 50%, #94a3b8 50%),
    linear-gradient(135deg, #94a3b8 50%, transparent 50%);
  background-position:
    calc(100% - 18px) calc(50% - 3px),
    calc(100% - 12px) calc(50% - 3px);
  background-size: 6px 6px, 6px 6px;
  background-repeat: no-repeat;
  padding-right: 34px;
}

.field-select option {
  background: #1e293b;
  color: #e2e8f0;
}

.field-input:focus,
.field-select:focus {
  border-color: #4299e1;
  outline: none;
}

.filter-pair {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.reset-btn {
  min-height: 40px;
  padding: 0 14px;
  border-radius: 8px;
  border: 1px solid #334155;
  background: #0f172a;
  color: #cbd5e0;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.reset-btn:hover {
  border-color: #4299e1;
  color: #fff;
}

@media (max-width: 1320px) {
  .filter-bar {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 860px) {
  .filter-bar {
    grid-template-columns: 1fr;
  }

  .filter-pair {
    grid-template-columns: 1fr;
  }
}
</style>
