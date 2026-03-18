<script setup>
/**
 * Sidebar.vue
 * Left sidebar control panel for DB-BenchMind.
 * Navicat-style light theme.
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
    <div class="sidebar-section">
      <div class="section-header">
        <h3 class="section-title">Connection</h3>
        <button class="btn-add" @click="openCreateConnectionModal" title="Add Connection">+</button>
      </div>
      <ConnectionList
        v-model="selectedConnectionId"
        :disabled="isRunning"
        @connection-selected="handleConnectionSelected"
      />
      <div v-if="selectedConnection" class="connection-actions">
        <button class="btn btn-small" @click="handleTestConnection" :disabled="connectionStore.loading">Test</button>
        <button class="btn btn-small" @click="openEditConnectionModal" :disabled="isRunning">Edit</button>
        <button class="btn btn-small btn-danger" @click="handleDeleteConnection" :disabled="isRunning">Delete</button>
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
    <div class="sidebar-section">
      <div class="section-header">
        <h3 class="section-title">Template</h3>
      </div>
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
    <div class="sidebar-section">
      <div class="section-header">
        <h3 class="section-title">Parameters</h3>
      </div>
      <!-- Dynamic template parameters -->
      <template v-if="templateParams.length > 0">
        <div v-for="param in templateParams" :key="param.name" class="form-group">
          <label>{{ param.label }}:</label>
          <!-- Select for options -->
          <select
            v-if="param.options && param.options.length > 0"
            v-model="paramValues[param.name]"
            class="select"
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
            class="input"
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
            class="input"
            :disabled="isRunning"
            @change="templateStore.setParamValue(param.name, paramValues[param.name])"
          />
        </div>
      </template>
      <!-- Default parameters if no template selected -->
      <template v-else>
        <div class="form-group">
          <label>Threads:</label>
          <input v-model="threads" type="number" class="input" :disabled="isRunning" />
        </div>
        <div class="form-group">
          <label>Duration (s):</label>
          <input v-model="duration" type="number" class="input" :disabled="isRunning" />
        </div>
        <div class="form-group">
          <label>Warmup (s):</label>
          <input v-model="rampup" type="number" class="input" :disabled="isRunning" />
        </div>
      </template>
    </div>

    <!-- Control Panel -->
    <div class="sidebar-section control-panel">
      <div class="section-header">
        <h3 class="section-title">Control</h3>
      </div>
      <div class="button-grid">
        <button class="btn" @click="handlePrepare" :disabled="isRunning || !selectedConnectionId || !selectedTemplateId">Prepare</button>
        <button class="btn btn-success" @click="handleRun" :disabled="isRunning || !selectedConnectionId || !selectedTemplateId">Run</button>
        <button class="btn btn-danger" @click="handleStop" :disabled="!isRunning">Stop</button>
        <button class="btn" @click="handleCleanup" :disabled="isRunning || !selectedConnectionId || !selectedTemplateId">Cleanup</button>
      </div>
    </div>

    <!-- Status Section -->
    <div class="sidebar-section status-section">
      <div class="section-header">
        <h3 class="section-title">Status</h3>
      </div>
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
    <div class="sidebar-section log-section">
      <LogPanel max-height="200px" :auto-scroll="true" :max-lines="200" />
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
  width: 280px;
  min-width: 280px;
  height: 100%;
  background-color: var(--bg-secondary);
  padding: var(--spacing-md);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  overflow-y: auto;
  border-right: 1px solid var(--border-color);
}

/* Section */
.sidebar-section {
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-sm);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-sm);
}

.section-title {
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin: 0;
}

.btn-add {
  width: 22px;
  height: 22px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background-color: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
}

.btn-add:hover {
  background-color: var(--bg-hover);
  border-color: var(--primary);
  color: var(--primary);
}

/* Connection Actions */
.connection-actions {
  display: flex;
  gap: var(--spacing-xs);
  margin-top: var(--spacing-sm);
  padding-top: var(--spacing-sm);
  border-top: 1px solid var(--border-light);
}

.btn-small {
  padding: 4px 10px;
  font-size: var(--font-size-xs);
}

.btn-danger {
  color: var(--danger);
}

.btn-danger:hover:not(:disabled) {
  background-color: var(--danger-bg);
  border-color: var(--danger-border);
}

/* Test Result */
.test-result {
  margin-top: var(--spacing-sm);
  padding: var(--spacing-sm);
  border-radius: var(--radius-md);
  font-size: var(--font-size-xs);
}

.test-result.success {
  background-color: var(--success-bg);
  border: 1px solid var(--success-border);
}

.test-result.error {
  background-color: var(--danger-bg);
  border: 1px solid var(--danger-border);
}

.result-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}

.result-icon {
  font-size: var(--font-size-sm);
}

.test-result.success .result-icon,
.test-result.success .result-status {
  color: var(--success);
}

.test-result.error .result-icon,
.test-result.error .result-status {
  color: var(--danger);
}

.result-status {
  font-weight: 600;
}

.result-details {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.result-details div {
  margin-top: 2px;
}

.result-error {
  margin-top: 4px;
  color: var(--danger);
  word-break: break-all;
}

/* Form */
.form-group {
  margin-bottom: var(--spacing-sm);
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  margin-bottom: 2px;
}

.input,
.select {
  width: 100%;
  padding: 4px 8px;
  background-color: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  height: 26px;
}

.input:focus,
.select:focus {
  outline: none;
  border-color: var(--border-focus);
}

.input:disabled,
.select:disabled {
  background-color: var(--bg-secondary);
  color: var(--text-muted);
  cursor: not-allowed;
}

/* Control Panel */
.button-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-xs);
}

.control-panel .btn {
  padding: 6px 12px;
  font-size: var(--font-size-sm);
}

/* Status Section */
.status-section {
  flex-shrink: 0;
}

.status-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.status-label {
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.status-value {
  font-weight: 600;
  font-size: var(--font-size-sm);
}

.status-value.idle { color: var(--text-muted); }
.status-value.running { color: var(--success); }
.status-value.preparing { color: var(--warning); }
.status-value.warming_up { color: var(--warning); }
.status-value.completed { color: var(--success); }
.status-value.failed { color: var(--danger); }
.status-value.cancelled,
.status-value.timeout,
.status-value.force_stopped { color: var(--warning); }

.result-summary {
  margin-top: var(--spacing-sm);
  padding-top: var(--spacing-sm);
  border-top: 1px solid var(--border-light);
}

.result-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 2px 0;
}

.result-item .result-label {
  color: var(--text-muted);
  font-size: var(--font-size-xs);
}

.result-item .result-value {
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-weight: 600;
}

/* Log Section */
.log-section {
  flex: 1;
  min-height: 120px;
  display: flex;
  flex-direction: column;
}

.log-section :deep(.log-panel) {
  flex: 1;
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background-color: var(--bg-primary);
  border-radius: var(--radius-lg);
  padding: 0;
  width: auto;
  max-width: 90vw;
  max-height: 90vh;
  overflow: visible;
  box-shadow: var(--shadow-modal);
  border: 1px solid var(--border-color);
}

/* Template Info */
.template-info {
  margin-top: var(--spacing-sm);
  padding-top: var(--spacing-sm);
  border-top: 1px solid var(--border-light);
}

.template-description {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  font-style: italic;
}
</style>
