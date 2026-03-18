<template>
  <div class="filter-bar">
    <label class="search-box">
      <span class="field-label">Search</span>
      <input
        :value="filters.search"
        class="field-input"
        type="text"
        placeholder="Search templates..."
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

    <button class="reset-btn" @click="$emit('reset')">Reset</button>
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
  grid-template-columns: minmax(220px, 1.4fr) repeat(3, minmax(140px, 0.8fr)) auto;
  gap: var(--spacing-sm);
  align-items: end;
  padding: var(--spacing-md);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  background-color: var(--bg-primary);
}

.search-box,
.filter-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.field-label {
  font-size: var(--font-size-xs);
  font-weight: 500;
  color: var(--text-muted);
}

.field-input,
.field-select {
  width: 100%;
  min-height: 32px;
  padding: 5px 10px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background-color: var(--bg-input);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.field-select {
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23909399' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  padding-right: 30px;
  cursor: pointer;
}

.field-select option {
  background-color: var(--bg-primary);
  color: var(--text-primary);
}

.field-input:focus,
.field-select:focus {
  border-color: var(--border-focus);
  outline: none;
}

.field-input::placeholder {
  color: var(--text-placeholder);
}

.filter-pair {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--spacing-xs);
}

.reset-btn {
  min-height: 32px;
  padding: 0 var(--spacing-md);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  background-color: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.reset-btn:hover {
  border-color: var(--primary);
  color: var(--primary);
  background-color: var(--primary-light);
}

@media (max-width: 1200px) {
  .filter-bar {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 900px) {
  .filter-bar {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 680px) {
  .filter-bar {
    grid-template-columns: 1fr;
  }

  .filter-pair {
    grid-template-columns: 1fr;
  }
}
</style>
