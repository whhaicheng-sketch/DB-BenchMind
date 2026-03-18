<script setup>
/**
 * DiskIOChart.vue
 * Disk I/O chart component.
 * Blue color theme.
 */
import { computed } from 'vue'
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
const readBps = computed(() => monitorStore.diskReadBps)
const writeBps = computed(() => monitorStore.diskWriteBps)

// Format bytes per second to human readable
const formatBps = (bps) => {
  if (bps === null || bps === undefined) return 'N/A'
  if (bps >= 1000000000) return (bps / 1000000000).toFixed(2) + ' GB/s'
  if (bps >= 1000000) return (bps / 1000000).toFixed(2) + ' MB/s'
  if (bps >= 1000) return (bps / 1000).toFixed(2) + ' KB/s'
  return bps.toFixed(2) + ' B/s'
}

const readFormatted = computed(() => formatBps(readBps.value))
const writeFormatted = computed(() => formatBps(writeBps.value))

// Get history data for sparkline
const dataPoints = computed(() => {
  return monitorStore.diskIOHistory.map(p => ({
    timestamp: p.timestamp,
    value: p.readBps || 0
  }))
})

// Disk IO uses blue color
const readColor = '#2196F3'
const writeColor = '#1976D2'
</script>

<template>
  <div class="diskio-chart-container">
    <div class="diskio-header">
      <span class="diskio-label">Disk I/O</span>
    </div>
    
    <div class="diskio-content">
      <div class="io-row">
        <span class="io-label">Read:</span>
        <span class="io-value" :style="{ color: readColor }">{{ readFormatted }}</span>
      </div>
      <div class="io-row">
        <span class="io-label">Write:</span>
        <span class="io-value" :style="{ color: writeColor }">{{ writeFormatted }}</span>
      </div>
    </div>

    <div v-if="showSparkline && dataPoints.length > 0" class="sparkline-section">
      <span class="sparkline-label">Trend (60s)</span>
      <div class="sparkline-placeholder">IO Trend Chart</div>
    </div>
  </div>
</template>

<style scoped>
.diskio-chart-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
}

.diskio-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.diskio-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--primary);
}

.diskio-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.io-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.io-label {
  font-size: 12px;
  color: var(--text-muted);
}

.io-value {
  font-size: 14px;
  font-weight: 600;
}

.sparkline-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 8px;
}

.sparkline-label {
  font-size: 10px;
  color: var(--text-muted);
  text-transform: uppercase;
}

.sparkline-placeholder {
  height: 40px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-xs);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  color: var(--text-muted);
}
</style>
