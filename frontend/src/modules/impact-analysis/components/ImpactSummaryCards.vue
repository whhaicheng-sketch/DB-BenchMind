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
  background-color: #1a2332;
  border: 1px solid #2a3a4a;
  border-radius: 8px;
  padding: 16px 20px;
  transition: border-color 0.2s;
}

.summary-card:hover {
  border-color: #3a4a5a;
}

.summary-card.passed {
  border-left: 3px solid #48bb78;
}

.summary-card.failed {
  border-left: 3px solid #f56565;
}

.summary-card.pending {
  border-left: 3px solid #ecc94b;
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
  color: #a0aec0;
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
  color: #e2e8f0;
  font-family: 'SF Mono', Monaco, monospace;
}

.card-value .unit {
  font-size: 14px;
  color: #718096;
}

.consistency-value {
  font-size: 24px !important;
}

.summary-card.passed .consistency-value {
  color: #48bb78;
}

.summary-card.failed .consistency-value {
  color: #f56565;
}

.summary-card.pending .consistency-value {
  color: #ecc94b;
}

.card-description {
  font-size: 12px;
  color: #718096;
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
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: 4px;
}

.stat-row.success {
  border-left: 2px solid #48bb78;
}

.stat-row.error {
  border-left: 2px solid #f56565;
}

.stat-label {
  font-size: 13px;
  color: #a0aec0;
}

.stat-value {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
  font-family: 'SF Mono', Monaco, monospace;
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
