<script setup>
/**
 * TemplateList.vue
 * Dropdown selector for benchmark templates.
 * Displays template name, tool, and supported databases.
 */
import { ref, computed, onMounted, watch } from 'vue'
import { useTemplateStore } from '../../stores/template'
import { useConnectionStore } from '../../stores/connection'

// Props
const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  disabled: {
    type: Boolean,
    default: false
  },
  // Filter by database type (optional)
  dbType: {
    type: String,
    default: ''
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'template-selected'])

// Stores
const templateStore = useTemplateStore()
const connectionStore = useConnectionStore()

// Local state
const isDropdownOpen = ref(false)

// Computed
const selectedTemplate = computed(() => {
  if (!props.modelValue) return null
  return templateStore.templates.find(t => t.id === props.modelValue)
})

const displayText = computed(() => {
  if (selectedTemplate.value) {
    return `${selectedTemplate.value.name} (${templateStore.toolLabels[selectedTemplate.value.tool] || selectedTemplate.value.tool})`
  }
  return 'Select template...'
})

// Filter templates by database type
const filteredTemplates = computed(() => {
  if (!props.dbType) return templateStore.templates
  return templateStore.templates.filter(t =>
    t.database_types?.includes(props.dbType)
  )
})

// Group filtered templates by tool
const groupedTemplates = computed(() => {
  const grouped = {}
  for (const tmpl of filteredTemplates.value) {
    if (!grouped[tmpl.tool]) {
      grouped[tmpl.tool] = []
    }
    grouped[tmpl.tool].push(tmpl)
  }
  return grouped
})

// Methods
const toggleDropdown = () => {
  if (!props.disabled) {
    isDropdownOpen.value = !isDropdownOpen.value
  }
}

const selectTemplate = async (template) => {
  emit('update:modelValue', template.id)
  emit('template-selected', template)
  isDropdownOpen.value = false

  // Load template details
  await templateStore.selectTemplate(template.id)
}

const handleClickOutside = (event) => {
  if (!event.target.closest('.template-list')) {
    isDropdownOpen.value = false
  }
}

// Get database type badge color
const getDbTypeClass = (dbType) => {
  const classes = {
    mysql: 'mysql',
    postgresql: 'postgresql',
    oracle: 'oracle',
    sqlserver: 'sqlserver'
  }
  return classes[dbType] || ''
}

// Lifecycle
onMounted(async () => {
  await templateStore.fetchTemplates()
  document.addEventListener('click', handleClickOutside)
})

// Watch dbType to filter templates
watch(() => props.dbType, async (newType) => {
  if (newType) {
    // Templates are filtered locally, no need to refetch
    // But if current selection doesn't support new type, clear it
    if (selectedTemplate.value && !selectedTemplate.value.database_types?.includes(newType)) {
      emit('update:modelValue', '')
      templateStore.clearSelection()
    }
  }
})
</script>

<template>
  <div class="template-list" :class="{ disabled: disabled }">
    <div
      class="dropdown-trigger"
      @click="toggleDropdown"
      :class="{ open: isDropdownOpen }"
    >
      <span class="trigger-text" :class="{ placeholder: !selectedTemplate }">
        {{ displayText }}
      </span>
      <span class="trigger-arrow" :class="{ rotated: isDropdownOpen }">&#9660;</span>
    </div>

    <div v-if="isDropdownOpen" class="dropdown-menu">
      <template v-for="(templates, tool) in groupedTemplates" :key="tool">
        <div class="dropdown-group-header">
          {{ templateStore.toolLabels[tool] || tool }}
        </div>
        <div
          v-for="tmpl in templates"
          :key="tmpl.id"
          class="dropdown-item"
          :class="{ selected: tmpl.id === modelValue }"
          @click="selectTemplate(tmpl)"
        >
          <div class="item-header">
            <span class="item-name">{{ tmpl.name }}</span>
          </div>
          <div class="item-description">{{ tmpl.description }}</div>
          <div class="item-db-types">
            <span
              v-for="dbType in tmpl.database_types"
              :key="dbType"
              class="db-type-badge"
              :class="getDbTypeClass(dbType)"
            >
              {{ connectionStore.typeLabels[dbType] || dbType }}
            </span>
          </div>
        </div>
      </template>

      <div v-if="filteredTemplates.length === 0" class="dropdown-empty">
        <template v-if="dbType">
          No templates available for {{ connectionStore.typeLabels[dbType] || dbType }}
        </template>
        <template v-else>
          No templates available
        </template>
      </div>
    </div>

    <!-- Loading indicator -->
    <div v-if="templateStore.loading" class="loading-indicator">
      Loading...
    </div>

    <!-- Error display -->
    <div v-if="templateStore.error" class="error-message">
      {{ templateStore.error }}
    </div>
  </div>
</template>

<style scoped>
.template-list {
  position: relative;
  width: 100%;
}

.template-list.disabled {
  opacity: 0.6;
  pointer-events: none;
}

.dropdown-trigger {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background-color: #1e2a3a;
  border: 1px solid #3a4a5a;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.dropdown-trigger:hover {
  border-color: #4299e1;
}

.dropdown-trigger.open {
  border-color: #4299e1;
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
}

.trigger-text {
  flex: 1;
  color: #e2e8f0;
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.trigger-text.placeholder {
  color: #718096;
}

.trigger-arrow {
  color: #718096;
  font-size: 10px;
  transition: transform 0.2s ease;
  margin-left: 8px;
}

.trigger-arrow.rotated {
  transform: rotate(180deg);
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background-color: #1e2a3a;
  border: 1px solid #4299e1;
  border-top: none;
  border-radius: 0 0 6px 6px;
  max-height: 300px;
  overflow-y: auto;
  z-index: 100;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.dropdown-group-header {
  padding: 8px 12px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  color: #718096;
  background-color: #1a2433;
  border-bottom: 1px solid #2a3a4a;
}

.dropdown-item {
  padding: 10px 12px;
  cursor: pointer;
  transition: background-color 0.2s ease;
  border-bottom: 1px solid #2a3a4a;
}

.dropdown-item:last-child {
  border-bottom: none;
}

.dropdown-item:hover {
  background-color: #2a3a4a;
}

.dropdown-item.selected {
  background-color: #2d4a5a;
}

.item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.item-name {
  color: #e2e8f0;
  font-size: 14px;
  font-weight: 500;
}

.item-description {
  font-size: 12px;
  color: #718096;
  margin-bottom: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-db-types {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.db-type-badge {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 3px;
  text-transform: uppercase;
  font-weight: 600;
}

.db-type-badge.mysql {
  background-color: #00758f33;
  color: #00758f;
}

.db-type-badge.postgresql {
  background-color: #33679133;
  color: #699eca;
}

.db-type-badge.oracle {
  background-color: #f8000033;
  color: #ff6b6b;
}

.db-type-badge.sqlserver {
  background-color: #cc292733;
  color: #cc2927;
}

.dropdown-empty {
  padding: 20px;
  text-align: center;
  color: #718096;
  font-size: 14px;
}

.loading-indicator {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  padding: 8px;
  text-align: center;
  color: #4299e1;
  font-size: 12px;
  background-color: #1e2a3a;
  border: 1px solid #3a4a5a;
  border-top: none;
}

.error-message {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  padding: 8px;
  color: #fc8181;
  font-size: 12px;
  background-color: #1e2a3a;
  border: 1px solid #e53e3e;
  border-top: none;
  border-radius: 0 0 6px 6px;
}
</style>
