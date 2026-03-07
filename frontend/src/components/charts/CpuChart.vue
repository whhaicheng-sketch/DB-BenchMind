<script setup>
/**
 * CpuChart.vue
 * CPU usage chart component.
 * Green color theme with real-time updates.
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
const currentValue = computed(() => {
  return monitorStore.cpuPercent
})

const dataPoints = computed(() => {
  return monitorStore.cpuHistory
})

// CPU uses green color
const color = '#4CAF50'
</script>

<template>
  <div class="cpu-chart-container">
    <MetricChart
      metric-key="cpu"
      label="CPU"
      :current-value="currentValue"
      :data-points="dataPoints"
      :color="color"
      :height="height"
      :width="width"
      :show-stats="false"
      :show-sparkline="showSparkline"
      unit="%"
    />
  </div>
</template>

<style scoped>
.cpu-chart-container {
  width: 100%;
}
</style>
