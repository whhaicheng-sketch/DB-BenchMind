<template>
  <div class="impact-summary-cards">
    <!-- Business Interruption Card -->
    <div class="summary-card">
      <div class="card-header">
        <span class="card-icon">⏸️</span>
        <span class="card-title">Business Interruption</span>
      </div>
      <div class="card-value">
        <span class="value">{{ formatValue(summaryData?.businessInterruption?.value) }}</span>
        <span class="unit">{{ summaryData?.businessInterruption?.unit || 'ms' }}</span>
      </div>
      <div class="card-description">
        Duration of TPS dropping to zero
      </div>
    </div>

    <!-- RTO Card -->
    <div class="summary-card">
      <div class="card-header">
        <span class="card-icon">⏱️</span>
        <span class="card-title">RTO</span>
      </div>
      <div class="card-value">
        <span class="value">{{ formatValue(summaryData?.rto?.value) }}</span>
        <span class="unit">{{ summaryData?.rto?.unit || 'ms' }}</span>
      </div>
      <div class="card-description">
        Recovery Time Objective
      </div>
    </div>

    <!-- Consistency Card -->
    <div class="summary-card" :class="consistencyClass">
      <div class="card-header">
        <span class="card-icon">{{ consistencyIcon }}</span>
        <span class="card-title">Consistency</span>
      </div>
      <div class="card-value">
        <span class="value consistency-value">{{ consistencyLabel }}</span>
      </div>
      <div class="card-description">
        Committed data verification
      </div>
    </div>

    <!-- Commit Stats Card -->
    <div class="summary-card">
      <div class="card-header">
        <span class="card-icon">📊</span>
        <span class="card-title">Commit Stats</span>
      </div>
      <div class="card-stats">
        <div class="stat-row success">
          <span class="stat-label">Success</span>
          <span class="stat-value">{{ formatNumber(summaryData?.commitStats?.successCount) }}</span>
        </div>
        <div class="stat-row error">
          <span class="stat-label">Errors</span>
          <span class="stat-value">{{ formatNumber(summaryData?.commitStats?.errorCount) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { ConsistencyResult } from '../constants'

const props = defineProps({
  summaryData: {
    type: Object,
    default: () => null
  }
})

const consistencyClass = computed(() => {
  const result = props.summaryData?.consistency?.result
  if (result === ConsistencyResult.PASSED) return 'passed'
  if (result === ConsistencyResult.FAILED) return 'failed'
  return 'pending'
})

const consistencyIcon = computed(() => {
  const result = props.summaryData?.consistency?.result
  if (result === ConsistencyResult.PASSED) return '✅'
  if (result === ConsistencyResult.FAILED) return '❌'
  return '⏳'
})

const consistencyLabel = computed(() => {
  const result = props.summaryData?.consistency?.result
  if (result === ConsistencyResult.PASSED) return 'Passed'
  if (result === ConsistencyResult.FAILED) return 'Failed'
  return 'Pending'
})

function formatValue(value) {
  if (value === null || value === undefined) return '--'
  return value.toLocaleString()
}

function formatNumber(value) {
  if (value === null || value === undefined) return '0'
  return value.toLocaleString()
}
</script>

<style scoped>
.impact-summary-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.summary-card {
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px 20px;
  transition: border-color var(--transition-fast);
}

.summary-card:hover {
  border-color: var(--border-dark);
}

.summary-card.passed {
  border-left: 3px solid var(--success);
}

.summary-card.failed {
  border-left: 3px solid var(--danger);
}

.summary-card.pending {
  border-left: 3px solid var(--warning);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.card-icon {
  font-size: 18px;
}

.card-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.card-value {
  display: flex;
  align-items: baseline;
  gap: 6px;
  margin-bottom: 8px;
}

.card-value .value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  font-family: var(--font-family-mono);
}

.card-value .unit {
  font-size: 14px;
  color: var(--text-muted);
}

.consistency-value {
  font-size: 24px !important;
}

.summary-card.passed .consistency-value {
  color: var(--success);
}

.summary-card.failed .consistency-value {
  color: var(--danger);
}

.summary-card.pending .consistency-value {
  color: var(--warning);
}

.card-description {
  font-size: 12px;
  color: var(--text-muted);
}

.card-stats {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 10px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-sm);
}

.stat-row.success {
  border-left: 2px solid var(--success);
}

.stat-row.error {
  border-left: 2px solid var(--danger);
}

.stat-label {
  font-size: 13px;
  color: var(--text-muted);
}

.stat-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  font-family: var(--font-family-mono);
}

@media (max-width: 1200px) {
  .impact-summary-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .impact-summary-cards {
    grid-template-columns: 1fr;
  }
}
</style>
