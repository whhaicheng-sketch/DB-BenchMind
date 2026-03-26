<template>
  <div class="report-detail-panel" v-if="report">
    <!-- Header -->
    <div class="detail-header">
      <div class="header-left">
        <h2 class="detail-title">Report Details</h2>
        <span class="report-id">{{ report.id }}</span>
      </div>
      <div class="header-actions">
        <button class="btn btn-secondary" @click="$emit('close')">
          Close
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <span>Loading report details...</span>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error-state">
      <span class="error-icon">!</span>
      <span class="error-message">{{ error }}</span>
    </div>

    <!-- Report Content -->
    <div v-else class="detail-content">
      <!-- Basic Info Section -->
      <section class="detail-section">
        <h3 class="section-title">Basic Information</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">Source</span>
            <span class="info-value">{{ formatSourceType(report.source_type) }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Database</span>
            <span class="info-value">{{ report.database_type }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Connection</span>
            <span class="info-value">{{ report.connection_name || report.connection_id }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Template</span>
            <span class="info-value">{{ report.template_name || report.template_id || 'N/A' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Status</span>
            <span class="info-value">
              <span class="status-badge" :class="getStatusClass(report.status)">
                {{ formatStatus(report.status) }}
              </span>
            </span>
          </div>
          <div class="info-item" v-if="report.error_message">
            <span class="info-label">Error</span>
            <span class="info-value error-text">{{ report.error_message }}</span>
          </div>
        </div>
      </section>

      <!-- Timing Section -->
      <section class="detail-section">
        <h3 class="section-title">Timing</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">Started</span>
            <span class="info-value">{{ formatDateTime(report.started_at) }}</span>
          </div>
          <div class="info-item" v-if="report.ended_at">
            <span class="info-label">Ended</span>
            <span class="info-value">{{ formatDateTime(report.ended_at) }}</span>
          </div>
          <div class="info-item" v-if="report.duration_ms">
            <span class="info-label">Duration</span>
            <span class="info-value">{{ formatDuration(report.duration_ms) }}</span>
          </div>
        </div>
      </section>

      <!-- Performance Metrics Section -->
      <section class="detail-section">
        <h3 class="section-title">Performance Metrics</h3>
        <div class="metrics-grid">
          <div class="metric-card">
            <span class="metric-label">TPM</span>
            <span class="metric-value">{{ formatNumber(report.tpm) }}</span>
            <span class="metric-unit">trans/min</span>
          </div>
          <div class="metric-card">
            <span class="metric-label">TPS</span>
            <span class="metric-value">{{ formatNumber(report.tps) }}</span>
            <span class="metric-unit">trans/sec</span>
          </div>
          <div class="metric-card" v-if="report.qps">
            <span class="metric-label">QPS</span>
            <span class="metric-value">{{ formatNumber(report.qps) }}</span>
            <span class="metric-unit">queries/sec</span>
          </div>
          <div class="metric-card" v-if="report.throughput">
            <span class="metric-label">Throughput</span>
            <span class="metric-value">{{ formatNumber(report.throughput) }}</span>
            <span class="metric-unit">ops/sec</span>
          </div>
        </div>
      </section>

      <!-- Latency Section -->
      <section class="detail-section" v-if="hasLatencyData">
        <h3 class="section-title">Latency</h3>
        <div class="metrics-grid">
          <div class="metric-card">
            <span class="metric-label">Average</span>
            <span class="metric-value">{{ formatNumber(report.latency_avg_ms) }}</span>
            <span class="metric-unit">ms</span>
          </div>
          <div class="metric-card" v-if="report.latency_p95_ms">
            <span class="metric-label">P95</span>
            <span class="metric-value">{{ formatNumber(report.latency_p95_ms) }}</span>
            <span class="metric-unit">ms</span>
          </div>
          <div class="metric-card" v-if="report.latency_p99_ms">
            <span class="metric-label">P99</span>
            <span class="metric-value">{{ formatNumber(report.latency_p99_ms) }}</span>
            <span class="metric-unit">ms</span>
          </div>
        </div>
      </section>

      <!-- Error Count Section -->
      <section class="detail-section" v-if="report.error_count">
        <h3 class="section-title">Errors</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">Error Count</span>
            <span class="info-value error-count">{{ report.error_count }}</span>
          </div>
        </div>
      </section>

      <!-- Metrics Detail Section -->
      <section class="detail-section" v-if="metrics">
        <h3 class="section-title">Detailed Metrics</h3>
        <div class="metrics-detail">
          <pre class="metrics-json">{{ JSON.stringify(metrics, null, 2) }}</pre>
        </div>
      </section>

      <!-- Suite Info -->
      <section class="detail-section" v-if="report.suite_id && report.suite_id !== 'standalone'">
        <h3 class="section-title">Suite Information</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">Suite ID</span>
            <span class="info-value">{{ report.suite_id }}</span>
          </div>
          <div class="info-item" v-if="report.suite_item_id">
            <span class="info-label">Suite Item ID</span>
            <span class="info-value">{{ report.suite_item_id }}</span>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useReportStore } from '../../stores/report'

const props = defineProps({
  reportId: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['close'])

const reportStore = useReportStore()
const report = ref(null)
const metrics = ref(null)
const loading = ref(true)
const error = ref(null)

const hasLatencyData = computed(() => {
  return props.report && (report.value?.latency_avg_ms || report.value?.latency_p95_ms || report.value?.latency_p99_ms)
})

const fetchReportDetail = async () => {
  loading.value = true
  error.value = null

  try {
    // Fetch report details
    await reportStore.fetchReport(props.reportId)
    report.value = reportStore.selectedReport

    // Fetch metrics
    const metricsData = await reportStore.fetchReportMetrics(props.reportId)
    metrics.value = metricsData
  } catch (err) {
    error.value = err.message || 'Failed to load report details'
  } finally {
    loading.value = false
  }
}

watch(() => props.reportId, () => {
  if (props.reportId) {
    fetchReportDetail()
  }
})

onMounted(() => {
  if (props.reportId) {
    fetchReportDetail()
  }
})

const formatSourceType = (source) => {
  const labels = {
    benchmark: 'Single Benchmark',
    autobench: 'AutoBench Suite'
  }
  return labels[source] || source
}

const formatStatus = (status) => {
  return status?.charAt(0).toUpperCase() + status?.slice(1) || 'Unknown'
}

const getStatusClass = (status) => {
  const classMap = {
    completed: 'status-success',
    failed: 'status-error',
    cancelled: 'status-warning',
    running: 'status-info',
    pending: 'status-default'
  }
  return classMap[status] || 'status-default'
}

const formatDateTime = (dateStr) => {
  if (!dateStr) return 'N/A'
  try {
    return new Date(dateStr).toLocaleString()
  } catch {
    return dateStr
  }
}

const formatDuration = (ms) => {
  if (!ms) return 'N/A'
  const seconds = Math.floor(ms / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m ${seconds % 60}s`
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`
  }
  return `${seconds}s`
}

const formatNumber = (num) => {
  if (num === null || num === undefined) return 'N/A'
  return typeof num === 'number' ? num.toFixed(2) : num
}
</script>

<style scoped>
.report-detail-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-primary);
  overflow: hidden;
}

/* Header */
.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--border-color);
  background-color: var(--bg-secondary);
}

.detail-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.report-id {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  font-family: var(--font-family-mono);
}

.header-actions {
  display: flex;
  gap: var(--spacing-sm);
}

/* Buttons */
.btn {
  padding: 8px 16px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-secondary {
  background-color: var(--bg-primary);
  color: var(--text-primary);
}

.btn-secondary:hover {
  background-color: var(--bg-hover);
  border-color: var(--border-dark);
}

/* Loading & Error States */
.loading-state,
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  color: var(--text-muted);
  gap: var(--spacing-md);
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background-color: var(--danger-bg);
  color: var(--danger);
  font-size: 24px;
  font-weight: bold;
}

.error-message {
  color: var(--danger);
}

/* Detail Content */
.detail-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-lg);
}

/* Sections */
.detail-section {
  margin-bottom: var(--spacing-xl);
}

.section-title {
  font-size: var(--font-size-base);
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: var(--spacing-md);
  padding-bottom: var(--spacing-xs);
  border-bottom: 1px solid var(--border-light);
}

/* Info Grid */
.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: var(--spacing-md);
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-label {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.info-value {
  font-size: var(--font-size-base);
  color: var(--text-primary);
  font-weight: 500;
}

.error-text {
  color: var(--danger);
}

.error-count {
  color: var(--danger);
  font-weight: 600;
}

/* Status Badge */
.status-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.status-success {
  background-color: var(--success-bg);
  color: var(--success);
}

.status-error {
  background-color: var(--danger-bg);
  color: var(--danger);
}

.status-warning {
  background-color: var(--warning-bg);
  color: var(--warning);
}

.status-info {
  background-color: var(--primary-light);
  color: var(--primary);
}

.status-default {
  background-color: var(--bg-secondary);
  color: var(--text-muted);
}

/* Metrics Grid */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: var(--spacing-md);
}

.metric-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--spacing-md);
  background-color: var(--bg-secondary);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-light);
}

.metric-label {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 4px;
}

.metric-value {
  font-size: var(--font-size-xl);
  font-weight: 700;
  color: var(--text-primary);
  font-family: var(--font-family-mono);
}

.metric-unit {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  margin-top: 2px;
}

/* Metrics Detail */
.metrics-detail {
  background-color: var(--bg-secondary);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  overflow-x: auto;
}

.metrics-json {
  margin: 0;
  font-size: var(--font-size-sm);
  font-family: var(--font-family-mono);
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
