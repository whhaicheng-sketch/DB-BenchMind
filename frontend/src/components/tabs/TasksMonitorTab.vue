<template>
  <div class="tasks-monitor-tab">
    <div class="tab-header">
      <h2>Tasks & Monitor</h2>
      <div class="status" :class="{ running: isRunning }">
        {{ isRunning ? 'Running' : 'Idle' }}
      </div>
    </div>
    
    <div class="monitor-content">
      <!-- Benchmark Controls -->
      <div class="controls">
        <button class="btn btn-primary" :disabled="isRunning" @click="handleRun">
          Run Benchmark
        </button>
        <button class="btn btn-danger" :disabled="!isRunning" @click="handleStop">
          Stop
        </button>
      </div>

      <!-- Real-time Metrics -->
      <div class="metrics-grid">
        <div class="metric-card">
          <div class="metric-label">TPM</div>
          <div class="metric-value">{{ currentTPM }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">TPS</div>
          <div class="metric-value">{{ currentTPS }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">CPU</div>
          <div class="metric-value">{{ cpuUsage }}%</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Disk Read</div>
          <div class="metric-value">{{ diskRead }}</div>
        </div>
      </div>

      <!-- Log Output -->
      <div class="log-panel">
        <h3>Log Output</h3>
        <div class="log-content">
          <div v-for="(line, idx) in logLines" :key="idx" class="log-line">
            {{ line }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useMonitorStore } from '../../stores/monitor'
import { useBenchmarkStore } from '../../stores/benchmark'

const monitorStore = useMonitorStore()
const benchmarkStore = useBenchmarkStore()

const isRunning = computed(() => benchmarkStore.isRunning)
const currentTPM = computed(() => monitorStore.currentTPM || 'N/A')
const currentTPS = computed(() => monitorStore.currentTPS || 'N/A')
const cpuUsage = computed(() => monitorStore.cpuUsage?.toFixed(2) || '0')
const diskRead = computed(() => '0 B/s')
const logLines = computed(() => benchmarkStore.logLines || [])

const handleRun = async () => {
  // TODO: Implement benchmark run
}

const handleStop = async () => {
  await benchmarkStore.stopBenchmark()
}
</script>

<style scoped>
.tasks-monitor-tab {
  height: 100%;
}

.tab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.tab-header h2 {
  font-size: 24px;
  font-weight: 600;
}

.status {
  padding: 8px 16px;
  background-color: #4a5568;
  border-radius: 20px;
  font-size: 14px;
}

.status.running {
  background-color: #48bb78;
}

.monitor-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.controls {
  display: flex;
  gap: 12px;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.metric-card {
  background-color: #2a3a4a;
  padding: 20px;
  border-radius: 8px;
  text-align: center;
}

.metric-label {
  font-size: 14px;
  color: #a0aec0;
  margin-bottom: 8px;
}

.metric-value {
  font-size: 24px;
  font-weight: 600;
}

.log-panel {
  background-color: #2a3a4a;
  border-radius: 8px;
  padding: 16px;
}

.log-panel h3 {
  font-size: 16px;
  margin-bottom: 12px;
}

.log-content {
  max-height: 300px;
  overflow: auto;
  font-family: monospace;
  font-size: 12px;
  background-color: #1e2a3a;
  padding: 12px;
  border-radius: 4px;
}

.log-line {
  padding: 2px 0;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
}

.btn-primary {
  background-color: #4299e1;
  color: white;
}

.btn-danger {
  background-color: #e53e3e;
  color: white;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
