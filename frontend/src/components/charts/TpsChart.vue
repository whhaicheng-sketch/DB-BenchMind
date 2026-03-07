<script setup>
/**
 * TpsChart.vue
 * TPS (Transactions Per Second) chart component.
 * Dynamic color based on CV fluctuation analysis.
 */
import { computed } from 'vue'
import MetricChart from './MetricChart.vue'
import { useMonitorStore } from '../../stores/monitor'

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

// Store
const monitorStore = useMonitorStore()

// Computed
const currentValue = computed(() => monitorStore.currentTPS)
const dataPoints = computed(() => monitorStore.tpsHistory)
const stats = computed(() => monitorStore.tpsStats)
const isRunning = computed(() => monitorStore.isMonitoring)

// Dynamic color based on CV value (T26: Color Alert)
const color = computed(() => monitorStore.tpsColor)

// Default orange color
const defaultColor = '#FFA500'

// Status info for display
const statusInfo = computed(() => {
  const status = monitorStore.tpsStatus
  const cv = monitorStore.tpsStats.cv
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

const label = 'TPS'
const unit = '/sec'
</script>

<template>
  <div class="tps-chart-container">
    <!-- Empty state when not running -->
    <div v-if="!isRunning" class="empty-state">
      <div class="empty-label">TPS</div>
      <div class="empty-value">N/A</div>
      <div class="empty-hint">Start a benchmark to see real-time metrics</div>
    </div>

    <!-- Active chart -->
    <div v-else class="active-chart">
      <!-- Status indicator -->
      <div v-if="statusInfo.text" class="status-indicator" :class="statusInfo.class">
        <span class="status-icon">{{ statusInfo.icon }}</span>
        <span class="status-text">{{ statusInfo.text }}</span>
        <span class="status-cv">CV: {{ monitorStore.formattedTPSCV }}</span>
      </div>

      <MetricChart
        metric-key="tps"
        label="TPS"
        :current-value="currentValue"
        :data-points="dataPoints"
        :color="color"
        :height="height"
        :width="width"
        :show-stats="true"
        :show-sparkline="showSparkline"
        unit="/sec"
      />
    </div>
  </div>
</template>

<style scoped>
.tps-chart-container {
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
  color: #f6ad55;
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
