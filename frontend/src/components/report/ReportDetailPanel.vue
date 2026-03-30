<template>
  <div class="report-detail-panel" v-if="report">
    <!-- Header -->
    <div class="detail-header">
      <div class="header-left">
        <h2 class="detail-title">报告详情</h2>
        <span class="report-id">{{ report.id }}</span>
      </div>
      <div class="header-actions">
        <!-- Export Buttons -->
        <div class="export-buttons">
          <button
            class="btn btn-export"
            :disabled="reportStore.exporting"
            @click="handleExportJSON"
            title="导出 JSON"
          >
            <span v-if="reportStore.exporting" class="btn-spinner"></span>
            <span v-else>📄</span>
            JSON
          </button>
          <button
            class="btn btn-export"
            :disabled="reportStore.exporting"
            @click="handleExportHTML"
            title="导出 HTML"
          >
            <span v-if="reportStore.exporting" class="btn-spinner"></span>
            <span v-else>🌐</span>
            HTML
          </button>
          <button
            class="btn btn-export"
            :disabled="reportStore.exporting"
            @click="handleCopyJSON"
            title="复制 JSON 到剪贴板"
          >
            📋 复制
          </button>
        </div>
        <button class="btn btn-secondary" @click="$emit('close')">
          关闭
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <span>加载报告详情...</span>
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
        <h3 class="section-title">基本信息</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">来源</span>
            <span class="info-value">{{ formatSourceType(report.source_type) }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">数据库</span>
            <span class="info-value">{{ report.database_type }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">连接</span>
            <span class="info-value">{{ report.connection_name || report.connection_id }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">模板</span>
            <span class="info-value">{{ report.template_name || report.template_id || 'N/A' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">状态</span>
            <span class="info-value">
              <span class="status-badge" :class="getStatusClass(report.status)">
                {{ formatStatus(report.status) }}
              </span>
            </span>
          </div>
          <div class="info-item" v-if="report.error_message">
            <span class="info-label">错误</span>
            <span class="info-value error-text">{{ report.error_message }}</span>
          </div>
        </div>
      </section>

      <!-- Timing Section -->
      <section class="detail-section">
        <h3 class="section-title">执行时间</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">开始时间</span>
            <span class="info-value">{{ formatDateTime(report.started_at) }}</span>
          </div>
          <div class="info-item" v-if="report.ended_at">
            <span class="info-label">结束时间</span>
            <span class="info-value">{{ formatDateTime(report.ended_at) }}</span>
          </div>
          <div class="info-item" v-if="report.duration_ms">
            <span class="info-label">持续时间</span>
            <span class="info-value">{{ formatDuration(report.duration_ms) }}</span>
          </div>
        </div>
      </section>

      <!-- Performance Charts Section -->
      <section class="detail-section" v-if="hasPerformanceData">
        <h3 class="section-title">性能指标图表</h3>
        <div class="charts-grid">
          <!-- TPM Chart -->
          <div class="chart-card" v-if="metrics?.summary?.tpm || report.tpm">
            <ReportMetricChart
              metric-key="tpm"
              label="TPM"
              :current-value="metrics?.summary?.tpm || report.tpm"
              :data-points="tpsTimeSeriesData"
              color="#DC5050"
              unit="/min"
              :show-sparkline="hasTimeSeriesData"
            />
          </div>

          <!-- TPS Chart -->
          <div class="chart-card" v-if="metrics?.summary?.tps || report.tps">
            <ReportMetricChart
              metric-key="tps"
              label="TPS"
              :current-value="metrics?.summary?.tps || report.tps"
              :data-points="tpsTimeSeriesData"
              color="#FFA500"
              unit="/sec"
              :show-sparkline="hasTimeSeriesData"
            />
          </div>

          <!-- QPS Chart -->
          <div class="chart-card" v-if="metrics?.summary?.qps || report.qps">
            <ReportMetricChart
              metric-key="qps"
              label="QPS"
              :current-value="metrics?.summary?.qps || report.qps"
              :data-points="[]"
              color="#4CAF50"
              unit="/sec"
              :show-sparkline="false"
            />
          </div>

          <!-- Latency Chart -->
          <div class="chart-card" v-if="hasLatencyData">
            <ReportMetricChart
              metric-key="latency"
              label="延迟"
              :current-value="metrics?.summary?.latency_avg_ms || report.latency_avg_ms"
              :data-points="latencyTimeSeriesData"
              color="#9C27B0"
              unit="ms"
              :show-sparkline="hasTimeSeriesData"
            />
          </div>
        </div>
      </section>

      <!-- Performance Metrics Summary Section -->
      <section class="detail-section">
        <h3 class="section-title">性能指标汇总</h3>
        <div class="metrics-grid">
          <div class="metric-card">
            <span class="metric-label">TPM</span>
            <span class="metric-value">{{ formatNumber(metrics?.summary?.tpm || report.tpm) }}</span>
            <span class="metric-unit">trans/min</span>
          </div>
          <div class="metric-card">
            <span class="metric-label">TPS</span>
            <span class="metric-value">{{ formatNumber(metrics?.summary?.tps || report.tps) }}</span>
            <span class="metric-unit">trans/sec</span>
          </div>
          <div class="metric-card" v-if="metrics?.summary?.qps || report.qps">
            <span class="metric-label">QPS</span>
            <span class="metric-value">{{ formatNumber(metrics?.summary?.qps || report.qps) }}</span>
            <span class="metric-unit">queries/sec</span>
          </div>
          <div class="metric-card" v-if="metrics?.summary?.throughput || report.throughput">
            <span class="metric-label">吞吐量</span>
            <span class="metric-value">{{ formatNumber(metrics?.summary?.throughput || report.throughput) }}</span>
            <span class="metric-unit">ops/sec</span>
          </div>
        </div>
      </section>

      <!-- Latency Section -->
      <section class="detail-section" v-if="hasLatencyData">
        <h3 class="section-title">延迟分布</h3>
        <div class="metrics-grid">
          <div class="metric-card latency-avg">
            <span class="metric-label">平均值</span>
            <span class="metric-value">{{ formatNumber(metrics?.summary?.latency_avg_ms || report.latency_avg_ms) }}</span>
            <span class="metric-unit">ms</span>
          </div>
          <div class="metric-card latency-p95" v-if="metrics?.summary?.latency_p95_ms || report.latency_p95_ms">
            <span class="metric-label">P95</span>
            <span class="metric-value">{{ formatNumber(metrics?.summary?.latency_p95_ms || report.latency_p95_ms) }}</span>
            <span class="metric-unit">ms</span>
          </div>
          <div class="metric-card latency-p99" v-if="metrics?.summary?.latency_p99_ms || report.latency_p99_ms">
            <span class="metric-label">P99</span>
            <span class="metric-value">{{ formatNumber(metrics?.summary?.latency_p99_ms || report.latency_p99_ms) }}</span>
            <span class="metric-unit">ms</span>
          </div>
        </div>

        <!-- Percentiles Detail -->
        <div class="percentiles-section" v-if="hasPercentilesData">
          <h4 class="subsection-title">百分位详情</h4>
          <div class="percentiles-grid">
            <div
              v-for="(value, percentile) in metrics.percentiles"
              :key="percentile"
              class="percentile-item"
            >
              <span class="percentile-label">{{ percentile }}</span>
              <span class="percentile-value">{{ formatNumber(value) }} ms</span>
            </div>
          </div>
        </div>
      </section>

      <!-- Error Count Section -->
      <section class="detail-section" v-if="metrics?.summary?.error_count || report.error_count">
        <h3 class="section-title">错误信息</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">错误数量</span>
            <span class="info-value error-count">{{ metrics?.summary?.error_count || report.error_count }}</span>
          </div>
        </div>
      </section>

      <!-- Suite Info -->
      <section class="detail-section" v-if="report.suite_id && report.suite_id !== 'standalone'">
        <h3 class="section-title">套件信息</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">套件 ID</span>
            <span class="info-value">{{ report.suite_id }}</span>
          </div>
          <div class="info-item" v-if="report.suite_item_id">
            <span class="info-label">套件项 ID</span>
            <span class="info-value">{{ report.suite_item_id }}</span>
          </div>
        </div>
      </section>

      <!-- Raw Metrics JSON (Collapsible) -->
      <section class="detail-section" v-if="metrics">
        <div class="collapsible-header" @click="toggleRawMetrics">
          <h3 class="section-title">原始指标数据</h3>
          <span class="collapse-icon">{{ showRawMetrics ? '▼' : '▶' }}</span>
        </div>
        <div class="metrics-detail" v-if="showRawMetrics">
          <pre class="metrics-json">{{ JSON.stringify(metrics, null, 2) }}</pre>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useReportStore } from '../../stores/report'
import ReportMetricChart from './ReportMetricChart.vue'

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
const showRawMetrics = ref(false)

// Computed properties for chart data
const hasPerformanceData = computed(() => {
  return report.value?.tpm || report.value?.tps || report.value?.qps ||
         metrics.value?.summary?.tpm || metrics.value?.summary?.tps || metrics.value?.summary?.qps
})

const hasLatencyData = computed(() => {
  return report.value?.latency_avg_ms || report.value?.latency_p95_ms || report.value?.latency_p99_ms ||
         metrics.value?.summary?.latency_avg_ms || metrics.value?.summary?.latency_p95_ms || metrics.value?.summary?.latency_p99_ms
})

const hasTimeSeriesData = computed(() => {
  return metrics.value?.time_series && metrics.value.time_series.length > 0
})

const hasPercentilesData = computed(() => {
  return metrics.value?.percentiles && Object.keys(metrics.value.percentiles).length > 0
})

// Convert time series data to chart format
const tpsTimeSeriesData = computed(() => {
  if (!metrics.value?.time_series) return []
  return metrics.value.time_series.map(item => ({
    value: item.tps,
    timestamp: item.timestamp
  }))
})

const latencyTimeSeriesData = computed(() => {
  if (!metrics.value?.time_series) return []
  return metrics.value.time_series.map(item => ({
    value: item.latency_avg,
    timestamp: item.timestamp
  }))
})

const fetchReportDetail = async () => {
  loading.value = true
  error.value = null

  try {
    // Fetch report details
    await reportStore.fetchReport(props.reportId)
    report.value = reportStore.selectedReport

    // Fetch metrics
    await reportStore.fetchReportMetrics(props.reportId)
    metrics.value = reportStore.selectedReportMetrics
  } catch (err) {
    error.value = err.message || '加载报告详情失败'
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

const toggleRawMetrics = () => {
  showRawMetrics.value = !showRawMetrics.value
}

const formatSourceType = (source) => {
  const labels = {
    benchmark: '单次压测',
    autobench: 'AutoBench 套件'
  }
  return labels[source] || source
}

const formatStatus = (status) => {
  const labels = {
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
    running: '运行中',
    pending: '等待中'
  }
  return labels[status] || status
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
    return new Date(dateStr).toLocaleString('zh-CN')
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
    return `${hours}小时 ${minutes % 60}分 ${seconds % 60}秒`
  }
  if (minutes > 0) {
    return `${minutes}分 ${seconds % 60}秒`
  }
  return `${seconds}秒`
}

const formatNumber = (num) => {
  if (num === null || num === undefined) return 'N/A'
  return typeof num === 'number' ? num.toFixed(2) : num
}

// 导出处理函数
const handleExportJSON = async () => {
  if (!report.value) return
  const filename = `report-${report.value.id}-${new Date().toISOString().slice(0, 10)}.json`
  await reportStore.downloadJSON(report.value.id, filename)
}

const handleExportHTML = async () => {
  if (!report.value) return
  const filename = `report-${report.value.id}-${new Date().toISOString().slice(0, 10)}.html`
  await reportStore.downloadHTML(report.value.id, filename)
}

const handleCopyJSON = async () => {
  if (!report.value) return
  await reportStore.copyJSONToClipboard(report.value.id)
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
  margin-right: 8px;
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

.btn-export {
  background-color: var(--primary-light);
  color: var(--primary);
  border-color: var(--primary);
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-export:hover:not(:disabled) {
  background-color: var(--primary);
  color: white;
}

.btn-export:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.export-buttons {
  display: flex;
  gap: var(--spacing-xs);
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

.subsection-title {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--text-secondary);
  margin-top: var(--spacing-md);
  margin-bottom: var(--spacing-sm);
}

/* Collapsible Header */
.collapsible-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  user-select: none;
}

.collapsible-header:hover .section-title {
  color: var(--primary);
}

.collapse-icon {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
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

/* Charts Grid */
.charts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--spacing-md);
}

.chart-card {
  background-color: var(--bg-secondary);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-light);
  padding: var(--spacing-sm);
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

/* Latency color coding */
.latency-avg .metric-value {
  color: var(--primary);
}

.latency-p95 .metric-value {
  color: var(--warning);
}

.latency-p99 .metric-value {
  color: var(--danger);
}

/* Percentiles Grid */
.percentiles-section {
  margin-top: var(--spacing-md);
}

.percentiles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: var(--spacing-sm);
}

.percentile-item {
  display: flex;
  justify-content: space-between;
  padding: var(--spacing-xs) var(--spacing-sm);
  background-color: var(--bg-secondary);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
}

.percentile-label {
  color: var(--text-muted);
}

.percentile-value {
  font-family: var(--font-family-mono);
  font-weight: 500;
  color: var(--text-primary);
}

/* Metrics Detail */
.metrics-detail {
  background-color: var(--bg-secondary);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  overflow-x: auto;
  margin-top: var(--spacing-sm);
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
