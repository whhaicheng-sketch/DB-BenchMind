<script setup>
/**
 * MainContent.vue
 * Right panel for real-time monitoring charts.
 * Displays TPM/TPS and System charts (CPU, Disk IO, Disk Space).
 */
import { computed, onMounted, onUnmounted } from 'vue'
import { useMonitorStore } from '../../stores/monitor'
import { useBenchmarkStore } from '../../stores/benchmark'
import TpmChart from '../charts/TpmChart.vue'
import TpsChart from '../charts/TpsChart.vue'
import CpuChart from '../charts/CpuChart.vue'
import DiskIOChart from '../charts/DiskIOChart.vue'
import DiskSpaceChart from '../charts/DiskSpaceChart.vue'

// Stores
const monitorStore = useMonitorStore()
const benchmarkStore = useBenchmarkStore()

// Computed - check if benchmark is running
const isRunning = computed(() => {
  return monitorStore.isMonitoring ||
         benchmarkStore.isRunning ||
         benchmarkStore.isPreparing
})

// Computed - system monitoring is always on
const systemMonitoring = computed(() => monitorStore.systemMonitoring)

// Computed - get chart data from store
const tpmData = computed(() => ({
  current: monitorStore.currentTPM,
  history: monitorStore.tpmHistory,
  stats: monitorStore.tpmStats
}))

const tpsData = computed(() => ({
  current: monitorStore.currentTPS,
  history: monitorStore.tpsHistory,
  stats: monitorStore.tpsStats
}))

// Lifecycle
onMounted(async () => {
  // Fetch initial state
  await monitorStore.fetchState()
  // Start system monitoring on mount
  await monitorStore.initSystemMonitoring()
})

onUnmounted(() => {
  // Cleanup is handled by the store
})
</script>

<template>
  <div class="main-content">
    <h2 class="main-title">Real-time Monitor</h2>

    <!-- Running status indicator -->
    <div class="status-indicator" :class="{ active: isRunning }">
      <span class="status-dot"></span>
      <span class="status-text">{{ isRunning ? 'Benchmark Running' : 'No Benchmark Running' }}</span>
    </div>

    <!-- System monitoring status -->
    <div class="system-status" :class="{ active: systemMonitoring }">
      <span class="system-dot"></span>
      <span class="system-text">{{ systemMonitoring ? 'System Monitoring Active' : 'System Monitoring Inactive' }}</span>
    </div>

    <!-- Charts Container -->
    <div class="charts-container">
      <!-- Benchmark Section -->
      <div class="section-header">
        <span class="section-title">Benchmark Metrics</span>
      </div>

      <!-- TPM Chart -->
      <div class="chart-row">
        <TpmChart
          :current-value="tpmData.current"
          :data-points="tpmData.history"
          :stats="tpmData.stats"
          :is-running="isRunning"
        />
      </div>

      <!-- TPS Chart -->
      <div class="chart-row">
        <TpsChart
          :current-value="tpsData.current"
          :data-points="tpsData.history"
          :stats="tpsData.stats"
          :is-running="isRunning"
        />
      </div>

      <!-- System Section -->
      <div class="section-header">
        <span class="section-title">System Metrics</span>
      </div>

      <!-- CPU Chart -->
      <div class="chart-row">
        <CpuChart :show-sparkline="true" />
      </div>

      <!-- Disk IO Chart -->
      <div class="chart-row">
        <DiskIOChart :show-sparkline="true" />
      </div>

      <!-- Disk Space Chart -->
      <div class="chart-row">
        <DiskSpaceChart :show-sparkline="false" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.main-content {
  flex: 1;
  height: 100%;
  background-color: #1a202c;
  padding: 20px;
  overflow-y: auto;
}

.main-title {
  font-size: 18px;
  font-weight: 600;
  color: #e2e8f0;
  margin-bottom: 16px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: #2a3a4a;
  border-radius: 6px;
  margin-bottom: 8px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #718096;
}

.status-indicator.active .status-dot {
  background-color: #48bb78;
  animation: pulse 2s ease-in-out infinite;
}

.status-text {
  font-size: 12px;
  color: #a0aec0;
}

.status-indicator.active .status-text {
  color: #48bb78;
}

.system-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  background-color: #1e2a3a;
  border-radius: 4px;
  margin-bottom: 16px;
}

.system-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: #4a5568;
}

.system-status.active .system-dot {
  background-color: #4caf50;
}

.system-text {
  font-size: 11px;
  color: #718096;
}

.system-status.active .system-text {
  color: #4caf50;
}

.charts-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-header {
  padding: 8px 0;
  border-bottom: 1px solid #2a3a4a;
  margin-top: 8px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #a0aec0;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.chart-row {
  background-color: #2a3a4a;
  border-radius: 8px;
  padding: 12px;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
</style>
