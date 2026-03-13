<template>
  <div class="tasks-monitor-tab">
    <div class="tab-header">
      <div>
        <h2>Tasks &amp; Monitor</h2>
        <p class="page-subtitle">Template / Connection 选择、Preview、单任务执行、实时指标和 tail 风格日志都在这个工作台完成。</p>
      </div>
      <div class="top-controls">
        <button class="btn btn-primary primary-action" :disabled="!canStart" @click="openPreview">
          Start
        </button>
        <button class="btn btn-danger danger-action" :disabled="!canStop" @click="handleStop">
          Stop
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
        <span class="status-pill" :class="statusStrip.badgeClass">{{ statusStrip.status }}</span>
        <span class="status-strip-phase">{{ statusStrip.phase }}</span>
        <span class="status-strip-time">{{ statusStrip.time }}</span>
      </div>
      <div class="status-strip-meta" v-if="statusStrip.meta">
        {{ statusStrip.meta }}
      </div>
    </section>

    <div class="workspace-grid">
      <div class="left-column">
        <section class="panel card">
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
                {{ connection.name }} · {{ connection.type }}
              </option>
            </select>
          </label>

          <label class="field">
            <span>Template</span>
            <select v-model="draft.template_id" class="select-dark" :disabled="!draft.database_type || !draft.connection_id">
              <option value="">{{ draft.connection_id ? 'Select template' : 'Select connection first' }}</option>
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
              <span>Duration (s)</span>
              <input v-model.number="draft.overrides.duration" type="number" min="1">
            </label>
          </div>

          <p class="compact-help">Use the top Start button to preview and launch the task.</p>
        </section>
      </div>

      <div class="right-column">
        <section class="panel metrics-panel">
          <div class="card-head">
            <h3>Current Throughput</h3>
            <span>TPS / TPM</span>
          </div>
          <div class="metric-grid">
            <div class="metric-card" v-for="metric in businessMetrics" :key="metric.label">
              <div class="metric-card-head">
                <span>{{ metric.label }}</span>
                <span class="metric-status" :class="metric.statusClass">{{ metric.statusLabel }}</span>
              </div>
              <div class="metric-main">
                <strong>{{ metric.current }}</strong>
                <span>{{ metric.unit }}</span>
              </div>
              <div class="metric-stats">
                <div>
                  <span>avg</span>
                  <strong>{{ metric.avg }}</strong>
                </div>
                <div>
                  <span>max</span>
                  <strong>{{ metric.max }}</strong>
                </div>
              </div>
              <div class="metric-history">
                <svg class="history-chart" viewBox="0 0 320 116" preserveAspectRatio="none">
                  <path :d="metric.areaPath" :fill="metric.fill" />
                  <polyline :points="metric.points" :stroke="metric.stroke" stroke-width="3" fill="none" />
                  <line :x1="0" :x2="320" :y1="metric.avgLineY" :y2="metric.avgLineY" class="history-baseline" />
                </svg>
              </div>
            </div>
          </div>
        </section>

        <section class="panel metrics-panel system-panel" :class="{ disabled: !systemEnabled }">
          <div class="card-head">
            <h3>System Metrics</h3>
            <span v-if="!systemEnabled">{{ systemMessage }}</span>
          </div>
          <div class="system-grid">
            <div v-for="chart in systemCharts" :key="chart.label" class="system-card">
              <div class="system-card-head">
                <div>
                  <span>{{ chart.label }}</span>
                  <strong>{{ chart.summary }}</strong>
                </div>
                <div class="chart-legend">
                  <span v-for="line in chart.lines" :key="line.label" class="legend-item">
                    <i :style="{ background: line.color }"></i>
                    {{ line.label }}
                  </span>
                </div>
              </div>
              <div class="system-chart-wrap">
                <svg class="system-chart" viewBox="0 0 320 124" preserveAspectRatio="none">
                  <line v-for="grid in chart.gridLines" :key="grid" :x1="0" :x2="320" :y1="grid" :y2="grid" class="chart-gridline" />
                  <polyline
                    v-for="line in chart.lines"
                    :key="line.label"
                    :points="line.points"
                    :stroke="line.color"
                    stroke-width="3"
                    fill="none"
                  />
                </svg>
              </div>
              <p class="system-caption">{{ chart.caption }}</p>
            </div>
          </div>
        </section>
      </div>
    </div>

    <section class="panel log-summary">
      <div class="card-head">
        <h3>Log Output</h3>
        <button class="btn btn-secondary" :disabled="!currentTask" @click="openLogViewer">Open Log Viewer</button>
      </div>
      <p class="log-meta" v-if="currentTask">
        UI keeps the latest 500 lines. Full phase logs are persisted on disk:
        {{ logPathSummary }}
      </p>
      <div class="terminal-summary">
        <div v-for="(line, index) in summaryLines" :key="`${line.timestamp}-${index}`" class="terminal-line" :class="lineClass(line)">
          <span class="phase-tag">{{ line.phase }}</span>
          <span class="stream-tag">{{ line.stream }}</span>
          <span>{{ line.content }}</span>
        </div>
      </div>
    </section>

    <div v-if="previewOpen" class="modal-backdrop" @click.self="closePreview">
      <div class="modal">
        <div class="card-head">
          <h3>Task Preview</h3>
          <button class="text-action" @click="closePreview">Close</button>
        </div>
        <div class="preview-grid" v-if="previewTask">
          <div>
            <span class="eyebrow">Template</span>
            <strong>{{ previewTask.template_snapshot.name }}</strong>
            <p>{{ previewTask.template_snapshot.tool }} · {{ previewTask.template_snapshot.db_family }}</p>
          </div>
          <div>
            <span class="eyebrow">Connection</span>
            <strong>{{ previewTask.connection_snapshot.name }}</strong>
            <p>{{ previewTask.connection_snapshot.type }} · {{ previewTask.connection_snapshot.host }}</p>
          </div>
          <div>
            <span class="eyebrow">Action</span>
            <strong>{{ actionLabel(previewTask.action) }}</strong>
          </div>
          <div>
            <span class="eyebrow">Monitoring</span>
            <strong>{{ previewTask.readiness?.ssh_available ? 'SSH metrics enabled' : 'SSH unavailable' }}</strong>
            <p>{{ previewTask.readiness?.ssh_message }}</p>
          </div>
        </div>
        <pre class="params-preview">{{ JSON.stringify(previewTask?.resolved_params || {}, null, 2) }}</pre>
        <div class="actions">
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
            <span class="phase-tag">{{ line.phase }}</span>
            <span class="stream-tag">{{ line.stream }}</span>
            <span>{{ line.content }}</span>
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

const appStore = useAppStore()
const templateStore = useTemplateStore()
const connectionStore = useConnectionStore()
const taskStore = useTaskStore()

const { templates } = storeToRefs(templateStore)
const { connections } = storeToRefs(connectionStore)
const { currentTask, activeTask } = storeToRefs(taskStore)
const TASKS_DRAFT_STORAGE_KEY = 'db-benchmind.tasks-monitor.draft.v1'

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
const logQuery = ref('')
const logPhase = ref('')
const autoScroll = ref(true)
const logViewport = ref(null)
const nowTick = ref(Date.now())

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
  if (!draft.database_type || !draft.connection_id) return []
  return templateStore.templatesByDatabase(draft.database_type)
    .filter((template) => template.dbFamily === draft.database_type || template.database_types?.includes(draft.database_type))
})

const concurrencyKey = computed(() => {
  if (selectedTemplate.value?.tool === 'hammerdb' || selectedTemplate.value?.tool === 'swingbench') return 'virtual_users'
  return 'threads'
})

const concurrencyLabel = computed(() => concurrencyKey.value === 'threads' ? 'Threads' : 'Virtual Users')
const showThreads = computed(() => !!selectedTemplate.value)
const actionOptions = computed(() => {
  const phases = selectedTemplate.value?.phases || {}
  const allowFullPipeline = Boolean(phases.prepare?.enabled && phases.run?.enabled && phases.cleanup?.enabled)
  return [
    { value: 'prepare', label: 'Prepare', disabled: !phases.prepare?.enabled },
    { value: 'run', label: 'Run', disabled: !phases.run?.enabled },
    { value: 'cleanup', label: 'Cleanup', disabled: !phases.cleanup?.enabled },
    { value: 'full_pipeline', label: 'Run Full Flow (Prepare -> Run -> Cleanup)', disabled: !allowFullPipeline }
  ]
})
const selectedActionOption = computed(() => actionOptions.value.find((action) => action.value === draft.action) || null)
const canPreview = computed(() => !!draft.database_type && !!draft.template_id && !!draft.connection_id && selectedActionOption.value && !selectedActionOption.value.disabled)
const canStart = computed(() => canPreview.value && !taskStore.hasActiveTask)
const canStop = computed(() => !!activeTask.value)
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
      { label: 'Elapsed', value: elapsedLabel.value },
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
    metricCard('TPS', metrics.tps, '/sec', '#f4a261', 'rgba(244, 162, 97, 0.28)'),
    metricCard('TPM', metrics.tpm, '/min', '#4dd0a8', 'rgba(77, 208, 168, 0.24)')
  ]
})

const systemEnabled = computed(() => !!currentTask.value?.metrics?.system_enabled)
const systemMessage = computed(() => currentTask.value?.metrics?.system_message || 'SSH unavailable, benchmark continues without system metrics')

const systemCharts = computed(() => {
  const metrics = currentTask.value?.metrics || {}
  return [
    multiMetricChart('CPU', [
      { label: 'user', metric: metrics.cpu_user, color: '#7dd3fc', formatter: formatPercent },
      { label: 'sys', metric: metrics.cpu_sys, color: '#f4a261', formatter: formatPercent },
      { label: 'iowait', metric: metrics.cpu_iowait, color: '#f87171', formatter: formatPercent }
    ], 'Recent CPU composition across user, sys, and iowait.'),
    multiMetricChart('Disk IO', [
      { label: 'read', metric: metrics.disk_read_bps, color: '#60a5fa', formatter: formatBytes },
      { label: 'write', metric: metrics.disk_write_bps, color: '#34d399', formatter: formatBytes }
    ], 'Read and write throughput over the current rolling window.')
  ]
})

const summaryLines = computed(() => (currentTask.value?.log_tail || []).slice(-10))
const elapsedLabel = computed(() => formatElapsed(currentTask.value?.started_at, currentTask.value?.completed_at, nowTick.value))
const logPathSummary = computed(() => {
  const paths = Object.entries(currentTask.value?.run_log_paths || {})
    .filter(([, path]) => !!path)
    .map(([phase, path]) => `${phase}: ${path}`)
  return paths.length ? paths.join(' | ') : 'pending'
})
const statusStrip = computed(() => {
  if (!currentTask.value) {
    return {
      status: 'Idle',
      phase: 'No active task',
      time: 'Ready',
      meta: '',
      badgeClass: 'idle',
      stateClass: 'idle'
    }
  }

  const status = statusLabel(currentTask.value.status)
  const phase = phaseLabel(currentTask.value.current_phase, currentTask.value.status)
  const time = formatElapsed(currentTask.value.started_at, currentTask.value.completed_at, nowTick.value)
  const metaParts = [
    currentTask.value.benchmark_tool || '',
    currentTask.value.connection_snapshot?.type || '',
    currentTask.value.template_snapshot?.name || ''
  ].filter(Boolean)

  return {
    status,
    phase,
    time,
    meta: metaParts.join(' · '),
    badgeClass: currentTask.value.status || 'idle',
    stateClass: isActive(currentTask.value.status) ? 'active' : currentTask.value.status || 'idle'
  }
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

onMounted(async () => {
  await Promise.all([templateStore.loadTemplates?.() || templateStore.fetchTemplates?.() || Promise.resolve(), connectionStore.fetchConnections(), taskStore.fetchTasks()])
  restoreDraft()
  sanitizeDraftState()
  taskStore.startPolling()
  clockTimer = setInterval(() => {
    nowTick.value = Date.now()
  }, 1000)
  if (pendingTaskTemplate.value?.templateId) {
    const handoffTemplate = templates.value.find((template) => template.id === pendingTaskTemplate.value.templateId)
    if (handoffTemplate) {
      draft.database_type = handoffTemplate.dbFamily || handoffTemplate.database_types?.[0] || ''
    }
    appStore.clearPendingTaskTemplate()
  }
  sanitizeDraftState()
})

onBeforeUnmount(() => {
  taskStore.stopPolling()
  clearTimeout(validationTimer)
  if (clockTimer) clearInterval(clockTimer)
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
})

watch(() => draft.connection_id, () => {
  if (selectedTemplate.value && !filteredTemplates.value.some((template) => template.id === selectedTemplate.value.id)) {
    draft.template_id = ''
  }
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
  if (!activeTask.value) return
  await taskStore.stopTask(activeTask.value.id)
}

async function openLogViewer() {
  if (!currentTask.value) return
  logViewerOpen.value = true
  await taskStore.fetchLogs({ taskId: currentTask.value.id, query: logQuery.value, phase: logPhase.value })
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

function metricCard(label, metric = {}, unit, stroke, fill) {
  const current = Number(metric.current || 0)
  const avg = Number(metric.avg || 0)
  const max = Number(metric.max || 0)
  const series = Array.isArray(metric.series) ? metric.series : []
  const trend = buildTrendData(series, 320, 116)
  return {
    label,
    current: formatNumber(current),
    avg: formatNumber(avg),
    max: formatNumber(max),
    unit,
    stroke,
    fill,
    points: trend.points,
    areaPath: trend.areaPath,
    avgLineY: mapSeriesValueToY(avg, trend.min, trend.max, 116),
    statusLabel: metricStatus(metric),
    statusClass: `metric-${metricStatus(metric).toLowerCase()}`
  }
}

function multiMetricChart(label, definitions, caption) {
  const values = definitions.map((item) => Number(item.metric?.current || 0))
  const allSeriesValues = definitions.flatMap((item) => (item.metric?.series || []).map((point) => Number(point.value || 0)))
  const min = label === 'CPU' ? 0 : Math.min(...allSeriesValues, 0)
  const max = label === 'CPU' ? Math.max(100, ...allSeriesValues, 1) : Math.max(...allSeriesValues, 1)
  return {
    label,
    summary: definitions.map((item, index) => `${item.label} ${item.formatter(values[index])}`).join(' · '),
    caption,
    gridLines: [20, 62, 104],
    lines: definitions.map((item) => ({
      label: item.label,
      color: item.color,
      points: buildTrend(item.metric?.series || [], 320, 124, { min, max })
    }))
  }
}

function buildTrend(series = [], width, height, bounds = null) {
  return buildTrendData(series, width, height, bounds).points
}

function buildTrendData(series = [], width, height, bounds = null) {
  if (!series?.length) {
    return { points: '', areaPath: '', min: bounds?.min ?? 0, max: bounds?.max ?? 1 }
  }
  const values = series.map((point) => Number(point.value || 0))
  const max = bounds?.max ?? Math.max(...values, 1)
  const min = bounds?.min ?? Math.min(...values, 0)
  const points = values.map((value, index) => {
    const x = series.length === 1 ? 0 : (index / (series.length - 1)) * width
    const y = mapSeriesValueToY(value, min, max, height)
    return `${x},${y}`
  }).join(' ')
  const areaPath = `M 0,${height - 2} L ${points.split(' ').join(' L ')} L ${width},${height - 2} Z`
  return { points, areaPath, min, max }
}

function mapSeriesValueToY(value, min, max, height) {
  const range = Math.max(max - min, 1)
  return height - ((value - min) / range) * (height - 10) - 5
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

function formatPercent(value) {
  return `${Number(value || 0).toFixed(1)}%`
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
  const parts = [`duration=${Number(draft.overrides.duration || 0)}s`]
  if (selectedTemplate.value) {
    parts.push(`${concurrencyKey.value}=${Number(draft.overrides[concurrencyKey.value] || 0)}`)
  }
  return parts.join(', ')
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
  if (!draft.database_type || !draft.connection_id) return false
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
  gap: 20px;
  color: #e8edf5;
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
.log-meta {
  margin: 4px 0 0;
  color: #94a3b8;
}

.workspace-grid {
  display: grid;
  grid-template-columns: 360px minmax(0, 1fr);
  gap: 20px;
}

.left-column,
.right-column {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.panel {
  background: #131b24;
  border: 1px solid #243446;
  border-radius: 16px;
  padding: 18px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.03);
}

.status-pill,
.handoff-badge,
.phase-tag,
.stream-tag {
  border-radius: 999px;
  padding: 6px 12px;
  font-size: 12px;
  background: rgba(120, 135, 160, 0.18);
}

.top-controls {
  display: flex;
  align-items: center;
  gap: 12px;
  position: relative;
}

.btn {
  border: 1px solid transparent;
  border-radius: 12px;
  padding: 10px 14px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.18s ease, border-color 0.18s ease, background 0.18s ease, box-shadow 0.18s ease, opacity 0.18s ease;
}

.btn:hover:not(:disabled) {
  transform: translateY(-1px);
}

.btn:disabled {
  cursor: not-allowed;
  opacity: 0.42;
  transform: none;
  box-shadow: none;
}

.btn-secondary {
  background: #111923;
  border-color: #31475d;
  color: #dce4ef;
}

.btn-primary {
  background: linear-gradient(135deg, #1b8f55, #36c275);
  color: #f4fff9;
  box-shadow: 0 12px 30px rgba(40, 167, 93, 0.28);
}

.btn-danger {
  background: linear-gradient(135deg, #962f3f, #d65262);
  color: #fff5f5;
  box-shadow: 0 12px 30px rgba(214, 82, 98, 0.24);
}

.primary-action,
.danger-action {
  min-width: 112px;
  padding: 13px 20px;
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.status-entry {
  position: relative;
}

.status-summary {
  border: 1px solid #31475d;
  border-radius: 999px;
  background: #0f151d;
  color: #edf2f7;
  padding: 8px 12px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.status-summary.ready {
  border-color: rgba(77, 208, 168, 0.55);
  color: #8ef0cc;
}

.status-summary.warning {
  border-color: rgba(244, 162, 97, 0.55);
  color: #f4c287;
}

.status-summary.blocked {
  border-color: rgba(255, 102, 102, 0.55);
  color: #ff9b9b;
}

.status-summary-badge {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.status-summary-icon {
  font-weight: 700;
}

.status-popover {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  width: 360px;
  border: 1px solid #243446;
  border-radius: 14px;
  background: #0f151d;
  padding: 14px;
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.45);
  z-index: 20;
}

.status-popover-head {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
}

.status-popover-head span {
  color: #94a3b8;
  font-size: 13px;
}

.status-popover-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.status-popover-item {
  border: 1px solid #243446;
  border-radius: 10px;
  background: #0b1118;
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-popover-item span {
  color: #7f90a7;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.status-pill.running,
.status-pill.preparing,
.status-pill.cleaning,
.status-pill.starting,
.status-pill.stopping {
  background: rgba(76, 208, 168, 0.18);
  color: #8ef0cc;
}

.status-pill.failed,
.status-pill.stopped {
  background: rgba(255, 102, 102, 0.2);
  color: #ff9b9b;
}

.status-pill.success {
  background: rgba(77, 208, 168, 0.22);
  color: #9ff4d3;
}

.status-pill.idle {
  background: rgba(120, 135, 160, 0.18);
  color: #c1ccda;
}

.status-strip {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 12px 16px;
  background: #101720;
}

.status-strip.active {
  border-color: rgba(77, 208, 168, 0.3);
}

.status-strip-main,
.status-strip-meta {
  display: flex;
  align-items: center;
  gap: 10px;
}

.status-strip-phase,
.status-strip-time,
.status-strip-meta {
  font-size: 12px;
  color: #9ba9bc;
  letter-spacing: 0.04em;
}

.status-strip-phase {
  text-transform: lowercase;
}

.status-strip-time {
  color: #dce4ef;
  font-weight: 600;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 14px;
}

.field span,
.eyebrow {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #7f90a7;
}

input,
select,
.params-preview {
  width: 100%;
  border: 1px solid #31475d;
  border-radius: 10px;
  background: #0f151d;
  color: #edf2f7;
  padding: 10px 12px;
}

.readonly-field {
  width: 100%;
  border: 1px solid #31475d;
  border-radius: 10px;
  background: linear-gradient(180deg, #0b1118, #0f151d);
  color: #edf2f7;
  padding: 10px 12px;
  min-height: 42px;
  display: flex;
  align-items: center;
}

.readonly-field.placeholder {
  color: #7f90a7;
}

.compact-help {
  margin: 14px 0 0;
  color: #94a3b8;
  font-size: 13px;
}

.select-dark {
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  background-color: #0b1118;
  background-image:
    linear-gradient(45deg, transparent 50%, #8ec5ff 50%),
    linear-gradient(135deg, #8ec5ff 50%, transparent 50%);
  background-position:
    calc(100% - 18px) calc(50% - 2px),
    calc(100% - 12px) calc(50% - 2px);
  background-size: 6px 6px, 6px 6px;
  background-repeat: no-repeat;
  border-color: #31475d;
  color: #edf2f7;
  padding-right: 34px;
}

.select-dark:hover {
  border-color: #45627f;
  background-color: #101925;
}

.select-dark:focus {
  outline: none;
  border-color: #5d88b3;
  box-shadow: 0 0 0 3px rgba(93, 136, 179, 0.18);
}

.select-dark:disabled {
  background-color: #0a0f15;
  border-color: #243446;
  color: #5f7288;
  cursor: not-allowed;
}

.select-dark option {
  background: #0b1118;
  color: #edf2f7;
}

.override-grid,
.metric-grid,
.system-grid,
.preview-grid {
  display: grid;
  gap: 14px;
}

.override-grid,
.preview-grid {
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
  border: 1px solid #243446;
  border-radius: 12px;
  padding: 16px;
  background: #0f151d;
}

.text-action {
  background: transparent;
  border: 0;
  color: #8ec5ff;
  cursor: pointer;
}

.text-action.danger {
  color: #ff8d8d;
}

.metric-card,
.system-card {
  color: #dce4ef;
}

.metric-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 260px;
  background:
    radial-gradient(circle at top right, rgba(255, 255, 255, 0.04), transparent 38%),
    linear-gradient(180deg, rgba(17, 24, 33, 0.98), rgba(9, 14, 20, 0.98));
}

.metric-card-head,
.metric-main,
.metric-stats,
.system-card-head,
.chart-legend {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.metric-card-head > span:first-child,
.system-card-head span {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #8ea0b5;
}

.metric-status {
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 11px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.metric-stable {
  color: #8ef0cc;
  border-color: rgba(77, 208, 168, 0.35);
  background: rgba(77, 208, 168, 0.12);
}

.metric-fluctuating {
  color: #f4c287;
  border-color: rgba(244, 162, 97, 0.35);
  background: rgba(244, 162, 97, 0.12);
}

.metric-sawtooth {
  color: #ff9b9b;
  border-color: rgba(248, 113, 113, 0.35);
  background: rgba(248, 113, 113, 0.12);
}

.metric-main {
  align-items: baseline;
  justify-content: flex-start;
}

.metric-main strong {
  font-size: clamp(34px, 4vw, 52px);
  line-height: 1;
  color: #f7fafc;
  letter-spacing: -0.03em;
}

.metric-main span {
  font-size: 14px;
  color: #90a4ba;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.metric-stats {
  justify-content: flex-start;
  gap: 20px;
}

.metric-stats div {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.metric-stats span {
  font-size: 11px;
  color: #6f8297;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.metric-stats strong {
  color: #dce4ef;
  font-size: 15px;
}

.metric-history,
.system-chart-wrap {
  flex: 1;
  min-height: 0;
  border-radius: 14px;
  border: 1px solid rgba(74, 90, 109, 0.55);
  background: linear-gradient(180deg, rgba(13, 19, 27, 0.92), rgba(8, 12, 18, 0.98));
  padding: 10px;
}

.history-chart,
.system-chart {
  width: 100%;
  height: 100%;
  display: block;
}

.history-baseline,
.chart-gridline {
  stroke: rgba(148, 163, 184, 0.18);
  stroke-width: 1;
  stroke-dasharray: 3 5;
}

.system-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 240px;
}

.system-card-head {
  align-items: flex-start;
}

.system-card-head strong {
  display: block;
  margin-top: 6px;
  color: #e5edf7;
  font-size: 14px;
  line-height: 1.5;
}

.chart-legend {
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 10px;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: #9fb0c4;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.legend-item i {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  display: inline-block;
}

.system-caption {
  margin: 0;
  color: #70839a;
  font-size: 12px;
}

.system-panel.disabled {
  opacity: 0.62;
}

.terminal-summary,
.terminal-viewer {
  background: #05080d;
  border-radius: 12px;
  border: 1px solid #182431;
  font-family: 'JetBrains Mono', 'SFMono-Regular', monospace;
  overflow: auto;
}

.terminal-summary {
  max-height: 200px;
  padding: 10px 12px;
}

.terminal-viewer {
  height: 420px;
  padding: 12px;
}

.terminal-line {
  display: flex;
  gap: 10px;
  padding: 2px 0;
  color: #a8b4c5;
  white-space: pre-wrap;
}

.stream-tag {
  min-width: 64px;
  text-align: center;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  background: rgba(55, 74, 95, 0.35);
}

.terminal-line.error {
  color: #ff8d8d;
}

.terminal-line.muted {
  color: #95a3b6;
}

.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(2, 6, 11, 0.78);
  display: grid;
  place-items: center;
  padding: 20px;
  z-index: 20;
}

.modal {
  width: min(720px, 100%);
  background: #111923;
  border: 1px solid #2a3e54;
  border-radius: 18px;
  padding: 20px;
}

.modal-wide {
  width: min(1100px, 100%);
}

.empty-state {
  color: #7f90a7;
}

@media (max-width: 1180px) {
  .workspace-grid,
  .metric-grid,
  .preview-grid,
  .system-grid,
  .override-grid {
    grid-template-columns: 1fr;
  }

  .tab-header,
  .top-controls,
  .status-strip {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
