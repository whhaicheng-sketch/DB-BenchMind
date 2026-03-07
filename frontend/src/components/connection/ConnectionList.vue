<script setup>
/**
 * ConnectionList.vue
 * Dropdown selector for database connections.
 * Displays connection name, type, and host info.
 */
import { ref, computed, onMounted, watch } from 'vue'
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
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'connection-selected'])

// Store
const connectionStore = useConnectionStore()

// Local state
const isDropdownOpen = ref(false)

// Computed
const selectedConnection = computed(() => {
  if (!props.modelValue) return null
  return connectionStore.connections.find(c => c.id === props.modelValue)
})

const displayText = computed(() => {
  if (selectedConnection.value) {
    return `${selectedConnection.value.name} (${connectionStore.typeLabels[selectedConnection.value.type] || selectedConnection.value.type})`
  }
  return 'Select connection...'
})

// Methods
const toggleDropdown = () => {
  if (!props.disabled) {
    isDropdownOpen.value = !isDropdownOpen.value
  }
}

const selectConnection = (connection) => {
  emit('update:modelValue', connection.id)
  emit('connection-selected', connection)
  isDropdownOpen.value = false
}

const handleClickOutside = (event) => {
  if (!event.target.closest('.connection-list')) {
    isDropdownOpen.value = false
  }
}

// Lifecycle
onMounted(async () => {
  await connectionStore.fetchConnections()
  document.addEventListener('click', handleClickOutside)
})

// Watch for modelValue changes
watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    const conn = connectionStore.connections.find(c => c.id === newVal)
    if (conn) {
      emit('connection-selected', conn)
    }
  }
})
</script>

<template>
  <div class="connection-list" :class="{ disabled: disabled }">
    <div
      class="dropdown-trigger"
      @click="toggleDropdown"
      :class="{ open: isDropdownOpen }"
    >
      <span class="trigger-text" :class="{ placeholder: !selectedConnection }">
        {{ displayText }}
      </span>
      <span class="trigger-arrow" :class="{ rotated: isDropdownOpen }">▼</span>
    </div>

    <div v-if="isDropdownOpen" class="dropdown-menu">
      <div
        v-for="conn in connectionStore.connections"
        :key="conn.id"
        class="dropdown-item"
        :class="{ selected: conn.id === modelValue }"
        @click="selectConnection(conn)"
      >
        <div class="item-header">
          <span class="item-name">{{ conn.name }}</span>
          <span class="item-type" :class="conn.type">
            {{ connectionStore.typeLabels[conn.type] || conn.type }}
          </span>
        </div>
        <div class="item-info">
          {{ conn.username }}@{{ conn.host }}:{{ conn.port }}
          <span v-if="conn.database">/{{ conn.database }}</span>
        </div>
      </div>

      <div v-if="connectionStore.connections.length === 0" class="dropdown-empty">
        No connections available
      </div>
    </div>

    <!-- Loading indicator -->
    <div v-if="connectionStore.loading" class="loading-indicator">
      Loading...
    </div>

    <!-- Error display -->
    <div v-if="connectionStore.error" class="error-message">
      {{ connectionStore.error }}
    </div>
  </div>
</template>

<style scoped>
.connection-list {
  position: relative;
  width: 100%;
}

.connection-list.disabled {
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
  max-height: 250px;
  overflow-y: auto;
  z-index: 100;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.dropdown-item {
  padding: 10px 12px;
  cursor: pointer;
  transition: background-color 0.2s ease;
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

.item-type {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  text-transform: uppercase;
  font-weight: 600;
}

.item-type.mysql {
  background-color: #00758f33;
  color: #00758f;
}

.item-type.postgresql {
  background-color: #33679133;
  color: #699eca;
}

.item-type.oracle {
  background-color: #f8000033;
  color: #ff6b6b;
}

.item-type.sqlserver {
  background-color: #cc292733;
  color: #cc2927;
}

.item-info {
  font-size: 12px;
  color: #718096;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
