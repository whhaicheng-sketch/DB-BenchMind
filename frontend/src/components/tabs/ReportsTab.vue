<template>
  <div class="reports-tab">
    <!-- Page Header -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">Reports</h1>
        <p class="page-subtitle">View benchmark reports and results</p>
      </div>
      <div class="header-right">
        <button class="btn" @click="refreshReports">
          Refresh
        </button>
      </div>
    </div>

    <!-- Main Content -->
    <div class="reports-content">
      <!-- Reports List -->
      <div class="reports-list-section">
        <!-- Filter Bar -->
        <div class="filter-bar">
          <select v-model="statusFilter" class="filter-select" @change="onFilterChange">
            <option value="">All Status</option>
            <option value="completed">Completed</option>
            <option value="failed">Failed</option>
            <option value="cancelled">Cancelled</option>
            <option value="running">Running</option>
          </select>
        </div>

        <!-- Loading State -->
        <div v-if="reportStore.loading" class="loading-state">
          <div class="spinner"></div>
          <span>Loading reports...</span>
        </div>

        <!-- Empty State -->
        <div v-else-if="reportStore.reports.length === 0" class="empty-state">
          <div class="empty-state-icon">📋</div>
          <p class="empty-state-title">No benchmark reports</p>
          <p class="empty-state-description">Run a benchmark task to see results appear here.</p>
        </div>

        <!-- Reports List -->
        <div v-else class="reports-list">
          <div
            v-for="report in reportStore.reports"
            :key="report.id"
            class="report-item"
            :class="{ selected: selectedReportId === report.id }"
            @click="selectReport(report.id)"
          >
            <div class="report-info">
              <div class="report-name">
                {{ report.template_name || report.database_type || 'Unnamed Report' }}
              </div>
              <div class="report-meta">
                <span class="report-date">{{ formatDate(report.started_at) }}</span>
                <span class="report-source">{{ formatSourceType(report.source_type) }}</span>
              </div>
            </div>
            <div class="report-metrics">
              <div class="metric">
                <span class="metric-label">TPM</span>
                <span class="metric-value">{{ formatNumber(report.tpm) }}</span>
              </div>
              <div class="metric">
                <span class="metric-label">TPS</span>
                <span class="metric-value">{{ formatNumber(report.tps) }}</span>
              </div>
              <div class="metric">
                <span class="metric-label">Latency</span>
                <span class="metric-value">{{ formatLatency(report.latency_avg_ms) }}</span>
              </div>
            </div>
            <div class="report-status">
              <span class="status-badge" :class="getStatusClass(report.status)">
                {{ getStatusText(report.status) }}
              </span>
            </div>
          </div>
        </div>

        <!-- Pagination -->
        <div v-if="reportStore.reports.length > 0" class="pagination">
          <span class="pagination-info">
            {{ reportStore.pagination.total }} reports
          </span>
        </div>
      </div>

      <!-- Detail Panel -->
      <div class="detail-section" v-if="selectedReportId">
        <ReportDetailPanel
          :report-id="selectedReportId"
          @close="closeDetail"
        />
      </div>
    </div>

    <!-- Notice Toast -->
    <div v-if="reportStore.notice" class="notice-toast" :class="`notice-${reportStore.notice.tone}`">
      {{ reportStore.notice.message }}
      <button class="notice-close" @click="reportStore.clearNotice()">×</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useReportStore } from '../../stores/report'
import ReportDetailPanel from '../report/ReportDetailPanel.vue'

const reportStore = useReportStore()
const selectedReportId = ref(null)
const statusFilter = ref('')

onMounted(() => {
  refreshReports()
})

const refreshReports = async () => {
  await reportStore.fetchReports()
}

const selectReport = (id) => {
  selectedReportId.value = id
}

const closeDetail = () => {
  selectedReportId.value = null
}

const onFilterChange = async () => {
  reportStore.setFilter('status', statusFilter.value)
  await reportStore.fetchReports()
}

const formatDate = (date) => {
  if (!date) return ''
  try {
    return new Date(date).toLocaleString()
  } catch {
    return date
  }
}

const formatSourceType = (source) => {
  const labels = {
    benchmark: 'Single',
    autobench: 'Suite'
  }
  return labels[source] || source
}

const formatNumber = (num) => {
  if (num === null || num === undefined) return 'N/A'
  return typeof num === 'number' ? num.toFixed(2) : num
}

const formatLatency = (ms) => {
  if (ms === null || ms === undefined) return 'N/A'
  return `${ms.toFixed(2)}ms`
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

const getStatusText = (status) => {
  const textMap = {
    completed: 'Completed',
    failed: 'Failed',
    cancelled: 'Cancelled',
    running: 'Running',
    pending: 'Pending'
  }
  return textMap[status] || status
}
</script>

<style scoped>
.reports-tab {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* Page Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding-bottom: var(--spacing-lg);
  border-bottom: 1px solid var(--border-light);
  margin-bottom: var(--spacing-lg);
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  font-size: var(--font-size-title);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.page-subtitle {
  font-size: var(--font-size-md);
  color: var(--text-muted);
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.btn {
  padding: 8px 16px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background-color: var(--bg-primary);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn:hover {
  background-color: var(--bg-secondary);
  border-color: var(--border-dark);
}

/* Main Content */
.reports-content {
  flex: 1;
  display: flex;
  gap: var(--spacing-lg);
  overflow: hidden;
}

.reports-list-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 300px;
  overflow: hidden;
}

/* Filter Bar */
.filter-bar {
  display: flex;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
}

.filter-select {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background-color: var(--bg-primary);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  cursor: pointer;
}

/* Loading State */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
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

/* Reports List */
.reports-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.report-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-md);
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  gap: var(--spacing-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.report-item:hover {
  border-color: var(--border-dark);
  background-color: var(--bg-hover);
}

.report-item.selected {
  border-color: var(--primary);
  background-color: var(--primary-light);
}

.report-info {
  flex: 1;
  min-width: 0;
}

.report-name {
  font-weight: 500;
  color: var(--text-primary);
  font-size: var(--font-size-base);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.report-meta {
  display: flex;
  gap: var(--spacing-sm);
  margin-top: 4px;
}

.report-date {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
}

.report-source {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
  padding: 2px 6px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-sm);
}

.report-metrics {
  display: flex;
  gap: var(--spacing-lg);
}

.metric {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.metric-label {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.metric-value {
  font-size: var(--font-size-md);
  font-weight: 600;
  color: var(--text-primary);
  font-family: var(--font-family-mono);
}

.report-status {
  display: flex;
  align-items: center;
}

.status-badge {
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

/* Pagination */
.pagination {
  display: flex;
  justify-content: center;
  padding: var(--spacing-md) 0;
}

.pagination-info {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
}

/* Detail Section */
.detail-section {
  width: 400px;
  flex-shrink: 0;
  border-left: 1px solid var(--border-light);
  overflow: hidden;
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.empty-state-icon {
  font-size: 48px;
  margin-bottom: var(--spacing-lg);
  opacity: 0.5;
}

.empty-state-title {
  font-size: var(--font-size-lg);
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: var(--spacing-sm);
}

.empty-state-description {
  font-size: var(--font-size-md);
  color: var(--text-muted);
  max-width: 400px;
}

/* Notice Toast */
.notice-toast {
  position: fixed;
  bottom: 20px;
  right: 20px;
  padding: 12px 20px;
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1000;
}

.notice-info {
  background-color: var(--primary);
  color: white;
}

.notice-success {
  background-color: var(--success);
  color: white;
}

.notice-warning {
  background-color: var(--warning);
  color: white;
}

.notice-error {
  background-color: var(--danger);
  color: white;
}

.notice-close {
  background: none;
  border: none;
  color: inherit;
  font-size: 18px;
  cursor: pointer;
  padding: 0;
  line-height: 1;
  opacity: 0.8;
}

.notice-close:hover {
  opacity: 1;
}
</style>
