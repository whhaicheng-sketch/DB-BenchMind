<template>
  <div class="impact-toolbar">
    <div class="toolbar-left">
      <!-- Connection Selector -->
      <div class="toolbar-group connection-selector">
        <label class="toolbar-label">Cluster Connection</label>
        <select
          v-model="localConnectionId"
          class="form-select"
          :disabled="isAnalyzing"
          @change="onConnectionChange"
        >
          <option value="" disabled>Select a connection</option>
          <option v-for="conn in connections" :key="conn.id" :value="conn.id">
            {{ conn.name }}
          </option>
        </select>
      </div>

      <!-- Connection Info -->
      <div v-if="selectedConnection" class="toolbar-group connection-info">
        <div class="info-item">
          <span class="info-label">VIP:</span>
          <span class="info-value">{{ selectedConnection.vip || '--' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">Primary:</span>
          <span class="info-value">{{ selectedConnection.primaryNodeIp || '--' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">Secondary:</span>
          <span class="info-value">{{ selectedConnection.secondaryNodeIp || '--' }}</span>
        </div>
      </div>
    </div>

    <div class="toolbar-center">
      <!-- Configuration -->
      <div class="toolbar-group config-group">
        <div class="config-item">
          <label class="toolbar-label">Mode</label>
          <select v-model="localConfig.connectionMode" class="form-select small" :disabled="isAnalyzing">
            <option value="long">Long</option>
            <option value="short">Short</option>
          </select>
        </div>

        <div class="config-item">
          <label class="toolbar-label">Workload</label>
          <select v-model="localConfig.workloadType" class="form-select small" :disabled="isAnalyzing">
            <option value="insert">Insert</option>
            <option value="select">Select</option>
          </select>
        </div>

        <div class="config-item">
          <label class="toolbar-label">Rate (TPS)</label>
          <input
            v-model.number="localConfig.writeRate"
            type="number"
            class="form-input small"
            min="1"
            max="10000"
            :disabled="isAnalyzing"
          />
        </div>

        <div class="config-item checkbox-item">
          <label class="checkbox-label">
            <input
              v-model="localConfig.consistencyCheckEnabled"
              type="checkbox"
              :disabled="isAnalyzing"
            />
            <span>Consistency Check</span>
          </label>
        </div>
      </div>
    </div>

    <div class="toolbar-right">
      <!-- Action Buttons -->
      <div class="toolbar-actions">
        <button
          v-if="!isAnalyzing"
          class="btn btn-start"
          :disabled="!selectedConnection || !localConnectionId"
          @click="startAnalysis"
        >
          <span class="btn-icon">▶</span>
          Start Analysis
        </button>

        <button
          v-else
          class="btn btn-stop"
          @click="stopAnalysis"
        >
          <span class="btn-icon">⏹</span>
          Stop Analysis
        </button>

        <button
          v-if="isAnalyzing || session?.status === 'completed'"
          class="btn btn-reset"
          @click="resetAnalysis"
        >
          <span class="btn-icon">↺</span>
          Reset
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { ConnectionMode, WorkloadType } from '../constants'

const props = defineProps({
  connections: {
    type: Array,
    default: () => []
  },
  selectedConnectionId: {
    type: String,
    default: null
  },
  selectedConnection: {
    type: Object,
    default: null
  },
  config: {
    type: Object,
    default: () => ({
      connectionMode: ConnectionMode.LONG,
      workloadType: WorkloadType.INSERT,
      writeRate: 100,
      consistencyCheckEnabled: true
    })
  },
  isAnalyzing: {
    type: Boolean,
    default: false
  },
  session: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:connectionId', 'update:config', 'start', 'stop', 'reset'])

const localConnectionId = ref(props.selectedConnectionId)
const localConfig = ref({ ...props.config })

// Watch for prop changes
watch(() => props.selectedConnectionId, (newVal) => {
  localConnectionId.value = newVal
})

watch(() => props.config, (newVal) => {
  localConfig.value = { ...newVal }
}, { deep: true })

function onConnectionChange() {
  emit('update:connectionId', localConnectionId.value)
}

function startAnalysis() {
  emit('update:config', { ...localConfig.value })
  emit('start')
}

function stopAnalysis() {
  emit('stop')
}

function resetAnalysis() {
  emit('reset')
}
</script>

<style scoped>
.impact-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  padding: 16px 20px;
  background-color: #1a2332;
  border: 1px solid #2a3a4a;
  border-radius: 8px;
  margin-bottom: 20px;
}

.toolbar-left,
.toolbar-center,
.toolbar-right {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.toolbar-left {
  flex: 0 0 auto;
}

.toolbar-center {
  flex: 1;
  justify-content: center;
}

.toolbar-right {
  flex: 0 0 auto;
}

.toolbar-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.toolbar-label {
  font-size: 11px;
  font-weight: 600;
  color: #a0aec0;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.connection-selector {
  min-width: 240px;
}

.connection-info {
  display: flex;
  flex-direction: row;
  gap: 16px;
  align-items: center;
  padding-top: 22px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.info-label {
  font-size: 12px;
  color: #718096;
}

.info-value {
  font-size: 13px;
  font-weight: 500;
  color: #e2e8f0;
  font-family: 'SF Mono', Monaco, monospace;
}

.config-group {
  display: flex;
  flex-direction: row;
  gap: 16px;
  align-items: flex-end;
}

.config-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.checkbox-item {
  padding-top: 20px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #e2e8f0;
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.form-select,
.form-input {
  padding: 8px 12px;
  background-color: #0f1724;
  border: 1px solid #2a3a4a;
  border-radius: 4px;
  color: #e2e8f0;
  font-size: 13px;
  transition: border-color 0.2s;
}

.form-select:focus,
.form-input:focus {
  outline: none;
  border-color: #4299e1;
}

.form-select.small,
.form-input.small {
  width: 100px;
}

.form-select:disabled,
.form-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.toolbar-actions {
  display: flex;
  gap: 10px;
  padding-top: 16px;
}

.btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-icon {
  font-size: 12px;
}

.btn-start {
  background-color: #48bb78;
  color: white;
}

.btn-start:hover:not(:disabled) {
  background-color: #38a169;
}

.btn-stop {
  background-color: #f56565;
  color: white;
}

.btn-stop:hover {
  background-color: #e53e3e;
}

.btn-reset {
  background-color: #4a5568;
  color: white;
}

.btn-reset:hover {
  background-color: #3a4558;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
