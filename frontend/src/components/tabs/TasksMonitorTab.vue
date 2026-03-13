<template>
  <div class="tasks-monitor-tab">
    <div class="tab-header">
      <div class="header-main">
        <h2>Tasks & Monitor</h2>
        <p class="page-subtitle">Bind templates to connections in a later phase. This page now accepts a template handoff shell from Templates.</p>
      </div>
      <div class="status-group">
        <div class="status" :class="{ running: isRunning }">
          {{ isRunning ? 'Running' : 'Idle' }}
        </div>
      </div>
    </div>

    <div v-if="pendingTaskTemplate" class="template-handoff-card">
      <div class="handoff-head">
        <div>
          <div class="handoff-label">Pending Task Shell</div>
          <h3>{{ pendingTaskTemplate.templateName }}</h3>
        </div>
        <button class="btn btn-secondary" @click="clearPendingTask">Clear</button>
      </div>
      <div class="handoff-grid">
        <div class="handoff-item">
          <span>Template ID</span>
          <strong>{{ pendingTaskTemplate.templateId }}</strong>
        </div>
        <div class="handoff-item">
          <span>Tool</span>
          <strong>{{ pendingTaskTemplate.tool }}</strong>
        </div>
        <div class="handoff-item">
          <span>Database</span>
          <strong>{{ pendingTaskTemplate.dbFamily }}</strong>
        </div>
        <div class="handoff-item">
          <span>Workload</span>
          <strong>{{ pendingTaskTemplate.workloadFamily }}</strong>
        </div>
      </div>
      <p class="handoff-note">Connection binding and real task submission are intentionally deferred. This shell only verifies front-end navigation and state handoff.</p>
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
import { useAppStore } from '../../stores/app'
import { useMonitorStore } from '../../stores/monitor'
import { useBenchmarkStore } from '../../stores/benchmark'

const monitorStore = useMonitorStore()
const benchmarkStore = useBenchmarkStore()
const appStore = useAppStore()

const isRunning = computed(() => benchmarkStore.isRunning)
const currentTPM = computed(() => monitorStore.currentTPM || 'N/A')
const currentTPS = computed(() => monitorStore.currentTPS || 'N/A')
const cpuUsage = computed(() => monitorStore.cpuUsage?.toFixed(2) || '0')
const diskRead = computed(() => '0 B/s')
const logLines = computed(() => benchmarkStore.logLines || [])
const pendingTaskTemplate = computed(() => appStore.pendingTaskTemplate)

const handleRun = async () => {
  // TODO: Implement benchmark run
}

const handleStop = async () => {
  await benchmarkStore.stopBenchmark()
}

const clearPendingTask = () => {
  appStore.clearPendingTaskTemplate()
}
</script>

<style scoped>
.tasks-monitor-tab {
  height: 100%;
}

.tab-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.tab-header h2 {
  font-size: 24px;
  font-weight: 600;
}

.page-subtitle {
  margin-top: 4px;
  font-size: 14px;
  color: #718096;
  max-width: 760px;
}

.status-group {
  display: flex;
  align-items: center;
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

.template-handoff-card {
  margin-bottom: 20px;
  padding: 18px;
  border-radius: 12px;
  border: 1px solid #2d3748;
  background: linear-gradient(135deg, rgba(49, 130, 206, 0.12), rgba(17, 24, 39, 0.96));
}

.handoff-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
  flex-wrap: wrap;
}

.handoff-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #90cdf4;
  margin-bottom: 4px;
}

.handoff-head h3 {
  font-size: 20px;
}

.handoff-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-top: 16px;
}

.handoff-item {
  border-radius: 10px;
  padding: 12px;
  background: rgba(15, 23, 42, 0.85);
  border: 1px solid #1e293b;
}

.handoff-item span {
  display: block;
  font-size: 11px;
  color: #94a3b8;
  margin-bottom: 4px;
}

.handoff-note {
  margin-top: 14px;
  color: #cbd5e0;
  font-size: 13px;
  line-height: 1.6;
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

.btn-secondary {
  background: #1a202c;
  color: #e2e8f0;
  border: 1px solid #3a4a5a;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 980px) {
  .handoff-grid,
  .metrics-grid {
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: 640px) {
  .handoff-grid,
  .metrics-grid {
    grid-template-columns: 1fr;
  }
}
</style>
