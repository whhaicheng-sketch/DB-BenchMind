<script setup>
/**
 * DiskSpaceChart.vue
 * Disk Space chart component.
 * Purple color theme.
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

// Computed values
const currentValue = computed(() => monitorStore.diskUsedPercent)

// Get history data for sparkline
const dataPoints = computed(() => {
  return monitorStore.diskSpaceHistory.map(p => ({
    timestamp: p.timestamp,
    value: p.value || 0
  }))
})

// Format disk space info
const usedGB = computed(() => {
  const gb = monitorStore.diskUsedGB
  if (gb === null || gb === undefined) return 'N/A'
  return gb.toFixed(1) + ' GB'
})

const totalGB = computed(() => {
  const gb = monitorStore.diskTotalGB
  if (gb === null || gb === undefined) return 'N/A'
  return gb.toFixed(1) + ' GB'
})

// Disk Space uses purple color
const color = '#9C27B0'
</script>

<template>
  <div class="diskspace-chart-container">
    <!-- Active chart -->
    <MetricChart
      metric-key="diskSpace"
      label="Disk Space"
      :current-value="currentValue"
      :data-points="dataPoints"
      :color="color"
      :height="height"
      :width="width"
      :show-stats="false"
      :show-sparkline="showSparkline"
      unit="%"
    />
    
    <!-- Additional disk info -->
    <div class="disk-info">
      <span class="disk-used">{{ usedGB }}</span>
      <span class="disk-separator">/</span>
      <span class="disk-total">{{ totalGB }}</span>
    </div>
  </div>
</template>

<style scoped>
.diskspace-chart-container {
  width: 100%;
}

.disk-info {
  display: flex;
  justify-content: center;
  gap: 4px;
  margin-top: 4px;
  font-size: 11px;
  color: #a0aec0;
}

.disk-used {
  color: #9C27B0;
  font-weight: 600;
}

.disk-separator {
  color: #4a5568;
}

.disk-total {
  color: #718096;
}
</style>
