<template>
  <div class="tasks-monitor-tab">
    <div class="tab-header">
      <div>
        <h2>Performance Analysis</h2>
        <p class="page-subtitle">Template / Connection 选择、Preview、单任务执行、实时指标和 tail 风格日志都在这个工作台完成。</p>
      </div>
      <div class="top-controls">
        <button class="btn btn-primary primary-action" :disabled="!canStart" @click="openPreview">
          Start
        </button>
        <button class="btn btn-danger danger-action" :disabled="!canStop" @click="handleStop">
          Stop
        </button>
        <button class="btn btn-secondary log-action" :disabled="!logViewerTask" @click="openLogViewer">
          Open Log Viewer
        </button>
        <div class="status-entry">
          <button class="status-summary" :class="statusSummary.stateClass" @click="statusPopoverOpen = !statusPopoverOpen">
            <span class="status-summary-badge">{{ statusSummary.label }}</span>
            <span class="status-summary-icon">{{ statusSummary.icon }}</span>
          </button>
          <div v-if="statusPopoverOpen" class="status-popover">
            <div class="status-popover-head">
              <strong>{{ statusPopoverTitle }}</strong>
              <span>{{ statusSummary.message }}</span>
            </div>
            <div class="status-popover-grid">
              <div v-for="item in statusPopoverItems" :key="item.label" class="status-popover-item">
                <span>{{ item.label }}</span>
                <strong>{{ item.value }}</strong>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <section class="panel status-strip" :class="statusStrip.stateClass">
      <div class="status-strip-main">
        <div class="status-strip-phase-block">
          <span class="status-pill" :class="statusStrip.badgeClass">{{ statusStrip.status }}</span>
          <div class="status-strip-phase-copy">
            <span class="status-strip-label">Phase</span>
            <strong class="status-strip-phase">{{ statusStrip.phase }}</strong>
            <span v-if="statusStrip.detail" class="status-strip-detail">{{ statusStrip.detail }}</span>
          </div>
        </div>
        <div class="status-strip-timings">
          <div v-for="item in statusStrip.timings" :key="item.label" class="status-strip-time-chip">
            <span class="status-strip-label">{{ item.label }}</span>
            <strong class="status-strip-time">{{ item.value }}</strong>
          </div>
        </div>
      </div>
      <div class="status-strip-meta" v-if="statusStrip.metaItems.length">
        <div v-for="item in statusStrip.metaItems" :key="item.label" class="status-strip-meta-chip">
          <span class="status-strip-label">{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
        </div>
      </div>
    </section>

    <section v-if="logActionNotice" class="panel action-notice" :class="logActionNotice.tone">
      {{ logActionNotice.message }}
    </section>

    <div class="workspace-grid">
      <div class="left-column">
        <section class="panel card" data-layout-section="create-task">
          <div class="card-head">
            <h3>Create Task</h3>
            <span v-if="pendingTaskTemplate" class="handoff-badge">Template handoff</span>
          </div>

          <label class="field">
            <span>Database Type</span>
            <select v-model="draft.database_type" class="select-dark">
              <option value="">Select database type</option>
              <option v-for="dbType in databaseTypeOptions" :key="dbType.value" :value="dbType.value">
                {{ dbType.label }}
              </option>
            </select>
          </label>

          <label class="field">
            <span>Connection</span>
            <select v-model="draft.connection_id" class="select-dark" :disabled="!draft.database_type">
              <option value="">{{ draft.database_type ? 'Select connection' : 'Select database type first' }}</option>
              <option v-for="connection in filteredConnections" :key="connection.id" :value="connection.id">
                {{ connection.name }}
              </option>
            </select>
          </label>

          <label class="field">
            <span>Template</span>
            <select v-model="draft.template_id" class="select-dark" :disabled="!draft.database_type">
              <option value="">{{ draft.database_type ? 'Select template' : 'Select database type first' }}</option>
              <option v-for="template in filteredTemplates" :key="template.id" :value="template.id">
                {{ template.name }} · {{ template.tool }}
              </option>
            </select>
          </label>

          <label class="field">
            <span>Action</span>
            <select v-model="draft.action" class="select-dark" :disabled="!draft.template_id">
              <option v-for="action in actionOptions" :key="action.value" :value="action.value" :disabled="action.disabled">
                {{ action.label }}
              </option>
            </select>
          </label>

          <div class="override-grid">
            <label v-if="showThreads" class="field">
              <span>{{ concurrencyLabel }}</span>
              <input v-model.number="draft.overrides[concurrencyKey]" type="number" min="1">
            </label>
            <label class="field">
              <span>Run Duration (s)</span>
              <input v-model.number="draft.overrides.duration" type="number" min="1">
            </label>
          </div>

          <p class="compact-help">Use the top Start button to preview and launch the task.</p>
        </section>
      </div>

      <div class="right-column">
        <section class="panel monitor-board">
          <div class="card-head">
            <h3>Monitor Overview</h3>
            <span v-if="!systemEnabled">{{ systemMessage }}</span>
          </div>

          <div class="monitor-board-grid">
            <div class="metric-grid">
              <div class="metric-card" v-for="metric in businessMetrics" :key="metric.label" :data-metric-card="metric.label">
                <div class="metric-card-head">
                  <div class="metric-inline-bar">
                    <span class="metric-title">{{ metric.label }}</span>
                    <div class="metric-stats">
                      <div v-for="item in metric.headerStats" :key="`${metric.label}-${item.label}`">
                        <span>{{ item.label }}</span>
                        <strong>{{ item.value }}</strong>
                      </div>
                    </div>
                  </div>
                  <span class="metric-status" :class="metric.statusClass">{{ metric.statusLabel }}</span>
                </div>
                <div class="metric-history">
                  <div class="chart-shell metric-chart-shell" data-testid="metric-chart-shell">
                    <div class="chart-axis chart-axis-left">
                      <span v-for="tick in metric.ticks" :key="`${metric.label}-${tick.label}`" :style="{ top: `${tick.top}%` }">{{ tick.label }}</span>
                    </div>
                    <div class="chart-canvas metric-chart-canvas" data-testid="metric-chart-canvas">
                      <div v-if="metric.overlay.kind !== 'none'" class="metric-overlay" :class="`metric-overlay-${metric.overlay.kind}`">
                        <strong>{{ metric.overlay.title }}</strong>
                        <span>{{ metric.overlay.body }}</span>
                      </div>
                      <div class="metric-chart-current" :class="{ 'metric-chart-current-muted': metric.overlay.kind !== 'none' }" data-testid="metric-current-value">
                        <strong>{{ metric.current }}</strong>
                        <span>{{ metric.unit }}</span>
                      </div>
                      <svg class="history-chart" viewBox="0 0 320 140" preserveAspectRatio="none">
                        <path :d="metric.areaPath" :fill="metric.fill" />
                        <line v-for="tick in metric.tickLines" :key="`${metric.label}-grid-${tick.label}`" :x1="tick.x1" :x2="tick.x2" :y1="tick.y" :y2="tick.y" class="chart-gridline" />
                        <polyline :points="metric.points" :stroke="metric.glow" :stroke-width="CHART_GLOW_WIDTH" stroke-linecap="round" stroke-linejoin="round" fill="none" class="history-glow metric-line-glow" />
                        <polyline :points="metric.points" :stroke="metric.stroke" :stroke-width="CHART_LINE_WIDTH" stroke-linecap="round" stroke-linejoin="round" fill="none" class="metric-line-main" />
                        <line :x1="metric.plotBounds.x1" :x2="metric.plotBounds.x2" :y1="metric.avgLineY" :y2="metric.avgLineY" class="history-baseline" />
                      </svg>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="system-grid">
              <div class="system-card" :class="{ disabled: !systemEnabled }" data-system-card="CPU">
                <div class="system-card-head">
                  <div class="system-title-block">
                    <span>{{ cpuChart.label }}</span>
                    <strong class="system-summary-lines">
                      <span v-for="(line, index) in cpuChart.summaryRows" :key="`cpu-summary-${index}`" class="summary-line">
                        <span v-for="item in line" :key="item" class="summary-chip">{{ item }}</span>
                      </span>
                    </strong>
                  </div>
                  <div class="chart-legend chart-legend-cpu">
                    <span v-for="line in cpuChart.lines" :key="line.label" class="legend-item">
                      <i :style="{ background: line.color }"></i>
                      {{ line.label }}
                    </span>
                  </div>
                </div>
                <div class="system-chart-wrap" data-system-chart-wrap="CPU">
                  <div class="chart-shell system-chart-shell" data-system-chart-shell="CPU">
                    <div class="chart-axis chart-axis-left chart-axis-cpu">
                      <span v-for="tick in cpuChart.leftTicks" :key="`cpu-${tick.label}`" :style="{ top: `${tick.top}%` }">{{ tick.label }}</span>
                    </div>
                    <div class="chart-canvas">
                      <svg class="system-chart" viewBox="0 0 320 140" preserveAspectRatio="none">
                        <line v-for="tick in cpuChart.leftTicks" :key="`cpu-grid-${tick.label}`" :x1="cpuChart.plotBounds.x1" :x2="cpuChart.plotBounds.x2" :y1="tick.y" :y2="tick.y" class="chart-gridline" />
                        <polyline
                          v-for="line in cpuChart.lines"
                          :key="line.label"
                          :points="line.points"
                          :stroke="line.color"
                          :stroke-width="CHART_LINE_WIDTH"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          fill="none"
                        />
                      </svg>
                    </div>
                  </div>
                </div>
                <p class="system-caption">{{ cpuChart.caption }}</p>
              </div>

              <div class="system-card" :class="{ disabled: !systemEnabled }" data-system-card="Disk IO">
                <div class="system-card-head">
                  <div class="system-title-block">
                    <span>{{ diskChart.label }}</span>
                    <strong class="system-summary-lines">
                      <span v-for="(line, index) in diskChart.summaryRows" :key="`disk-summary-${index}`" class="summary-line">
                        <span v-for="item in line" :key="item" class="summary-chip">{{ item }}</span>
                      </span>
                    </strong>
                  </div>
                  <div class="chart-legend chart-legend-single">
                    <span v-for="line in diskChart.lines" :key="line.label" class="legend-item">
                      <i :style="{ background: line.color }"></i>
                      {{ line.label }}
                    </span>
                  </div>
                </div>
                <div class="system-chart-wrap disk-chart-combined" data-system-chart-wrap="Disk IO">
                  <div class="chart-shell system-chart-shell disk-chart-shell" data-system-chart-shell="Disk IO">
                    <div class="chart-axis chart-axis-left chart-axis-bandwidth">
                      <span v-for="tick in diskChart.leftTicks" :key="`disk-left-${tick.label}`" :style="{ top: `${tick.top}%` }">{{ tick.label }}</span>
                    </div>
                    <div class="chart-canvas">
                      <svg class="system-chart" viewBox="0 0 320 140" preserveAspectRatio="none">
                        <line v-for="tick in diskChart.leftTicks" :key="`disk-grid-${tick.label}`" :x1="diskChart.plotBounds.x1" :x2="diskChart.plotBounds.x2" :y1="tick.y" :y2="tick.y" class="chart-gridline" />
                        <polyline
                          v-for="line in diskChart.lines"
                          :key="line.label"
                          :points="line.points"
                          :stroke="line.color"
                          :stroke-dasharray="line.dasharray || ''"
                          :stroke-width="CHART_LINE_WIDTH"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          fill="none"
                        />
                      </svg>
                    </div>
                    <div class="chart-axis chart-axis-right chart-axis-latency">
                      <span v-for="tick in diskChart.rightTicks" :key="`disk-right-${tick.label}`" :style="{ top: `${tick.top}%` }">{{ tick.label }}</span>
                    </div>
                  </div>
                </div>
                <p class="system-caption">{{ diskChart.caption }}</p>
              </div>
            </div>

          </div>
        </section>
      </div>
    </div>

    <div v-if="previewOpen" class="modal-backdrop" @click.self="closePreview">
      <div class="modal">
        <div class="card-head">
          <h3>Task Preview</h3>
          <button class="text-action" @click="closePreview">Close</button>
        </div>
        <div class="preview-meta-grid" v-if="previewTask">
          <div class="preview-meta-card">
            <span class="eyebrow">Template</span>
            <strong>{{ previewTask.template_snapshot.name }}</strong>
            <p>{{ previewTask.template_snapshot.tool }} · {{ previewTask.template_snapshot.db_family }}</p>
          </div>
          <div class="preview-meta-card">
            <span class="eyebrow">Connection</span>
            <strong>{{ previewTask.connection_snapshot.name }}</strong>
            <p>{{ previewTask.connection_snapshot.type }} · {{ previewTask.connection_snapshot.host }}</p>
          </div>
          <div class="preview-meta-card">
            <span class="eyebrow">Action</span>
            <strong>{{ actionLabel(previewTask.action) }}</strong>
            <p>{{ previewTask.name }}</p>
          </div>
          <div class="preview-meta-card">
            <span class="eyebrow">Monitoring</span>
            <strong>{{ previewTask.readiness?.ssh_available ? 'SSH metrics enabled' : 'SSH unavailable' }}</strong>
            <p>{{ previewTask.readiness?.ssh_message }}</p>
          </div>
        </div>
        <section class="preview-params-section">
          <span class="eyebrow">Resolved Parameters</span>
          <pre class="params-preview">{{ JSON.stringify(previewTask?.resolved_params || {}, null, 2) }}</pre>
        </section>
        <div class="actions modal-actions">
          <button class="btn btn-secondary" @click="closePreview">Back</button>
          <button v-if="previewConfirmable" class="btn btn-primary" :disabled="!previewTask?.readiness?.db_valid" @click="confirmCreateTask">Confirm</button>
        </div>
      </div>
    </div>

    <div v-if="logViewerOpen" class="modal-backdrop" @click.self="closeLogViewer">
      <div class="modal modal-wide">
        <div class="card-head">
          <h3>Task Logs</h3>
          <button class="text-action" @click="closeLogViewer">Close</button>
        </div>
        <div class="log-toolbar">
          <input v-model="logQuery" type="text" placeholder="Search log lines">
          <select v-model="logPhase" class="select-dark">
            <option value="">All phases</option>
            <option value="prepare">prepare</option>
            <option value="run">run</option>
            <option value="cleanup">cleanup</option>
          </select>
          <button class="btn btn-secondary" @click="autoScroll = !autoScroll">
            {{ autoScroll ? 'Pause Auto Scroll' : 'Resume Auto Scroll' }}
          </button>
        </div>
        <p class="log-meta" v-if="currentTask">Disk logs: {{ logPathSummary }}</p>
        <div ref="logViewport" class="terminal-viewer">
          <div v-for="(line, index) in taskStore.logLines" :key="`${line.timestamp}-${index}`" class="terminal-line" :class="lineClass(line)">
            <div class="terminal-line-badges">
              <span class="phase-tag">{{ line.phase }}</span>
              <span class="stream-tag">{{ line.stream }}</span>
            </div>
            <span class="terminal-line-message">{{ line.content }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '../../stores/app'
import { useConnectionStore } from '../../stores/connection'
import { useTaskStore } from '../../stores/task'
import { useTemplateStore } from '../../stores/template'
import { getOracleSwingbenchMetricOverlayState } from './tasksMonitorOracleSwingbenchState.mjs'
import { resolveTasksMonitorBinding } from './tasksMonitorTaskState.mjs'
import { getPreferredTemplateId } from './tasksMonitorTemplateSelection.mjs'
import { buildStatusStripModel } from './tasksMonitorStatusStrip.mjs'

const appStore = useAppStore()
const templateStore = useTemplateStore()
const connectionStore = useConnectionStore()
const taskStore = useTaskStore()

const { templates } = storeToRefs(templateStore)
const { connections } = storeToRefs(connectionStore)
const { currentTask, activeTask } = storeToRefs(taskStore)
const TASKS_DRAFT_STORAGE_KEY = 'db-benchmind.tasks-monitor.draft.v1'
const CHART_LINE_WIDTH = 1.4
const CHART_GLOW_WIDTH = 2.2
const METRIC_PLOT_BOUNDS = { x1: 38, x2: 316, y1: 16, y2: 130 }
const SYSTEM_PLOT_BOUNDS = { x1: 38, x2: 316, y1: 20, y2: 126 }
const DISK_PLOT_BOUNDS = { x1: 38, x2: 282, y1: 20, y2: 126 }

const draft = reactive({
  database_type: '',
  template_id: '',
  connection_id: '',
  action: 'full_pipeline',
  overrides: {
    duration: 60,
    threads: 8,
    virtual_users: 8
  }
})

const previewOpen = ref(false)
const previewTask = ref(null)
const previewConfirmable = ref(false)
const statusPopoverOpen = ref(false)
const logViewerOpen = ref(false)
const pendingStopTaskId = ref('')
const logQuery = ref('')
const logPhase = ref('')
const autoScroll = ref(true)
const logViewport = ref(null)
const nowTick = ref(Date.now())
const logActionNotice = ref(null)
let logNoticeTimer = null

const selectedTemplate = computed(() => templates.value.find((template) => template.id === draft.template_id) || null)
const selectedConnection = computed(() => connections.value.find((connection) => connection.id === draft.connection_id) || null)
const pendingTaskTemplate = computed(() => appStore.pendingTaskTemplate)
const databaseTypeOptions = computed(() => {
  const labelMap = connectionStore.typeLabels || {}
  const types = connectionStore.availableTypes?.length
    ? connectionStore.availableTypes
    : Object.keys(labelMap)
  return types.map((type) => ({ value: type, label: labelMap[type] || type }))
})
const filteredConnections = computed(() => {
  if (!draft.database_type) return []
  return connections.value.filter((connection) => connection.type === draft.database_type)
})
const filteredTemplates = computed(() => {
  if (!draft.database_type) return []
  return templateStore.templatesByDatabase(draft.database_type)
    .filter((template) => template.dbFamily === draft.database_type || template.database_types?.includes(draft.database_type))
})

const concurrencyKey = computed(() => {
  if (selectedTemplate.value?.tool === 'hammerdb' || selectedTemplate.value?.tool === 'swingbench') return 'virtual_users'
  return 'threads'
})

const concurrencyLabel = computed(() => concurrencyKey.value === 'threads' ? 'Threads' : 'Virtual Users')
const showThreads = computed(() => !!selectedTemplate.value)
const actionOptions = computed(() => [
  { value: 'prepare', label: 'Prepare', disabled: false },
  { value: 'run', label: 'Run', disabled: false },
  { value: 'cleanup', label: 'Cleanup', disabled: false },
  { value: 'full_pipeline', label: 'Run Full Flow (Prepare -> Run -> Cleanup)', disabled: false }
])
const selectedActionOption = computed(() => actionOptions.value.find((action) => action.value === draft.action) || null)
const canPreview = computed(() => !!draft.database_type && !!draft.template_id && !!draft.connection_id && selectedActionOption.value && !selectedActionOption.value.disabled)
const canStart = computed(() => canPreview.value && !taskBinding.value.startBlocked)
const taskBinding = computed(() => resolveTasksMonitorBinding({
  tasks: taskStore.tasks,
  draft,
  pendingStopTaskId: pendingStopTaskId.value
}))
const canStop = computed(() => taskBinding.value.stopEnabled)
const logViewerTask = computed(() => taskBinding.value.logViewerTask)
const taskNameStamp = ref('')
const autoGeneratedTaskNameReady = computed(() => !!selectedTemplate.value && !!selectedConnection.value)
const autoGeneratedTaskName = computed(() => {
  if (!autoGeneratedTaskNameReady.value) {
    return 'Auto-generated after selection'
  }
  return [
    sanitizeName(selectedTemplate.value.name),
    sanitizeName(selectedConnection.value.name),
    draft.action,
    taskNameStamp.value
  ].filter(Boolean).join('-')
})

const readiness = computed(() => taskStore.readiness || {})
const readinessItems = computed(() => [
  { label: 'Template selected', ok: !!readiness.value.template_selected, message: readiness.value.template_selected ? selectedTemplate.value?.name || 'Ready' : 'Select a template' },
  { label: 'Connection selected', ok: !!readiness.value.connection_selected, message: readiness.value.connection_selected ? selectedConnection.value?.name || 'Ready' : 'Select a connection' },
  { label: 'Action supported', ok: !!readiness.value.action_supported, message: readiness.value.action_supported ? draft.action : 'Template does not support this action' },
  { label: 'Runtime overrides valid', ok: !!readiness.value.runtime_valid, message: readiness.value.runtime_valid ? 'Ready' : 'Check overrides' },
  { label: 'DB connection valid', ok: !!readiness.value.db_valid, message: readiness.value.db_message || 'Not checked' },
  { label: 'SSH monitoring', ok: !!readiness.value.ssh_available, message: readiness.value.ssh_message || 'SSH required' }
])
const draftBlockers = computed(() => readinessItems.value.filter((item) => !item.ok).map((item) => item.message))
const latestWarningOrError = computed(() => {
  const lines = currentTask.value?.log_tail || []
  for (let index = lines.length - 1; index >= 0; index -= 1) {
    const line = lines[index]
    if (/error|failed|fatal|warn/i.test(line.content || '')) {
      return line.content
    }
  }
  return currentTask.value?.error_message || 'None'
})
const statusSummary = computed(() => {
  if (activeTask.value) {
    if (activeTask.value.readiness?.ssh_available === false) {
      return { label: 'Warning', icon: '!', stateClass: 'warning', message: activeTask.value.readiness?.ssh_message || 'SSH unavailable, benchmark continues' }
    }
    return { label: 'Ready', icon: 'i', stateClass: 'ready', message: `${statusLabel(activeTask.value.status)} task in progress` }
  }
  if (!canPreview.value || draftBlockers.value.length > 0) {
    return { label: 'Blocked', icon: '!', stateClass: 'blocked', message: draftBlockers.value[0] || 'Complete required selections before starting' }
  }
  if (readiness.value.ssh_available === false) {
    return { label: 'Warning', icon: '!', stateClass: 'warning', message: readiness.value.ssh_message || 'SSH unavailable, benchmark continues without system metrics' }
  }
  return { label: 'Ready', icon: 'i', stateClass: 'ready', message: 'Draft is ready to start' }
})
const statusPopoverTitle = computed(() => activeTask.value ? 'Current Task Details' : 'Draft Readiness')
const statusPopoverItems = computed(() => {
  if (activeTask.value) {
    return [
      { label: 'Task Name', value: activeTask.value.name },
      { label: 'Template', value: activeTask.value.template_snapshot?.name || 'N/A' },
      { label: 'Connection', value: activeTask.value.connection_snapshot?.name || 'N/A' },
      { label: 'Database Type', value: activeTask.value.connection_snapshot?.type || 'N/A' },
      { label: 'Action', value: actionLabel(activeTask.value.action) },
      { label: 'Current Phase', value: activeTask.value.current_phase || 'none' },
      { label: 'Started At', value: formatDate(activeTask.value.started_at || activeTask.value.created_at) },
      { label: 'Requested Run Duration', value: requestedRunDurationLabel.value },
      { label: 'Prepare', value: phaseTiming.value.prepare },
      { label: 'Run', value: phaseTiming.value.run },
      { label: 'Total', value: phaseTiming.value.total },
      { label: 'DB Validation', value: activeTask.value.readiness?.db_message || 'N/A' },
      { label: 'SSH Status', value: activeTask.value.readiness?.ssh_message || 'N/A' },
      { label: 'Disk Log Path', value: logPathSummary.value },
      { label: 'Recent Warning/Error', value: latestWarningOrError.value }
    ]
  }
  return [
    { label: 'Database Type', value: draft.database_type || 'Not selected' },
    { label: 'Connection', value: selectedConnection.value?.name || 'Not selected' },
    { label: 'Template', value: selectedTemplate.value?.name || 'Not selected' },
    { label: 'Action', value: actionLabel(draft.action) },
    { label: 'Runtime Overrides', value: runtimeOverridesSummary() },
    { label: 'DB Validation', value: readiness.value.db_message || 'Not checked' },
    { label: 'SSH Status', value: readiness.value.ssh_message || 'Not checked' },
    { label: 'Can Start', value: canStart.value ? 'Yes' : 'No' },
    { label: 'Blocker', value: draftBlockers.value[0] || 'None' }
  ]
})

const businessMetrics = computed(() => {
  const metrics = currentTask.value?.metrics || {}
  return [
    metricCard(currentTask.value, 'TPS', metrics.tps, '/sec', '#f4a261', 'rgba(244, 162, 97, 0.07)'),
    metricCard(currentTask.value, 'TPM', metrics.tpm, '/min', '#4dd0a8', 'rgba(77, 208, 168, 0.065)')
  ]
})

const systemEnabled = computed(() => !!currentTask.value?.metrics?.system_enabled)
const systemMessage = computed(() => currentTask.value?.metrics?.system_message || 'SSH unavailable, benchmark continues without system metrics')
const cpuChart = computed(() => {
  const metrics = currentTask.value?.metrics || {}
  return multiMetricChart('CPU', [
    { label: 'USER', metric: metrics.cpu_user, color: '#34d399', formatter: formatPercent },
    { label: 'SYS', metric: metrics.cpu_sys, color: '#f87171', formatter: formatPercent },
    { label: 'IOWAIT', metric: metrics.cpu_iowait, color: '#60a5fa', formatter: formatPercent },
    { label: 'ST', metric: metrics.cpu_steal, color: '#fbbf24', formatter: formatPercent }
  ], 'Recent CPU composition across user, sys, iowait, and steal time.')
})

const diskChart = computed(() => diskMetricChart(currentTask.value?.metrics || {}))

const elapsedLabel = computed(() => formatElapsed(currentTask.value?.started_at, currentTask.value?.completed_at, nowTick.value))
const phaseTiming = computed(() => ({
  prepare: formatDurationMs(currentTask.value?.timing?.prepare_ms || 0),
  run: formatDurationMs(currentTask.value?.timing?.run_ms || 0),
  total: formatDurationMs(currentTask.value?.timing?.total_ms || 0)
}))
const requestedRunDurationLabel = computed(() => formatDurationMs(currentTask.value?.timing?.run_duration_input_ms || 0))
const logPathSummary = computed(() => {
  const paths = Object.entries(currentTask.value?.run_log_paths || {})
    .filter(([, path]) => !!path)
    .map(([phase, path]) => `${phase}: ${path}`)
  return paths.length ? paths.join(' | ') : 'pending'
})
const statusStrip = computed(() => {
  return buildStatusStripModel(currentTask.value, phaseTiming.value, requestedRunDurationLabel.value, {
    statusLabel,
    phaseLabel,
    isActive
  })
})

let validationTimer = null
let clockTimer = null

watch(
  () => [draft.database_type, draft.template_id, draft.connection_id, draft.action, draft.overrides.duration, draft.overrides.threads, draft.overrides.virtual_users],
  () => {
    if (!draft.template_id && !draft.connection_id) return
    clearTimeout(validationTimer)
    validationTimer = setTimeout(() => {
      refreshValidation()
    }, 250)
  },
  { immediate: true }
)

watch([logQuery, logPhase], async () => {
  if (!logViewerOpen.value || !currentTask.value) return
  await taskStore.fetchLogs({ taskId: currentTask.value.id, query: logQuery.value, phase: logPhase.value })
  if (autoScroll.value) scrollLogsToBottom()
})

watch(
  () => taskStore.logLines.length,
  async () => {
    if (logViewerOpen.value && autoScroll.value) {
      await nextTick()
      scrollLogsToBottom()
    }
  }
)

watch(activeTask, (task) => {
  pendingStopTaskId.value = taskBinding.value.pendingStopTaskId
})

onMounted(async () => {
  await Promise.all([templateStore.loadTemplates?.() || templateStore.fetchTemplates?.() || Promise.resolve(), connectionStore.fetchConnections(), taskStore.fetchTasks()])
  restoreDraft()
  sanitizeDraftState()
  applyPreferredTemplateSelection()
  taskStore.startPolling()
  clockTimer = setInterval(() => {
    nowTick.value = Date.now()
  }, 1000)
  if (pendingTaskTemplate.value?.templateId) {
    const handoffTemplate = templates.value.find((template) => template.id === pendingTaskTemplate.value.templateId)
    if (handoffTemplate) {
      draft.database_type = handoffTemplate.dbFamily || handoffTemplate.database_types?.[0] || ''
      draft.template_id = handoffTemplate.id
    }
    appStore.clearPendingTaskTemplate()
  }
  sanitizeDraftState()
})

onBeforeUnmount(() => {
  taskStore.stopPolling()
  clearTimeout(validationTimer)
  if (clockTimer) clearInterval(clockTimer)
  if (logNoticeTimer) clearTimeout(logNoticeTimer)
})

watch(
  () => [draft.database_type, draft.connection_id, draft.template_id, draft.action],
  () => {
    taskNameStamp.value = buildTaskNameStamp()
  },
  { immediate: true }
)

watch(
  () => [
    draft.database_type,
    draft.connection_id,
    draft.template_id,
    draft.action,
    draft.overrides.duration,
    draft.overrides.threads,
    draft.overrides.virtual_users
  ],
  () => {
    persistDraft()
  },
  { deep: true }
)

watch(() => draft.database_type, (databaseType) => {
  if (!databaseType) {
    draft.connection_id = ''
    draft.template_id = ''
    return
  }
  if (selectedConnection.value && selectedConnection.value.type !== databaseType) {
    draft.connection_id = ''
  }
  if (selectedTemplate.value && !(selectedTemplate.value.dbFamily === databaseType || selectedTemplate.value.database_types?.includes(databaseType))) {
    draft.template_id = ''
  }
  applyPreferredTemplateSelection({ preferTestTemplate: true })
})

watch(() => draft.connection_id, () => {
  if (selectedTemplate.value && !filteredTemplates.value.some((template) => template.id === selectedTemplate.value.id)) {
    draft.template_id = ''
  }
  applyPreferredTemplateSelection({ preferTestTemplate: true })
})

watch(actionOptions, (options) => {
  if (!options.length) return
  const current = options.find((option) => option.value === draft.action)
  if (!current || current.disabled) {
    const nextAllowed = options.find((option) => !option.disabled)
    draft.action = nextAllowed?.value || 'full_pipeline'
  }
})

async function refreshValidation() {
  if (!canPreview.value) return
  await taskStore.validateDraft(payloadFromDraft())
}

async function openPreview() {
  if (!canStart.value) return
  const result = await taskStore.validateDraft(payloadFromDraft())
  if (result.task) {
    previewTask.value = result.task
    previewConfirmable.value = true
    previewOpen.value = true
  }
}

async function confirmCreateTask() {
  const result = await taskStore.createTask(payloadFromDraft())
  if (!result.error) {
    closePreview()
    await taskStore.fetchTasks()
  }
}

async function handleStop() {
  if (!taskBinding.value.stopTaskId) return
  if (taskBinding.value.stopInFlight) return
  pendingStopTaskId.value = taskBinding.value.stopTaskId
  try {
    const result = await taskStore.stopTask(taskBinding.value.stopTaskId)
    if (result?.error) {
      pendingStopTaskId.value = ''
    }
  } finally {
    pendingStopTaskId.value = resolveTasksMonitorBinding({
      tasks: taskStore.tasks,
      draft,
      pendingStopTaskId: pendingStopTaskId.value
    }).pendingStopTaskId
  }
}

async function openLogViewer() {
  if (!logViewerTask.value) return
  logViewerOpen.value = true
  const lines = await taskStore.fetchLogs({ taskId: logViewerTask.value.id, query: logQuery.value, phase: logPhase.value })
  if (!lines.length && taskStore.error) {
    setLogActionNotice(taskStore.error, 'error')
  }
  await nextTick()
  scrollLogsToBottom()
}

function payloadFromDraft() {
  return {
    task_name: autoGeneratedTaskNameReady.value ? autoGeneratedTaskName.value : '',
    database_type: draft.database_type,
    template_id: draft.template_id,
    connection_id: draft.connection_id,
    action: draft.action,
    preview_token: previewTask.value?.preview_token || '',
    overrides: {
      duration: Number(draft.overrides.duration || 0),
      [concurrencyKey.value]: Number(draft.overrides[concurrencyKey.value] || 0)
    }
  }
}

function actionLabel(action) {
  return actionOptions.value.find((option) => option.value === action)?.label || action || 'Not selected'
}

function metricCard(task, label, metric = {}, unit, stroke, fill) {
  const current = Number(metric.current || 0)
  const avg = Number(metric.avg || 0)
  const max = Number(metric.max || 0)
  const series = Array.isArray(metric.series) ? metric.series : []
  const plotBounds = METRIC_PLOT_BOUNDS
  const trend = buildTrendData(series, 320, 140, null, plotBounds)
  const ticks = buildAxisTicks(trend.min, trend.max, formatCompactNumber, 140, plotBounds)
  const overlay = getOracleSwingbenchMetricOverlayState(task, label)
  const showOverlay = overlay.kind !== 'none'
  const overlayLabel = overlay.kind === 'prepare'
    ? 'Prepare'
    : overlay.kind === 'run-waiting'
      ? 'Waiting'
      : null
  const metricStateLabel = showOverlay ? overlayLabel : metricStatus(metric)
  return {
    label,
    current: showOverlay ? '0' : formatNumber(current),
    avg: formatNumber(avg),
    max: formatNumber(max),
    headerStats: showOverlay
      ? [{ label: overlay.kind === 'prepare' ? 'PHASE' : 'STATE', value: overlayLabel }]
      : [
          { label: 'AVG', value: formatNumber(avg) },
          { label: 'MAX', value: formatNumber(max) }
        ],
    unit,
    stroke,
    glow: withAlpha(stroke, 0.11),
    fill: withAlpha(stroke, label === 'TPS' ? 0.05 : 0.045),
    points: trend.points,
    areaPath: trend.areaPath,
    avgLineY: mapSeriesValueToY(avg, trend.min, trend.max, 140, plotBounds),
    ticks,
    tickLines: ticks.map((tick) => ({ ...tick, x1: plotBounds.x1, x2: plotBounds.x2 })),
    plotBounds,
    statusLabel: metricStateLabel,
    statusClass: `metric-${metricStateLabel.toLowerCase()}`,
    overlay
  }
}

function multiMetricChart(label, definitions, caption) {
  const values = definitions.map((item) => Number(item.metric?.current || 0))
  const allSeriesValues = definitions.flatMap((item) => (item.metric?.series || []).map((point) => Number(point.value || 0)))
  const min = label === 'CPU' ? 0 : Math.min(...allSeriesValues, 0)
  const max = label === 'CPU' ? Math.max(100, ...allSeriesValues, 1) : Math.max(...allSeriesValues, 1)
  const plotBounds = SYSTEM_PLOT_BOUNDS
  const leftAxisFormatter = label === 'CPU' ? formatAxisPercent : (definitions[0]?.formatter || formatCompactNumber)
  const leftTicks = buildAxisTicks(min, max, leftAxisFormatter, 140, plotBounds, label === 'CPU' ? 5 : 4)
  return {
    label,
    summaryRows: [definitions.map((item, index) => `${item.label} ${item.formatter(values[index])}`)],
    caption: label === 'CPU'
      ? 'Percent scale stays on the left axis while USER, SYS, and IOWAIT keep their color-coded traces.'
      : caption,
    plotBounds,
    leftTicks,
    lines: definitions.map((item) => ({
      label: item.label,
      color: item.color,
      points: buildTrend(item.metric?.series || [], 320, 140, { min, max }, plotBounds)
    }))
  }
}

function diskMetricChart(metrics = {}) {
  const bandwidthDefs = [
    { label: 'READ', metric: metrics.disk_read_bps, color: '#60a5fa', formatter: formatBytes, axis: 'left' },
    { label: 'WRITE', metric: metrics.disk_write_bps, color: '#f59e0b', formatter: formatBytes, axis: 'left' }
  ]
  const latencyDefs = [
    { label: 'R_LAT', metric: metrics.disk_read_latency_ms, color: '#22c55e', formatter: formatLatency, axis: 'right', dasharray: '7 4' },
    { label: 'W_LAT', metric: metrics.disk_write_latency_ms, color: '#f472b6', formatter: formatLatency, axis: 'right', dasharray: '7 4' }
  ]
  const allDefs = [...bandwidthDefs, ...latencyDefs]
  const leftValues = bandwidthDefs.flatMap((item) => (item.metric?.series || []).map((point) => Number(point.value || 0)))
  const rightValues = latencyDefs.flatMap((item) => (item.metric?.series || []).map((point) => Number(point.value || 0)))
  const leftBounds = { min: 0, max: Math.max(...leftValues, 1) }
  const rightBounds = { min: 0, max: Math.max(...rightValues, 1) }
  const plotBounds = DISK_PLOT_BOUNDS
  return {
    label: 'Disk IO',
    summaryRows: [[
      `read ${formatBytes(metrics.disk_read_bps?.current)}`,
      `write ${formatBytes(metrics.disk_write_bps?.current)}`,
      `r_lat ${formatLatency(metrics.disk_read_latency_ms?.current)}`,
      `w_lat ${formatLatency(metrics.disk_write_latency_ms?.current)}`
    ]],
    caption: 'Left axis serves READ/WRITE bandwidth, right axis serves R_LAT/W_LAT milliseconds.',
    plotBounds,
    leftTicks: buildAxisTicks(leftBounds.min, leftBounds.max, formatAxisBytes, 140, plotBounds, 4),
    rightTicks: buildAxisTicks(rightBounds.min, rightBounds.max, formatAxisLatency, 140, plotBounds, 4),
    lines: allDefs.map((item) => ({
      label: item.label,
      color: item.color,
      dasharray: item.dasharray,
      points: buildTrend(item.metric?.series || [], 320, 140, item.axis === 'left' ? leftBounds : rightBounds, plotBounds)
    }))
  }
}

function buildTrend(series = [], width, height, bounds = null, plotBounds = null) {
  return buildTrendData(series, width, height, bounds, plotBounds).points
}

function buildTrendData(series = [], width, height, bounds = null, plotBounds = null) {
  const areaBounds = plotBounds || { x1: 0, x2: width, y1: 5, y2: height - 5 }
  if (!series?.length) {
    return { points: '', areaPath: '', min: bounds?.min ?? 0, max: bounds?.max ?? 1 }
  }
  const values = series.map((point) => Number(point.value || 0))
  const max = bounds?.max ?? Math.max(...values, 1)
  const min = bounds?.min ?? Math.min(...values, 0)
  const points = values.map((value, index) => {
    const x = series.length === 1 ? areaBounds.x1 : areaBounds.x1 + (index / (series.length - 1)) * (areaBounds.x2 - areaBounds.x1)
    const y = mapSeriesValueToY(value, min, max, height, areaBounds)
    return `${x},${y}`
  }).join(' ')
  const areaPath = `M ${areaBounds.x1},${areaBounds.y2} L ${points.split(' ').join(' L ')} L ${areaBounds.x2},${areaBounds.y2} Z`
  return { points, areaPath, min, max }
}

function mapSeriesValueToY(value, min, max, height, plotBounds = null) {
  const bounds = plotBounds || { y1: 5, y2: height - 5 }
  const range = Math.max(max - min, 1)
  return bounds.y2 - ((value - min) / range) * (bounds.y2 - bounds.y1)
}

function metricStatus(metric = {}) {
  const cv = Number(metric.cv || 0)
  if (cv >= 0.1) return 'Sawtooth'
  if (cv >= 0.05) return 'Fluctuating'
  return 'Stable'
}

function lineClass(line) {
  return {
    error: /error|failed|fatal/i.test(line.content || ''),
    muted: line.stream === 'info' || line.stream === 'event'
  }
}

function statusLabel(status) {
  const labels = {
    idle: 'Idle',
    queued: 'Queued',
    starting: 'Starting',
    preparing: 'Preparing',
    running: 'Running',
    cleaning: 'Cleaning',
    stopping: 'Stopping',
    stopped: 'Stopped',
    success: 'Success',
    failed: 'Failed',
    cancelled: 'Cancelled'
  }
  return labels[status] || status
}

function isActive(status) {
  return ['starting', 'preparing', 'running', 'cleaning', 'stopping'].includes(status)
}

function phaseLabel(phase, status) {
  if (['success', 'failed', 'stopped', 'cancelled'].includes(status)) {
    return 'finished'
  }
  return phase && phase !== 'none' ? phase : 'none'
}

function formatNumber(value) {
  return Number(value || 0).toLocaleString('en-US', { maximumFractionDigits: 1 })
}

function formatCompactNumber(value) {
  const abs = Math.abs(Number(value || 0))
  if (abs >= 1000000) return `${(value / 1000000).toFixed(1)}M`
  if (abs >= 1000) return `${(value / 1000).toFixed(1)}K`
  return Number(value || 0).toLocaleString('en-US', { maximumFractionDigits: 0 })
}

function formatBytes(value) {
  const units = ['B/s', 'KB/s', 'MB/s', 'GB/s']
  let current = value || 0
  let index = 0
  while (current >= 1024 && index < units.length - 1) {
    current /= 1024
    index += 1
  }
  return `${current.toFixed(1)} ${units[index]}`
}

function formatStorage(value) {
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let current = value || 0
  let index = 0
  while (current >= 1024 && index < units.length - 1) {
    current /= 1024
    index += 1
  }
  return `${current.toFixed(1)} ${units[index]}`
}

function formatPercent(value) {
  return `${Number(value || 0).toFixed(1)}%`
}

function formatAxisPercent(value) {
  return `${Math.round(Number(value || 0))}%`
}

function formatLatency(value) {
  return `${Number(value || 0).toFixed(1)} ms`
}

function formatAxisLatency(value) {
  return `${Number(value || 0).toFixed(1)} ms`
}

function formatAxisBytes(value) {
  const units = ['B/s', 'KB/s', 'MB/s', 'GB/s']
  let current = Number(value || 0)
  let index = 0
  while (current >= 1024 && index < units.length - 1) {
    current /= 1024
    index += 1
  }
  return `${current.toFixed(current >= 10 ? 0 : 1)} ${units[index]}`
}

function withAlpha(hexColor, alpha) {
  if (!hexColor?.startsWith('#') || (hexColor.length !== 7 && hexColor.length !== 4)) {
    return hexColor
  }
  let normalized = hexColor
  if (hexColor.length === 4) {
    normalized = `#${hexColor[1]}${hexColor[1]}${hexColor[2]}${hexColor[2]}${hexColor[3]}${hexColor[3]}`
  }
  const r = parseInt(normalized.slice(1, 3), 16)
  const g = parseInt(normalized.slice(3, 5), 16)
  const b = parseInt(normalized.slice(5, 7), 16)
  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}

function buildAxisTicks(min, max, formatter, height = 116, plotBounds = null, tickCount = 4) {
  const bounds = plotBounds || { y1: 5, y2: height - 5 }
  const safeTickCount = Math.max(2, tickCount)
  const range = Math.max(max - min, 1)
  const rawValues = Array.from({ length: safeTickCount }, (_, index) => {
    const ratio = index / (safeTickCount - 1)
    return max - ratio * range
  })
  const seen = new Set()
  return rawValues.filter((value) => {
    const label = formatter(value)
    if (seen.has(label)) return false
    seen.add(label)
    return true
  }).map((value) => {
    const y = mapSeriesValueToY(value, min, max, height, bounds)
    const top = (y / Math.max(height, 1)) * 100
    return {
      value,
      y,
      top,
      label: formatter(value)
    }
  })
}

function formatDate(value) {
  if (!value) return 'N/A'
  return new Date(value).toLocaleString()
}

function formatElapsed(startedAt, completedAt, nowValue) {
  if (!startedAt) return 'elapsed N/A'
  const start = new Date(startedAt).getTime()
  const end = completedAt ? new Date(completedAt).getTime() : nowValue
  const diff = Math.max(0, Math.floor((end - start) / 1000))
  const minutes = Math.floor(diff / 60)
  const seconds = diff % 60
  const prefix = completedAt ? 'duration' : 'elapsed'
  return `${prefix} ${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
}

function formatDurationMs(value) {
  const totalSeconds = Math.max(0, Math.floor(Number(value || 0) / 1000))
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
}

function setLogActionNotice(message, tone) {
  logActionNotice.value = { message, tone }
  if (logNoticeTimer) clearTimeout(logNoticeTimer)
  logNoticeTimer = setTimeout(() => {
    logActionNotice.value = null
  }, 5000)
}

function sanitizeName(value) {
  return String(value || '')
    .trim()
    .replace(/\s+/g, '-')
    .replace(/[^a-zA-Z0-9-_]/g, '')
}

function buildTaskNameStamp() {
  return new Date().toISOString().slice(0, 19).replace(/[-:T]/g, '')
}

function runtimeOverridesSummary() {
  const parts = [`run_duration=${Number(draft.overrides.duration || 0)}s`]
  if (selectedTemplate.value) {
    parts.push(`${concurrencyKey.value}=${Number(draft.overrides[concurrencyKey.value] || 0)}`)
  }
  return parts.join(', ')
}

function applyPreferredTemplateSelection({ preferTestTemplate = false } = {}) {
  if (!draft.database_type || !filteredTemplates.value.length) return
  if (draft.template_id && !preferTestTemplate) return

  const preferredTemplateId = getPreferredTemplateId(filteredTemplates.value, {
    fallbackTemplateId: draft.template_id
  })

  if (preferredTemplateId && preferredTemplateId !== draft.template_id) {
    draft.template_id = preferredTemplateId
  }
}

function restoreDraft() {
  const raw = window.localStorage.getItem(TASKS_DRAFT_STORAGE_KEY)
  if (!raw) return
  try {
    const saved = JSON.parse(raw)
    draft.database_type = typeof saved.database_type === 'string' ? saved.database_type : ''
    draft.connection_id = typeof saved.connection_id === 'string' ? saved.connection_id : ''
    draft.template_id = typeof saved.template_id === 'string' ? saved.template_id : ''
    draft.action = typeof saved.action === 'string' ? saved.action : 'full_pipeline'
    if (saved.overrides && typeof saved.overrides === 'object') {
      if (Number.isFinite(Number(saved.overrides.duration))) {
        draft.overrides.duration = Number(saved.overrides.duration)
      }
      if (Number.isFinite(Number(saved.overrides.threads))) {
        draft.overrides.threads = Number(saved.overrides.threads)
      }
      if (Number.isFinite(Number(saved.overrides.virtual_users))) {
        draft.overrides.virtual_users = Number(saved.overrides.virtual_users)
      }
    }
  } catch (error) {
    window.localStorage.removeItem(TASKS_DRAFT_STORAGE_KEY)
  }
}

function sanitizeDraftState() {
  if (!draft.database_type) {
    draft.connection_id = ''
    draft.template_id = ''
  }
  if (draft.connection_id && !filteredConnections.value.some((connection) => connection.id === draft.connection_id)) {
    draft.connection_id = ''
  }
  if (draft.template_id && !templateMatchesDatabaseAndConnection(draft.template_id)) {
    draft.template_id = ''
  }
  const currentAction = actionOptions.value.find((option) => option.value === draft.action)
  if (!currentAction || currentAction.disabled) {
    const fallbackAction = actionOptions.value.find((option) => !option.disabled)
    draft.action = fallbackAction?.value || 'full_pipeline'
  }
}

function templateMatchesDatabaseAndConnection(templateId) {
  if (!draft.database_type) return false
  return filteredTemplates.value.some((template) => template.id === templateId)
}

function persistDraft() {
  window.localStorage.setItem(TASKS_DRAFT_STORAGE_KEY, JSON.stringify({
    database_type: draft.database_type,
    connection_id: draft.connection_id,
    template_id: draft.template_id,
    action: draft.action,
    overrides: {
      duration: Number(draft.overrides.duration || 0),
      threads: Number(draft.overrides.threads || 0),
      virtual_users: Number(draft.overrides.virtual_users || 0)
    }
  }))
}

function scrollLogsToBottom() {
  if (logViewport.value) {
    logViewport.value.scrollTop = logViewport.value.scrollHeight
  }
}

function closePreview() {
  previewOpen.value = false
  previewTask.value = null
  previewConfirmable.value = false
}

function closeLogViewer() {
  logViewerOpen.value = false
}
</script>

<style scoped>
.tasks-monitor-tab {
  display: flex;
  flex-direction: column;
  gap: 8px;
  color: var(--text-primary);
  height: 100%;
  min-height: 0;
  overflow: hidden;
  font-size: 12px;
  --tm-font-body: 12px;
  --tm-font-meta: 10px;
  --tm-font-label: 10px;
  --tm-font-title: 20px;
  --tm-font-panel-title: 15px;
  --tm-font-button: 12px;
  --tm-font-strong: 13px;
  --tm-panel-pad: 12px;
  --tm-card-gap: 10px;
}

.tab-header,
.card-head,
.actions,
.log-toolbar,
.header-grid,
.header-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.tab-header h2,
.card-head h3,
.header-panel h3 {
  margin: 0;
}

.page-subtitle,
.header-panel p,
.preview-grid p,
.preview-meta-grid p,
.log-meta {
  margin: 4px 0 0;
  color: var(--text-muted);
  font-size: var(--tm-font-body);
}

.workspace-grid {
  display: grid;
  grid-template-columns: 250px minmax(0, 1fr);
  gap: 10px;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.left-column,
.right-column {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 0;
}

.panel {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--tm-panel-pad);
  box-shadow: var(--shadow-sm);
}

.status-pill,
.handoff-badge,
.phase-tag,
.stream-tag {
  border-radius: 999px;
  padding: 4px 10px;
  font-size: var(--tm-font-meta);
  background: rgba(120, 135, 160, 0.18);
}

.top-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
  flex-wrap: wrap;
}

.btn {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 8px 12px;
  font-weight: 500;
  font-size: var(--tm-font-button);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn:hover:not(:disabled) {
  background-color: var(--bg-hover);
  border-color: var(--border-dark);
}

.btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.btn-secondary {
  background: var(--bg-primary);
  border-color: var(--border-color);
  color: var(--text-primary);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--bg-hover);
  border-color: var(--border-dark);
}

.btn-primary {
  background: var(--primary);
  border-color: var(--primary);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--primary-hover);
  border-color: var(--primary-hover);
}

.btn-danger {
  background: var(--danger);
  border-color: var(--danger);
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: var(--danger-hover);
  border-color: var(--danger-hover);
}

.primary-action,
.danger-action,
.log-action {
  min-width: 100px;
  padding: 10px 14px;
  font-size: var(--tm-font-button);
  font-weight: 700;
  letter-spacing: 0.02em;
}

.status-entry {
  position: relative;
}

.status-summary {
  border: 1px solid var(--border-color);
  border-radius: 999px;
  background: var(--bg-primary);
  color: var(--text-primary);
  padding: 6px 10px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.status-summary.ready {
  border-color: var(--success);
  color: var(--success);
  background: var(--success-bg);
}

.status-summary.warning {
  border-color: var(--warning);
  color: var(--warning);
  background: var(--warning-bg);
}

.status-summary.blocked {
  border-color: var(--danger);
  color: var(--danger);
  background: var(--danger-bg);
}

.status-summary-badge {
  font-size: var(--tm-font-meta);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.status-summary-icon {
  font-weight: 700;
}

.status-popover {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 360px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-primary);
  padding: 12px;
  box-shadow: var(--shadow-dropdown);
  z-index: 20;
}

.status-popover-head {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 10px;
}

.status-popover-head span {
  color: var(--text-muted);
  font-size: var(--tm-font-body);
}

.status-popover-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.status-popover-item {
  border: 1px solid var(--border-light);
  border-radius: var(--radius-sm);
  background: var(--bg-secondary);
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-popover-item span {
  color: var(--text-muted);
  font-size: var(--tm-font-meta);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.status-pill.running,
.status-pill.preparing,
.status-pill.cleaning,
.status-pill.starting,
.status-pill.stopping {
  background: var(--success-bg);
  color: var(--success);
}

.status-pill.failed,
.status-pill.stopped {
  background: var(--danger-bg);
  color: var(--danger);
}

.status-pill.success {
  background: var(--success-bg);
  color: var(--success);
}

.status-pill.idle {
  background: var(--bg-secondary);
  color: var(--text-muted);
}

.status-strip {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: start;
  gap: 12px;
  padding: 12px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
}

.status-strip.active {
  border-color: var(--success);
}

.status-strip-main,
.status-strip-meta,
.status-strip-phase-block,
.status-strip-timings {
  display: flex;
  align-items: center;
  gap: 10px;
}

.status-strip-main {
  justify-content: flex-start;
  min-width: 0;
  gap: 16px;
}

.status-strip-phase-block {
  min-width: 0;
  align-items: center;
}

.status-strip-phase-copy {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.status-strip-timings,
.status-strip-meta {
  flex-wrap: wrap;
  justify-content: flex-start;
}

.status-strip-time-chip,
.status-strip-meta-chip {
  display: grid;
  gap: 4px;
  min-width: 88px;
  padding: 7px 10px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  background: var(--bg-primary);
}

.status-strip-label,
.status-strip-phase,
.status-strip-time,
.status-strip-meta strong {
  font-size: var(--tm-font-meta);
  letter-spacing: 0.04em;
}

.status-strip-label {
  color: var(--text-muted);
  text-transform: uppercase;
}

.status-strip-phase {
  color: var(--text-primary);
  text-transform: none;
  font-size: var(--tm-font-body);
}

.status-strip-detail {
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.35;
  max-width: 56ch;
  min-height: 1.35em;
}

.status-strip-time {
  color: var(--text-primary);
  font-weight: 600;
  font-size: var(--tm-font-body);
}

.action-notice {
  padding: 8px 12px;
  background: var(--bg-secondary);
  border-radius: var(--radius-sm);
}

.action-notice.success {
  color: var(--success);
  border: 1px solid var(--success);
}

.action-notice.error {
  color: var(--danger);
  border: 1px solid var(--danger);
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 10px;
}

.field span,
.eyebrow {
  font-size: var(--tm-font-label);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-muted);
}

input,
select,
.params-preview {
  width: 100%;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background: var(--bg-primary);
  color: var(--text-primary);
  padding: 8px 10px;
  font-size: var(--tm-font-body);
}

input:focus,
select:focus,
.params-preview:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 2px var(--primary-light);
}

.readonly-field {
  width: 100%;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background: var(--bg-secondary);
  color: var(--text-primary);
  padding: 8px 10px;
  min-height: 36px;
  display: flex;
  align-items: center;
  font-size: var(--tm-font-body);
}

.readonly-field.placeholder {
  color: var(--text-placeholder);
}

.compact-help {
  margin: 8px 0 0;
  color: var(--text-muted);
  font-size: var(--tm-font-body);
}

.select-dark {
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  background-color: var(--bg-primary);
  background-image:
    linear-gradient(45deg, transparent 50%, var(--primary) 50%),
    linear-gradient(135deg, var(--primary) 50%, transparent 50%);
  background-position:
    calc(100% - 18px) calc(50% - 2px),
    calc(100% - 12px) calc(50% - 2px);
  background-size: 6px 6px, 6px 6px;
  background-repeat: no-repeat;
  border-color: var(--border-color);
  color: var(--text-primary);
  padding-right: 34px;
}

.select-dark:hover {
  border-color: var(--border-dark);
  background-color: var(--bg-hover);
}

.select-dark:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 2px var(--primary-light);
}

.select-dark:disabled {
  background-color: var(--bg-secondary);
  border-color: var(--border-light);
  color: var(--text-muted);
  cursor: not-allowed;
}

.select-dark option {
  background: var(--bg-primary);
  color: var(--text-primary);
}

.override-grid,
.metric-grid,
.system-grid,
.preview-grid,
.preview-meta-grid {
  display: grid;
  gap: 10px;
}

.override-grid,
.preview-grid,
.preview-meta-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.metric-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.system-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.metric-card,
.system-card {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 12px;
  background: var(--bg-primary);
}

.monitor-board {
  min-height: 0;
  display: flex;
  flex-direction: column;
  flex: 1;
  height: 100%;
  overflow: hidden;
}

.monitor-board-grid {
  display: grid;
  grid-template-rows: minmax(244px, 0.82fr) minmax(286px, 1fr);
  gap: 10px;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.text-action {
  background: transparent;
  border: 0;
  color: var(--primary);
  cursor: pointer;
}

.text-action.danger {
  color: var(--danger);
}

.metric-card,
.system-card {
  color: var(--text-primary);
}

.metric-card {
  display: grid;
  grid-template-rows: 24px minmax(0, 1fr);
  gap: 6px;
  min-height: 0;
  overflow: hidden;
  background: var(--bg-primary);
}

.metric-card-head,
.metric-stats,
.system-card-head,
.chart-legend {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.metric-title,
.system-title-block > span {
  font-size: var(--tm-font-label);
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-muted);
}

.metric-card-head {
  min-height: 24px;
  gap: 8px;
}

.metric-inline-bar {
  display: inline-flex;
  align-items: baseline;
  gap: 10px;
  min-width: 0;
  flex-wrap: nowrap;
}

.metric-status {
  padding: 2px 7px;
  border-radius: 999px;
  font-size: var(--tm-font-meta);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  border: 1px solid rgba(255, 255, 255, 0.1);
  flex-shrink: 0;
}

.metric-stable {
  color: var(--success);
  border-color: var(--success);
  background: var(--success-bg);
}

.metric-fluctuating {
  color: var(--warning);
  border-color: var(--warning);
  background: var(--warning-bg);
}

.metric-sawtooth {
  color: var(--danger);
  border-color: var(--danger);
  background: var(--danger-bg);
}

.metric-prepare {
  color: var(--primary);
  border-color: var(--primary);
  background: var(--primary-light);
}

.metric-waiting {
  color: var(--warning);
  border-color: var(--warning);
  background: var(--warning-bg);
}

.metric-error {
  color: var(--danger);
  border-color: var(--danger);
  background: var(--danger-bg);
}

.metric-stats {
  justify-content: flex-start;
  align-items: baseline;
  gap: 8px;
  min-height: 0;
  flex-wrap: nowrap;
  min-width: 0;
}

.metric-stats div {
  display: inline-flex;
  align-items: baseline;
  gap: 4px;
  min-width: 0;
  white-space: nowrap;
}

.metric-stats span {
  font-size: var(--tm-font-meta);
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.metric-stats strong {
  color: var(--text-primary);
  font-size: 11px;
  line-height: 1;
}

.metric-history,
.system-chart-wrap {
  flex: 1;
  min-height: 0;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  background: var(--bg-secondary);
  padding: 8px;
  overflow: hidden;
}

.chart-shell {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr) 48px;
  height: 100%;
  min-height: 0;
  gap: 6px;
  overflow: hidden;
}

.metric-chart-shell {
  grid-template-columns: 48px minmax(0, 1fr);
}

.system-chart-shell {
  min-height: 0;
}

.system-card[data-system-card="CPU"] .system-chart-shell {
  grid-template-columns: 58px minmax(0, 1fr);
  gap: 8px;
}

.system-card[data-system-card="Disk IO"] .system-chart-shell {
  grid-template-columns: 62px minmax(0, 1fr) 62px;
  gap: 8px;
}

.chart-canvas {
  position: relative;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  border-radius: var(--radius-sm);
}

.metric-chart-canvas {
  padding: 14px 0 6px;
}

.metric-chart-current {
  position: absolute;
  top: 10px;
  left: 12px;
  z-index: 2;
  display: inline-flex;
  align-items: baseline;
  gap: 6px;
  max-width: calc(100% - 20px);
  pointer-events: none;
}

.metric-chart-current-muted {
  top: auto;
  left: auto;
  right: 12px;
  bottom: 10px;
  opacity: 0.28;
}

.metric-overlay {
  position: absolute;
  inset: 20px 18px 18px;
  z-index: 2;
  display: grid;
  align-content: center;
  justify-items: center;
  text-align: center;
  gap: 6px;
  padding: 18px 20px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid var(--border-color);
  backdrop-filter: blur(2px);
  pointer-events: none;
}

.metric-overlay strong {
  font-size: 13px;
  color: var(--text-primary);
}

.metric-overlay span {
  max-width: 32ch;
  font-size: 11px;
  line-height: 1.45;
  color: var(--text-secondary);
}

.metric-overlay-run-waiting {
  border-color: var(--warning);
}

.metric-overlay-run-error {
  border-color: var(--danger);
  background: var(--danger-bg);
}

.metric-chart-current strong {
  font-size: clamp(16px, 1.6vw, 22px);
  line-height: 1;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.metric-chart-current span {
  font-size: 9px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  white-space: nowrap;
}

.chart-axis {
  position: relative;
  min-height: 0;
  padding: 18px 0 8px;
}

.chart-axis::after {
  content: '';
  position: absolute;
  top: 18px;
  bottom: 8px;
  width: 1px;
  background: var(--border-light);
  pointer-events: none;
}

.chart-axis span {
  position: absolute;
  transform: translateY(-50%);
  font-size: 9px;
  line-height: 1;
  color: var(--text-muted);
  white-space: nowrap;
}

.chart-axis-left span {
  right: 0;
}

.chart-axis-right span {
  left: 0;
}

.chart-axis-left::after {
  right: 0;
}

.chart-axis-right::after {
  left: 0;
}

.axis-unit {
  position: absolute;
  top: 6px;
  transform: none;
  font-size: 9px;
  line-height: 1;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: var(--text-muted);
  white-space: nowrap;
  z-index: 1;
  pointer-events: none;
}

.axis-unit-left {
  right: 0;
}

.axis-unit-right {
  left: 0;
}

.chart-axis-cpu span {
  color: #9ed5c2;
}

.chart-axis-cpu::after {
  background: linear-gradient(180deg, rgba(52, 211, 153, 0.22), rgba(52, 211, 153, 0.08));
}

.chart-axis-bandwidth span {
  color: #93bfff;
}

.chart-axis-bandwidth .axis-unit {
  color: #60a5fa;
}

.chart-axis-bandwidth::after {
  background: linear-gradient(180deg, rgba(96, 165, 250, 0.24), rgba(96, 165, 250, 0.08));
}

.chart-axis-latency span {
  color: #a9e6bf;
}

.chart-axis-latency .axis-unit {
  color: #86efac;
}

.chart-axis-latency::after {
  background: linear-gradient(180deg, rgba(134, 239, 172, 0.24), rgba(134, 239, 172, 0.08));
}

.system-card[data-system-card="CPU"] .chart-axis,
.system-card[data-system-card="Disk IO"] .chart-axis {
  padding: 18px 0 8px;
}

.history-chart,
.system-chart {
  width: 100%;
  height: 100%;
  display: block;
}

.history-glow {
  opacity: 0.12;
}

.history-baseline,
.chart-gridline {
  stroke: var(--border-light);
  stroke-width: 0.8;
  stroke-dasharray: 2 5;
}

.system-grid .system-card {
  display: grid;
  grid-template-rows: minmax(46px, auto) minmax(0, 1fr) auto;
  gap: 4px;
  min-height: 0;
  overflow: hidden;
}

.system-card[data-system-card="CPU"] .system-chart-wrap,
.system-card[data-system-card="Disk IO"] .system-chart-wrap {
  align-self: start;
  padding: 6px 7px;
  min-height: 150px;
  height: 100%;
}

.system-card[data-system-card="CPU"] .system-chart-shell,
.system-card[data-system-card="Disk IO"] .system-chart-shell {
  height: 100%;
  min-height: 0;
}

.system-card.disabled {
  opacity: 0.62;
}

.system-card-head {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: flex-start;
  min-height: 46px;
  gap: 10px;
}

.system-card-head strong {
  display: block;
  margin-top: 2px;
  color: var(--text-primary);
  font-size: 11px;
  line-height: 1.2;
}

.system-title-block {
  min-height: 46px;
  display: grid;
  grid-template-rows: 14px minmax(0, 1fr);
  align-content: start;
  min-width: 0;
}

.system-summary-lines {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-height: 0;
  justify-content: flex-start;
  min-width: 0;
}

.summary-line {
  display: flex;
  align-items: baseline;
  gap: 6px;
  flex-wrap: nowrap;
  white-space: nowrap;
  overflow: visible;
  min-width: 0;
}

.summary-chip {
  display: inline-flex;
  white-space: nowrap;
  flex-shrink: 0;
}

.summary-chip + .summary-chip::before {
  content: '·';
  margin-right: 6px;
  color: var(--text-muted);
}

.chart-legend {
  justify-content: flex-end;
  flex-wrap: wrap;
  align-content: flex-start;
  gap: 6px;
  max-width: 128px;
}

.chart-legend-cpu {
  flex-wrap: nowrap;
  align-items: center;
  gap: 5px;
  max-width: none;
  white-space: nowrap;
}

.chart-legend-single {
  flex-wrap: nowrap;
  gap: 5px;
  white-space: nowrap;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 9px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.legend-item i {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  display: inline-block;
}

.system-caption {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--tm-font-meta);
  min-height: 0;
  line-height: 1.15;
  display: flex;
  align-items: flex-start;
}

.disk-chart-combined {
  position: relative;
}

.terminal-viewer {
  background: #1e2127;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  font-family: 'JetBrains Mono', 'SFMono-Regular', monospace;
  overflow: auto;
}

.terminal-viewer {
  height: 420px;
  padding: 12px;
}

.terminal-line {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  align-items: start;
  gap: 8px;
  padding: 6px 0;
  color: #abb2bf;
}

.terminal-line-badges {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
  flex: 0 0 auto;
}

.terminal-line-message {
  flex: 1;
  min-width: 0;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.45;
}

.stream-tag {
  min-width: auto;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  background: var(--bg-secondary);
}

.phase-tag,
.stream-tag {
  display: inline-flex;
  align-items: center;
  padding: 1px 6px;
  line-height: 1.1;
  border-radius: 999px;
  font-size: 10px;
  border: 1px solid var(--border-color);
}

.terminal-line.error {
  color: #e06c75;
}

.terminal-line.muted {
  color: #5c6370;
}

.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: grid;
  place-items: center;
  padding: 20px;
  z-index: 20;
}

.modal {
  width: min(720px, 100%);
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 20px;
  box-shadow: var(--shadow-lg);
}

.preview-meta-card {
  display: grid;
  gap: 5px;
  align-content: start;
  min-height: 88px;
  padding: 12px 13px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  background: var(--bg-secondary);
}

.preview-meta-card strong {
  line-height: 1.25;
  min-width: 0;
}

.preview-meta-card p {
  margin: 0;
  color: var(--text-muted);
  line-height: 1.35;
}

.preview-params-section {
  display: grid;
  gap: 8px;
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid var(--border-light);
}

.params-preview {
  margin: 0;
  min-height: 220px;
}

.modal-actions {
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
}

.modal-wide {
  width: min(1100px, 100%);
}

.empty-state {
  color: var(--text-muted);
}

.tab-header {
  gap: 12px;
}

.tab-header h2 {
  font-size: var(--tm-font-title);
  line-height: 1.1;
}

.card-head h3,
.header-panel h3 {
  font-size: var(--tm-font-panel-title);
  line-height: 1.15;
}

.card-head span,
.log-toolbar,
.status-entry,
.status-popover-item strong,
.preview-grid strong,
.preview-grid p,
.preview-meta-grid strong,
.preview-meta-grid p,
.params-preview,
.text-action,
.modal,
.terminal-line {
  font-size: var(--tm-font-body);
}

.tab-header,
.card-head,
.actions,
.log-toolbar,
.header-grid,
.header-meta {
  gap: 10px;
}

.card {
  height: 100%;
  min-height: 0;
}

.right-column {
  min-height: 0;
}

.metric-grid,
.system-grid {
  height: 100%;
  min-height: 0;
  align-items: stretch;
}

.metric-history {
  min-height: 0;
}

.system-chart-wrap {
  min-height: 0;
}

.monitor-board .card-head {
  margin-bottom: 6px;
}

.tab-header .page-subtitle {
  max-width: 540px;
  line-height: 1.35;
}

@media (max-width: 1180px) {
  .workspace-grid,
  .metric-grid,
  .preview-grid,
  .preview-meta-grid,
  .system-grid,
  .override-grid {
    grid-template-columns: 1fr;
  }

  .tab-header,
  .top-controls,
  .status-strip-main,
  .status-strip-meta {
    flex-direction: column;
    align-items: stretch;
  }

  .status-strip {
    grid-template-columns: 1fr;
  }

  .tasks-monitor-tab {
    overflow: auto;
  }

  .monitor-board-grid {
    grid-template-rows: auto;
  }

  .summary-line {
    flex-wrap: wrap;
    row-gap: 2px;
  }
}
</style>
