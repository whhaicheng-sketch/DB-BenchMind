<script setup>
/**
 * TpmChart.vue
 * TPM (Transactions Per Minute) chart component.
 * Dynamic color based on CV fluctuation analysis.
 */
import { computed } from 'vue'
import MetricChart from './MetricChart.vue'
import { useMonitorStore } from '../../stores/monitor'
import { useBenchmarkStore } from '../../stores/benchmark'

// Props
const props = defineProps({
  showSparkline: {
    type: Boolean,
    default: true
  },
  height: {
    type: String,
    default: '120px'
  },
  width: {
    type: String,
    default: '100%'
  }
})

// Stores
const monitorStore = useMonitorStore()
const benchmarkStore = useBenchmarkStore()

// Check if benchmark is running
const isRunning = computed(() => {
  return monitorStore.isMonitoring ||
         benchmarkStore.isRunning ||
         benchmarkStore.isPreparing
})

// Computed - show N/A when not running
const currentValue = computed(() => {
  if (!isRunning.value) return null
  return monitorStore.currentTPM
})

const dataPoints = computed(() => {
  if (!isRunning.value) return []
  return monitorStore.tpmHistory
})

const stats = computed(() => {
  if (!isRunning.value) return null
  return monitorStore.tpmStats
})

// Dynamic color based on CV value (T26: Color Alert)
const color = computed(() => monitorStore.tpmColor)

// Default red color for non-running state
const defaultColor = '#DC5050'

// Status info for display
const statusInfo = computed(() => {
  const status = monitorStore.tpmStatus
  const cv = monitorStore.tpmStats.cv
  if (!isRunning.value || cv === 0) {
    return { icon: '', text: '', class: '' }
  }
  if (status === 'stable') {
    return { icon: '●', text: 'Stable', class: 'status-stable' }
  } else if (status === 'fluctuating') {
    return { icon: '◐', text: 'Fluctuating', class: 'status-fluctuating' }
  }
  return { icon: '◉', text: 'Sawtooth', class: 'status-sawtooth' }
})
</script>

<template>
  <div class="tpm-chart-container">
    <!-- Empty state -->
    <div v-if="!isRunning" class="empty-state">
      <div class="empty-label">TPM</div>
      <div class="empty-value">N/A</div>
      <div class="empty-hint">Start a benchmark to see real-time metrics</div>
    </div>

    <!-- Active chart -->
    <div v-else class="active-chart">
      <!-- Status indicator -->
      <div v-if="statusInfo.text" class="status-indicator" :class="statusInfo.class">
        <span class="status-icon">{{ statusInfo.icon }}</span>
        <span class="status-text">{{ statusInfo.text }}</span>
        <span class="status-cv">CV: {{ monitorStore.formattedTPMCV }}</span>
      </div>

      <MetricChart
        metric-key="tpm"
        label="TPM"
        :current-value="currentValue"
        :data-points="dataPoints"
        :color="color"
        :height="height"
        :width="width"
        :show-stats="true"
        :show-sparkline="showSparkline"
        unit="/min"
      />
    </div>
  </div>
</template>

<style scoped>
.tpm-chart-container {
  width: 100%;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px;
  background-color: #2a3a4a;
  border-radius: 8px;
  min-height: 120px;
}

.empty-label {
  font-size: 14px;
  font-weight: 600;
  color: #dc5050;
  margin-bottom: 8px;
}

.empty-value {
  font-size: 24px;
  font-weight: 600;
  color: #718096;
  margin-bottom: 8px;
}

.empty-hint {
  font-size: 12px;
  color: #4a5568;
  font-style: italic;
}

.active-chart {
  position: relative;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 8px;
}

.status-stable {
  background-color: rgba(76, 175, 80, 0.2);
  color: #4CAF50;
}

.status-fluctuating {
  background-color: rgba(255, 193, 7, 0.2);
  color: #FFC107;
}

.status-sawtooth {
  background-color: rgba(244, 67, 54, 0.2);
  color: #F44336;
}

.status-icon {
  font-size: 10px;
}

.status-text {
  font-weight: 600;
}

.status-cv {
  opacity: 0.8;
  margin-left: auto;
}
</style>
