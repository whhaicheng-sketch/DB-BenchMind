<script setup>
/**
 * Sidebar.vue
 * Left sidebar control panel for DB-BenchMind.
 * Contains connection selector, template selector, parameters, controls, and logs.
 */
import { ref, computed, onMounted, watch, onUnmounted } from 'vue'
import { useConnectionStore } from '../../stores/connection'
import { useTemplateStore } from '../../stores/template'
import { useBenchmarkStore } from '../../stores/benchmark'
import ConnectionList from '../connection/ConnectionList.vue'
import ConnectionForm from '../connection/ConnectionForm.vue'
import TemplateList from '../template/TemplateList.vue'
import LogPanel from '../benchmark/LogPanel.vue'

// Stores
const connectionStore = useConnectionStore()
const templateStore = useTemplateStore()
const benchmarkStore = useBenchmarkStore()

// Local state
const selectedConnectionId = ref('')
const selectedTemplateId = ref('')
const threads = ref('4')
const duration = ref('60')
const rampup = ref('0')

// Modal state
const showConnectionModal = ref(false)
const connectionModalMode = ref('create') // 'create' or 'edit'

// Computed
const selectedConnection = computed(() => {
  if (!selectedConnectionId.value) return null
  return connectionStore.connections.find(c => c.id === selectedConnectionId.value)
})

const selectedDbType = computed(() => {
  return selectedConnection.value?.type || ''
})

// Get current parameter values
const paramValues = computed(() => templateStore.paramValues)
const templateParams = computed(() => templateStore.templateParams)

// Benchmark state from store
const isRunning = computed(() => benchmarkStore.isRunning || benchmarkStore.isPreparing || benchmarkStore.isCleaning)
const logLines = computed(() => benchmarkStore.logLines)
const currentState = computed(() => benchmarkStore.currentState)

// Handlers
const handleConnectionSelected = (connection) => {
  console.log('Selected connection:', connection)
}

const handleTemplateSelected = (template) => {
  console.log('Selected template:', template)
}

// Watch connection changes to reset template if incompatible
watch(selectedDbType, (newType) => {
  if (selectedTemplateId.value && templateStore.selectedTemplate) {
    if (!templateStore.selectedTemplate.database_types?.includes(newType)) {
      selectedTemplateId.value = ''
      templateStore.clearSelection()
    }
  }
})

// Build parameters for benchmark
const buildBenchmarkParams = () => {
  const params = { ...paramValues.value }
  // Add common parameters
  if (threads.value) params.threads = parseInt(threads.value)
  if (duration.value) params.duration = parseInt(duration.value)
  if (rampup.value) params.warmup_time = parseInt(rampup.value)
  return params
}

const handlePrepare = async () => {
  if (!selectedConnectionId.value || !selectedTemplateId.value) return
  benchmarkStore.clearLogs()
  await benchmarkStore.prepareOnly(
    selectedConnectionId.value,
    selectedTemplateId.value,
    buildBenchmarkParams()
  )
}

const handleRun = async () => {
  if (!selectedConnectionId.value || !selectedTemplateId.value) return
  benchmarkStore.clearLogs()
  await benchmarkStore.runBenchmark(
    selectedConnectionId.value,
    selectedTemplateId.value,
    buildBenchmarkParams(),
    { warmupTime: parseInt(rampup.value) || 0 }
  )
}

const handleStop = async () => {
  await benchmarkStore.stopBenchmark(false)
}

const handleCleanup = async () => {
  if (!selectedConnectionId.value || !selectedTemplateId.value) return
  benchmarkStore.clearLogs()
  await benchmarkStore.cleanupOnly(
    selectedConnectionId.value,
    selectedTemplateId.value,
    buildBenchmarkParams()
  )
}

// Connection modal handlers
const openCreateConnectionModal = () => {
  connectionModalMode.value = 'create'
  showConnectionModal.value = true
}

const openEditConnectionModal = () => {
  if (!selectedConnectionId.value) return
  connectionModalMode.value = 'edit'
  showConnectionModal.value = true
}

const closeConnectionModal = () => {
  showConnectionModal.value = false
}

const handleConnectionSaved = (conn) => {
  closeConnectionModal()
  // Select the new/updated connection
  if (conn && conn.id) {
    selectedConnectionId.value = conn.id
  }
}

const handleDeleteConnection = async () => {
  if (!selectedConnectionId.value) return
  if (confirm('Are you sure you want to delete this connection?')) {
    await connectionStore.deleteConnection(selectedConnectionId.value)
    selectedConnectionId.value = ''
  }
}

const handleTestConnection = async () => {
  if (!selectedConnectionId.value) return
  await connectionStore.testConnectionById(selectedConnectionId.value)
}

// Lifecycle
onMounted(async () => {
  await connectionStore.fetchConnections()
  await templateStore.fetchTemplates()
  // Subscribe to benchmark events
  benchmarkStore.subscribeToEvents()
})

onUnmounted(() => {
  benchmarkStore.stopStatusPolling()
  // Unsubscribe from benchmark events
  benchmarkStore.unsubscribeFromEvents()
})
</script>

<template>
  <div class="sidebar">
    <!-- Connection Section -->
    <div class="section">
      <div class="section-header">
        <h3 class="section-title">Connection</h3>
        <div class="section-actions">
          <button class="action-btn" @click="openCreateConnectionModal" title="Add Connection">
            +
          </button>
        </div>
      </div>
      <ConnectionList
        v-model="selectedConnectionId"
        :disabled="isRunning"
        @connection-selected="handleConnectionSelected"
      />
      <div v-if="selectedConnection" class="connection-actions">
        <button
          class="btn-small"
          @click="handleTestConnection"
          :disabled="connectionStore.loading"
        >
          Test
        </button>
        <button
          class="btn-small"
          @click="openEditConnectionModal"
          :disabled="isRunning"
        >
          Edit
        </button>
        <button
          class="btn-small btn-danger"
          @click="handleDeleteConnection"
          :disabled="isRunning"
        >
          Delete
        </button>
      </div>
      <!-- Test result display -->
      <div v-if="connectionStore.testResult" class="test-result" :class="connectionStore.testResult.success ? 'success' : 'error'">
        <div class="result-header">
          <span class="result-icon">{{ connectionStore.testResult.success ? '✓' : '✗' }}</span>
          <span class="result-status">{{ connectionStore.testResult.success ? 'Success' : 'Failed' }}</span>
        </div>
        <div v-if="connectionStore.testResult.success" class="result-details">
          <div>Latency: {{ connectionStore.testResult.latency_ms }}ms</div>
          <div v-if="connectionStore.testResult.database_version">Version: {{ connectionStore.testResult.database_version }}</div>
        </div>
        <div v-if="!connectionStore.testResult.success" class="result-error">
          {{ connectionStore.testResult.error }}
        </div>
      </div>
    </div>

    <!-- Template Section -->
    <div class="section">
      <h3 class="section-title">Template</h3>
      <TemplateList
        v-model="selectedTemplateId"
        :disabled="isRunning"
        :db-type="selectedDbType"
        @template-selected="handleTemplateSelected"
      />
      <!-- Template description -->
      <div v-if="templateStore.selectedTemplate" class="template-info">
        <div class="template-description">{{ templateStore.selectedTemplate.description }}</div>
      </div>
    </div>

    <!-- Parameters Section -->
    <div class="section">
      <h3 class="section-title">Parameters</h3>
      <!-- Dynamic template parameters -->
      <template v-if="templateParams.length > 0">
        <div v-for="param in templateParams" :key="param.name" class="form-group">
          <label>{{ param.label }}:</label>
          <!-- Select for options -->
          <select
            v-if="param.options && param.options.length > 0"
            v-model="paramValues[param.name]"
            class="text-input"
            :disabled="isRunning"
            @change="templateStore.setParamValue(param.name, paramValues[param.name])"
          >
            <option v-for="opt in param.options" :key="opt" :value="opt">{{ opt }}</option>
          </select>
          <!-- Number input -->
          <input
            v-else-if="param.type === 'int' || param.type === 'number'"
            v-model.number="paramValues[param.name]"
            type="number"
            class="text-input"
            :min="param.min"
            :max="param.max"
            :disabled="isRunning"
            @change="templateStore.setParamValue(param.name, paramValues[param.name])"
          />
          <!-- Text input for others -->
          <input
            v-else
            v-model="paramValues[param.name]"
            type="text"
            class="text-input"
            :disabled="isRunning"
            @change="templateStore.setParamValue(param.name, paramValues[param.name])"
          />
        </div>
      </template>
      <!-- Default parameters if no template selected -->
      <template v-else>
        <div class="form-group">
          <label>Threads:</label>
          <input
            v-model="threads"
            type="number"
            class="text-input"
            :disabled="isRunning"
          />
        </div>
        <div class="form-group">
          <label>Duration (s):</label>
          <input
            v-model="duration"
            type="number"
            class="text-input"
            :disabled="isRunning"
          />
        </div>
        <div class="form-group">
          <label>Warmup (s):</label>
          <input
            v-model="rampup"
            type="number"
            class="text-input"
            :disabled="isRunning"
          />
        </div>
      </template>
    </div>

    <!-- Control Panel -->
    <div class="section control-panel">
      <h3 class="section-title">Control</h3>
      <div class="button-group">
        <button
          class="btn btn-prepare"
          @click="handlePrepare"
          :disabled="isRunning || !selectedConnectionId || !selectedTemplateId"
        >
          Prepare
        </button>
        <button
          class="btn btn-run"
          @click="handleRun"
          :disabled="isRunning || !selectedConnectionId || !selectedTemplateId"
        >
          Run
        </button>
        <button
          class="btn btn-stop"
          @click="handleStop"
          :disabled="!isRunning"
        >
          Stop
        </button>
        <button
          class="btn btn-cleanup"
          @click="handleCleanup"
          :disabled="isRunning || !selectedConnectionId || !selectedTemplateId"
        >
          Cleanup
        </button>
      </div>
    </div>

    <!-- Status Section -->
    <div class="section status-section">
      <h3 class="section-title">Status</h3>
      <div class="status-row">
        <span class="status-label">Status:</span>
        <span :class="['status-value', currentState]">
          {{ benchmarkStore.stateLabels[currentState] || 'Idle' }}
        </span>
      </div>
      <!-- Show result summary if completed -->
      <div v-if="benchmarkStore.result" class="result-summary">
        <div class="result-item">
          <span class="result-label">TPS:</span>
          <span class="result-value">{{ benchmarkStore.result.tps?.toFixed(2) || 'N/A' }}</span>
        </div>
        <div class="result-item">
          <span class="result-label">TPM:</span>
          <span class="result-value">{{ benchmarkStore.result.tpm?.toFixed(2) || 'N/A' }}</span>
        </div>
        <div class="result-item">
          <span class="result-label">Latency:</span>
          <span class="result-value">{{ benchmarkStore.result.latency_avg_ms?.toFixed(2) || 'N/A' }} ms</span>
        </div>
      </div>
    </div>

    <!-- Log Section -->
    <div class="section log-section">
      <LogPanel
        max-height="200px"
        :auto-scroll="true"
        :max-lines="200"
      />
    </div>

    <!-- Connection Modal -->
    <div v-if="showConnectionModal" class="modal-overlay" @click.self="closeConnectionModal">
      <div class="modal-content">
        <ConnectionForm
          :mode="connectionModalMode"
          :connection-id="connectionModalMode === 'edit' ? selectedConnectionId : null"
          @saved="handleConnectionSaved"
          @cancelled="closeConnectionModal"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.sidebar {
  width: 300px;
  min-width: 300px;
  height: 100%;
  background-color: #1e2a3a;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
}

.section {
  background-color: #2a3a4a;
  border-radius: 8px;
  padding: 12px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #a0aec0;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin: 0;
}

.section-actions {
  display: flex;
  gap: 4px;
}

.action-btn {
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 4px;
  background-color: #3a4a5a;
  color: #a0aec0;
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.action-btn:hover {
  background-color: #4a5a6a;
  color: #e2e8f0;
}

.connection-actions {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.btn-small {
  padding: 4px 12px;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  background-color: #3a4a5a;
  color: #a0aec0;
  transition: all 0.2s ease;
}

.btn-small:hover:not(:disabled) {
  background-color: #4a5a6a;
  color: #e2e8f0;
}

.btn-small:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-small.btn-danger {
  background-color: #742a2a;
  color: #fc8181;
}

.btn-small.btn-danger:hover:not(:disabled) {
  background-color: #9a2a2a;
}

.test-result {
  margin-top: 12px;
  padding: 8px;
  border-radius: 6px;
  font-size: 12px;
}

.test-result.success {
  background-color: #1a3a2a;
  border: 1px solid #2d5a3a;
}

.test-result.error {
  background-color: #3a1a1a;
  border: 1px solid #5a2d2d;
}

.result-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}

.result-icon {
  font-size: 14px;
}

.test-result.success .result-icon {
  color: #68d391;
}

.test-result.error .result-icon {
  color: #fc8181;
}

.result-status {
  font-weight: 600;
}

.test-result.success .result-status {
  color: #68d391;
}

.test-result.error .result-status {
  color: #fc8181;
}

.result-details {
  color: #a0aec0;
  font-size: 11px;
}

.result-details div {
  margin-top: 2px;
}

.result-error {
  margin-top: 4px;
  color: #fc8181;
  font-size: 11px;
  word-break: break-all;
}

.select-input,
.text-input {
  width: 100%;
  padding: 8px 12px;
  background-color: #1e2a3a;
  border: 1px solid #3a4a5a;
  border-radius: 6px;
  color: #e2e8f0;
  font-size: 14px;
}

.select-input:focus,
.text-input:focus {
  outline: none;
  border-color: #4299e1;
}

.select-input:disabled,
.text-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.form-group {
  margin-bottom: 12px;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  font-size: 12px;
  color: #a0aec0;
  margin-bottom: 4px;
}

.button-group {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.btn {
  padding: 10px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-prepare {
  background-color: #4a5568;
  color: #fff;
}

.btn-prepare:hover:not(:disabled) {
  background-color: #5a6578;
}

.btn-run {
  background-color: #48bb78;
  color: #fff;
}

.btn-run:hover:not(:disabled) {
  background-color: #38a169;
}

.btn-stop {
  background-color: #e53e3e;
  color: #fff;
}

.btn-stop:hover:not(:disabled) {
  background-color: #c53030;
}

.btn-cleanup {
  background-color: #4a5568;
  color: #fff;
}

.btn-cleanup:hover:not(:disabled) {
  background-color: #5a6578;
}

.template-info {
  margin-top: 8px;
}

.template-description {
  font-size: 12px;
  color: #718096;
  font-style: italic;
}

.status-section {
  flex-shrink: 0;
}

.status-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-label {
  color: #a0aec0;
  font-size: 14px;
}

.status-value {
  font-weight: 600;
  font-size: 14px;
}

.status-value.idle {
  color: #a0aec0;
}

.status-value.running {
  color: #48bb78;
}

.status-value.preparing {
  color: #ecc94b;
}

.status-value.warming_up {
  color: #f6ad55;
}

.status-value.completed {
  color: #68d391;
}

.status-value.failed {
  color: #fc8181;
}

.status-value.cancelled,
.status-value.timeout,
.status-value.force_stopped {
  color: #f6ad55;
}

.result-summary {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #3a4a5a;
}

.result-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
}

.result-item .result-label {
  color: #718096;
  font-size: 12px;
}

.result-item .result-value {
  color: #e2e8f0;
  font-size: 14px;
  font-weight: 600;
}

.log-section {
  flex: 1;
  min-height: 150px;
  display: flex;
  flex-direction: column;
}

.log-section :deep(.log-panel) {
  flex: 1;
}

/* Modal styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background-color: #2a3a4a;
  border-radius: 8px;
  padding: 20px;
  width: 400px;
  max-width: 90vw;
  max-height: 90vh;
  overflow-y: auto;
}
</style>
