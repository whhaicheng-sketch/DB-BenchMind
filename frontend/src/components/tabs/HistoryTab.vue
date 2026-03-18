<template>
  <div class="history-tab">
    <!-- Page Header -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">History</h1>
        <p class="page-subtitle">View benchmark run history and results</p>
      </div>
      <div class="header-right">
        <button class="btn" @click="refreshHistory">
          Refresh
        </button>
      </div>
    </div>

    <!-- History List -->
    <div class="history-list">
      <div v-if="histories.length === 0" class="empty-state">
        <div class="empty-state-icon">📋</div>
        <p class="empty-state-title">No benchmark history</p>
        <p class="empty-state-description">Run a benchmark task to see results appear here.</p>
      </div>

      <div v-else class="history-items">
        <div v-for="item in histories" :key="item.id" class="history-item">
          <div class="history-info">
            <div class="history-name">{{ item.name || 'Unnamed Task' }}</div>
            <div class="history-date">{{ formatDate(item.start_time) }}</div>
          </div>
          <div class="history-metrics">
            <div class="metric">
              <span class="metric-label">TPM</span>
              <span class="metric-value">{{ item.tpm?.toFixed(2) || 'N/A' }}</span>
            </div>
            <div class="metric">
              <span class="metric-label">TPS</span>
              <span class="metric-value">{{ item.tps?.toFixed(2) || 'N/A' }}</span>
            </div>
            <div class="metric">
              <span class="metric-label">Latency</span>
              <span class="metric-value">{{ item.latency_avg_ms?.toFixed(2) || 'N/A' }}ms</span>
            </div>
          </div>
          <div class="history-status">
            <span class="status-badge" :class="getStatusClass(item.status)">
              {{ getStatusText(item.status) }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useBenchmarkStore } from '../../stores/benchmark'

const benchmarkStore = useBenchmarkStore()
const histories = ref([])

onMounted(() => {
  refreshHistory()
})

const refreshHistory = async () => {
  // TODO: Implement history fetch from store/API
  // For now, just use empty array
  histories.value = []
}

const formatDate = (date) => {
  if (!date) return ''
  return new Date(date).toLocaleString()
}

const getStatusClass = (status) => {
  const classMap = {
    completed: 'status-success',
    failed: 'status-error',
    cancelled: 'status-warning',
    running: 'status-info'
  }
  return classMap[status] || 'status-default'
}

const getStatusText = (status) => {
  const textMap = {
    completed: 'Completed',
    failed: 'Failed',
    cancelled: 'Cancelled',
    running: 'Running'
  }
  return textMap[status] || status
}
</script>

<style scoped>
.history-tab {
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

/* History List */
.history-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.history-items {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.history-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-md);
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  gap: var(--spacing-md);
}

.history-item:hover {
  border-color: var(--border-dark);
  background-color: var(--bg-hover);
}

.history-info {
  flex: 1;
  min-width: 0;
}

.history-name {
  font-weight: 500;
  color: var(--text-primary);
  font-size: var(--font-size-base);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.history-date {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
  margin-top: 2px;
}

.history-metrics {
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

.history-status {
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
</style>
