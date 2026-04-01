<template>
  <div class="reports-tab">
    <!-- Page Header -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">Reports</h1>
        <p class="page-subtitle">View benchmark reports grouped by suite</p>
      </div>
      <div class="header-right">
        <button v-if="checkedIds.length > 0" class="btn btn-danger" @click="handleDeleteSelected">
          Delete Selected ({{ checkedIds.length }})
        </button>
        <button v-if="allReportIds.length > 0" class="btn btn-danger-outline" @click="handleClearAll">
          Clear All
        </button>
        <button class="btn" @click="refreshReports">
          Refresh
        </button>
      </div>
    </div>

    <!-- Main Content -->
    <div class="reports-content">
      <!-- Reports List -->
      <div class="reports-list-section">
        <!-- Filter Bar -->
        <div class="filter-bar">
          <select v-model="statusFilter" class="filter-select" @change="onFilterChange">
            <option value="">All</option>
            <option value="success">成功</option>
            <option value="failed">失败</option>
            <option value="cancelled">Stop</option>
          </select>
        </div>

        <!-- Loading State -->
        <div v-if="reportStore.loading" class="loading-state">
          <div class="spinner"></div>
          <span>Loading reports...</span>
        </div>

        <!-- Empty State -->
        <div v-else-if="allReportIds.length === 0" class="empty-state">
          <div class="empty-state-icon">📋</div>
          <p class="empty-state-title">No benchmark reports</p>
          <p class="empty-state-description">Run a benchmark task to see results appear here.</p>
        </div>

        <!-- Grouped Reports List -->
        <div v-else class="reports-list">
          <div class="select-all-row" @click="toggleAll">
            <input type="checkbox" class="checkbox" :checked="allChecked" :indeterminate.prop="someChecked && !allChecked" />
            <span class="select-all-label">{{ allChecked ? 'Deselect All' : 'Select All' }}</span>
          </div>

          <template v-for="group in reportGroups" :key="group.key">
            <!-- Suite Group Row -->
            <div v-if="group.isSuite" class="suite-group">
              <div class="suite-row" @click="toggleGroup(group.key)">
                <span class="expand-icon">{{ expandedGroups[group.key] ? '▾' : '▸' }}</span>
                <input type="checkbox" class="checkbox" :checked="isGroupChecked(group)" @click.stop="toggleGroupCheck(group)" />
                <div class="suite-info">
                  <span class="suite-name">{{ group.name }}</span>
                  <span class="suite-meta">
                    {{ formatSuiteProgress(group.progress) }} · {{ formatDate(group.startedAt) }}
                  </span>
                </div>
                <span class="status-badge" :class="getStatusClass(group.status)">{{ getStatusText(group.status) }}</span>
                <button class="delete-btn" title="Delete suite" @click.stop="handleDeleteGroup(group)">x</button>
              </div>
              <!-- Expanded sub-items -->
              <div v-if="expandedGroups[group.key]" class="suite-items">
                <div
                  v-for="report in group.reports"
                  :key="report.id"
                  class="report-item sub-item"
                  :class="{ selected: selectedReportId === report.id, checked: checkedIds.includes(report.id) }"
                  @click="selectReport(report.id)"
                >
                  <input type="checkbox" class="checkbox" :checked="checkedIds.includes(report.id)" @click.stop="toggleCheck(report.id)" />
                  <div class="report-info">
                    <span class="sub-connection">{{ report.connection_name || report.database_type || '-' }}</span>
                    <span class="sub-type">{{ report.template_name || report.source_type }}</span>
                  </div>
                  <div class="report-metrics-compact">
                    <span v-if="report.tpm">TPM {{ formatNumber(report.tpm) }}</span>
                    <span v-if="report.tps">TPS {{ formatNumber(report.tps) }}</span>
                    <span v-if="report.latency_avg_ms">{{ formatLatency(report.latency_avg_ms) }}</span>
                  </div>
                  <span class="status-badge small" :class="getStatusClass(report.status)">{{ getStatusText(report.status) }}</span>
                  <button v-if="canViewReport(report.status)" class="view-btn small" title="View report" @click.stop="selectReport(report.id)">View</button>
                  <span v-else-if="report.status === 'running' || report.status === 'pending' || report.status === 'draft' || report.status === 'ready'" class="status-text">Running</span>
                  <span v-else-if="report.status === 'failed'" class="status-text error-text" :title="report.error_message || ''">Failed</span>
                  <button class="delete-btn small" title="Delete report" @click.stop="handleDelete(report.id)">x</button>
                </div>
              </div>
            </div>

            <!-- Standalone Report Row -->
            <div v-else class="report-item standalone" :class="{ selected: selectedReportId === group.reports[0]?.id, checked: checkedIds.includes(group.reports[0]?.id) }" @click="selectReport(group.reports[0]?.id)">
              <input type="checkbox" class="checkbox" :checked="checkedIds.includes(group.reports[0]?.id)" @click.stop="toggleCheck(group.reports[0]?.id)" />
              <div class="report-info">
                <div class="report-name">
                  {{ group.reports[0]?.template_name || group.reports[0]?.database_type || 'Single Benchmark' }}
                </div>
                <div class="report-meta">
                  <span class="report-date">{{ formatDate(group.reports[0]?.started_at) }}</span>
                  <span class="report-source">Single</span>
                </div>
              </div>
              <div class="report-metrics">
                <div class="metric">
                  <span class="metric-label">TPM</span>
                  <span class="metric-value">{{ formatNumber(group.reports[0]?.tpm) }}</span>
                </div>
                <div class="metric">
                  <span class="metric-label">TPS</span>
                  <span class="metric-value">{{ formatNumber(group.reports[0]?.tps) }}</span>
                </div>
                <div class="metric">
                  <span class="metric-label">Latency</span>
                  <span class="metric-value">{{ formatLatency(group.reports[0]?.latency_avg_ms) }}</span>
                </div>
              </div>
              <div class="report-status">
                <span class="status-badge" :class="getStatusClass(group.reports[0]?.status)">{{ getStatusText(group.reports[0]?.status) }}</span>
                <button class="delete-btn" title="Delete report" @click.stop="handleDelete(group.reports[0]?.id)">x</button>
              </div>
            </div>
          </template>
        </div>

        <!-- Pagination -->
        <div v-if="allReportIds.length > 0" class="pagination">
          <span class="pagination-info">{{ reportStore.pagination.total }} reports</span>
        </div>
      </div>

      <!-- Detail Panel -->
      <div class="detail-section" v-if="selectedReportId">
        <ReportDetailPanel
          :report-id="selectedReportId"
          @close="closeDetail"
        />
      </div>
    </div>

    <!-- Notice Toast -->
    <div v-if="reportStore.notice" class="notice-toast" :class="`notice-${reportStore.notice.tone}`">
      {{ reportStore.notice.message }}
      <button class="notice-close" @click="reportStore.clearNotice()">×</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onActivated, computed, watch } from 'vue'
import { useReportStore } from '../../stores/report'
import { useAppStore } from '../../stores/app'
import ReportDetailPanel from '../report/ReportDetailPanel.vue'

const reportStore = useReportStore()
const appStore = useAppStore()
const selectedReportId = ref(null)
const statusFilter = ref(reportStore.filters.status || '')
const checkedIds = ref([])
const expandedGroups = ref({})

// Map backend status values to frontend filter values
const backendToFrontend = { completed: 'success', cancelled: 'cancelled', failed: 'failed' }

// Keep statusFilter in sync with store on tab remount
watch(() => reportStore.filters.status, (val) => {
  const frontendVal = backendToFrontend[val] || val || ''
  if (frontendVal !== statusFilter.value) statusFilter.value = frontendVal
})

// All report IDs for select-all
const allReportIds = computed(() => reportStore.reports.map((r) => r.id))

const allChecked = computed(() => allReportIds.value.length > 0 && checkedIds.value.length === allReportIds.value.length)
const someChecked = computed(() => checkedIds.value.length > 0 && checkedIds.value.length < allReportIds.value.length)

// Group reports by suite_id. Suite reports stay grouped, standalone shown separately.
const reportGroups = computed(() => {
  const groups = []
  const bySuite = {}

  // Partition reports into AutoBench (grouped) vs standalone
  for (const report of reportStore.reports) {
    const isAutoBench = report.source_type === 'autobench' ||
      (report.suite_id && report.suite_id !== 'standalone')

    if (isAutoBench) {
      const sid = report.suite_id || 'unknown'
      if (!bySuite[sid]) {
        bySuite[sid] = []
      }
      bySuite[sid].push(report)
    } else {
      // Standalone: source_type is not "autobench" and no meaningful suite_id
      if (!bySuite['standalone']) {
        bySuite['standalone'] = []
      }
      bySuite['standalone'].push(report)
    }
  }

  // Build suite groups from the suite data
  const suiteMap = {}
  for (const s of reportStore.suites) {
    suiteMap[s.id] = s
  }

  // Suites first (ordered by suites array)
  for (const suite of reportStore.suites) {
    const reports = bySuite[suite.id] || []
    if (reports.length > 0) {
      const { status: aggregatedStatus, progress } = analyzeGroupReports(reports)
      groups.push({
        key: suite.id,
        isSuite: true,
        name: suite.name || `AutoBench Suite ${formatDate(suite.started_at)}`,
        status: aggregatedStatus,
        startedAt: suite.started_at,
        reports,
        progress
      })
      delete bySuite[suite.id]
    }
  }

  // Remaining suite groups (suites not in store but reports have suite_id)
  for (const [sid, reports] of Object.entries(bySuite)) {
    if (sid === 'standalone') continue
    const { status: orphanStatus, progress: orphanProgress } = analyzeGroupReports(reports)
    groups.push({
      key: sid,
      isSuite: true,
      name: reports[0]?.template_name || `Suite ${sid.slice(0, 8)}`,
      status: orphanStatus,
      startedAt: reports[0]?.started_at,
      reports,
      progress: orphanProgress
    })
  }

  // Standalone reports (only non-AutoBench reports)
  const standalone = bySuite['standalone'] || []
  for (const report of standalone) {
    groups.push({
      key: report.id,
      isSuite: false,
      reports: [report]
    })
  }

  return groups
})

// analyzeGroupReports computes both aggregate status and progress in a single pass.
function analyzeGroupReports(reports) {
  const total = reports.length
  let completed = 0
  let running = 0
  let failed = 0
  let cancelled = 0
  let partial = false

  for (const r of reports) {
    const s = r.status
    if (s === 'success' || s === 'completed') completed++
    else if (s === 'running' || s === 'pending' || s === 'draft' || s === 'ready') running++
    else if (s === 'failed') failed++
    else if (s === 'cancelled' || s === 'stopped' || s === 'interrupted') cancelled++
    else if (s === 'partial_success') partial = true
  }

  let status = 'success'
  if (running > 0) status = 'running'
  else if (failed > 0) status = 'failed'
  else if (cancelled > 0) status = 'cancelled'
  else if (partial) status = 'partial_success'

  return { status, progress: { total, completed, running, failed } }
}

function canViewReport(status) {
  return status === 'completed' || status === 'success'
}

function formatSuiteProgress(progress) {
  const parts = [`${progress.total} items`]
  if (progress.completed > 0) parts.push(`${progress.completed} completed`)
  if (progress.running > 0) parts.push(`${progress.running} running`)
  if (progress.failed > 0) parts.push(`${progress.failed} failed`)
  return parts.join(' · ')
}

onMounted(async () => {
  await refreshReports()
  const pendingId = appStore.consumePendingReportId()
  if (pendingId) autoSelectReport(pendingId)
})

// Refresh data when switching back to this tab (KeepAlive re-activation)
onActivated(async () => {
  await refreshReports()
  const pendingId = appStore.consumePendingReportId()
  if (pendingId) autoSelectReport(pendingId)
})

// Also watch for pending report IDs set while this tab is already mounted
// (e.g., user clicks View Report on AutoBench tab while Reports tab is already mounted)
watch(() => appStore.pendingReportId, (newId) => {
  if (newId) {
    const id = appStore.consumePendingReportId()
    if (id) autoSelectReport(id)
  }
})

// Auto-select a report and expand its parent group
function autoSelectReport(reportId) {
  if (!reportId) return
  selectedReportId.value = reportId
  // Find which group contains this report and expand it
  for (const group of reportGroups.value) {
    if (group.reports.some(r => r.id === reportId)) {
      expandedGroups.value[group.key] = true
      break
    }
  }
}

const refreshReports = async () => {
  await Promise.all([
    reportStore.fetchReports(),
    reportStore.fetchSuites()
  ])
}

const selectReport = (id) => {
  if (!id) return
  selectedReportId.value = id
}

const closeDetail = () => {
  selectedReportId.value = null
}

const onFilterChange = async () => {
  // Map user-visible filter to backend status values
  let backendFilter = statusFilter.value
  if (statusFilter.value === 'success') {
    backendFilter = 'completed'
  } else if (statusFilter.value === 'cancelled') {
    backendFilter = 'cancelled'
  }
  reportStore.setFilter('status', backendFilter)
  await refreshReports()
}

const toggleGroup = (key) => {
  expandedGroups.value[key] = !expandedGroups.value[key]
}

const formatDate = (date) => {
  if (!date) return ''
  try {
    return new Date(date).toLocaleString()
  } catch {
    return date
  }
}

const formatNumber = (num) => {
  if (num === null || num === undefined) return 'N/A'
  return typeof num === 'number' ? num.toFixed(2) : num
}

const formatLatency = (ms) => {
  if (ms === null || ms === undefined) return 'N/A'
  return `${ms.toFixed(2)}ms`
}

const getStatusClass = (status) => {
  const classMap = {
    completed: 'status-success',
    success: 'status-success',
    failed: 'status-error',
    cancelled: 'status-warning',
    stopped: 'status-warning',
    interrupted: 'status-warning',
    running: 'status-info',
    pending: 'status-default',
    partial_success: 'status-warning',
    draft: 'status-default',
    ready: 'status-default'
  }
  return classMap[status] || 'status-default'
}

const getStatusText = (status) => {
  const textMap = {
    completed: '成功',
    success: '成功',
    failed: '失败',
    cancelled: 'Stop',
    stopped: 'Stop',
    interrupted: 'Stop',
    running: '运行中',
    pending: '等待中',
    partial_success: '部分成功',
    draft: '草稿',
    ready: '就绪'
  }
  return textMap[status] || status
}

const handleDelete = async (id) => {
  if (!id || !confirm('Delete this report?')) return
  await reportStore.deleteReport(id)
  checkedIds.value = checkedIds.value.filter((cid) => cid !== id)
}

const handleDeleteGroup = async (group) => {
  const ids = group.reports.map((r) => r.id)
  if (ids.length === 0) return
  if (!confirm(`Delete all ${ids.length} report(s) in this group?`)) return
  await reportStore.deleteSelectedReports(ids)
  checkedIds.value = checkedIds.value.filter((cid) => !ids.includes(cid))
}

const toggleCheck = (id) => {
  if (!id) return
  const idx = checkedIds.value.indexOf(id)
  if (idx >= 0) {
    checkedIds.value.splice(idx, 1)
  } else {
    checkedIds.value.push(id)
  }
}

const isGroupChecked = (group) => {
  const ids = group.reports.map((r) => r.id)
  return ids.length > 0 && ids.every((id) => checkedIds.value.includes(id))
}

const toggleGroupCheck = (group) => {
  const ids = group.reports.map((r) => r.id)
  const allIn = ids.every((id) => checkedIds.value.includes(id))
  if (allIn) {
    checkedIds.value = checkedIds.value.filter((id) => !ids.includes(id))
  } else {
    for (const id of ids) {
      if (!checkedIds.value.includes(id)) {
        checkedIds.value.push(id)
      }
    }
  }
}

const toggleAll = () => {
  if (allChecked.value) {
    checkedIds.value = []
  } else {
    checkedIds.value = [...allReportIds.value]
  }
}

const handleDeleteSelected = async () => {
  if (checkedIds.value.length === 0) return
  if (!confirm(`Delete ${checkedIds.value.length} selected report(s)?`)) return
  const ids = [...checkedIds.value]
  await reportStore.deleteSelectedReports(ids)
  checkedIds.value = []
}

const handleClearAll = async () => {
  if (!confirm('Delete ALL reports? This cannot be undone.')) return
  await reportStore.deleteAllReports()
  checkedIds.value = []
}
</script>

<style scoped>
.reports-tab {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* Page Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding-bottom: var(--spacing-lg);
  border-bottom: 1px solid var(--border-light);
  margin-bottom: var(--spacing-lg);
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  font-size: var(--font-size-title);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.page-subtitle {
  font-size: var(--font-size-md);
  color: var(--text-muted);
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.btn {
  padding: 8px 16px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background-color: var(--bg-primary);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn:hover {
  background-color: var(--bg-secondary);
  border-color: var(--border-dark);
}

/* Main Content */
.reports-content {
  flex: 1;
  display: flex;
  gap: var(--spacing-lg);
  overflow: hidden;
}

.reports-list-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 300px;
  overflow: hidden;
}

/* Filter Bar */
.filter-bar {
  display: flex;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
}

.filter-select {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background-color: var(--bg-primary);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  cursor: pointer;
}

/* Loading State */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: var(--text-muted);
  gap: var(--spacing-md);
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Reports List */
.reports-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

/* Suite Group */
.suite-group {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.suite-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background-color: var(--bg-secondary);
  cursor: pointer;
  overflow: hidden;
  min-width: 0;
}

.suite-row:hover {
  background-color: var(--bg-hover);
}

.expand-icon {
  font-size: 12px;
  color: var(--text-secondary);
  width: 16px;
  flex-shrink: 0;
}

.suite-info {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: baseline;
  gap: 10px;
  overflow: hidden;
  overflow: hidden;
}

.suite-name {
  font-weight: 600;
  color: var(--text-primary);
  font-size: var(--font-size-base);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.suite-meta {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
  white-space: nowrap;
}

/* Sub-items in expanded suite */
.suite-items {
  border-top: 1px solid var(--border-color);
}

.report-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-md);
  background-color: var(--bg-primary);
  gap: var(--spacing-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.report-item:hover {
  border-color: var(--border-dark);
  background-color: var(--bg-hover);
}

.report-item.selected {
  border-color: var(--primary);
  background-color: var(--primary-light);
}

.report-item.checked {
  background-color: var(--primary-light);
  border-color: var(--primary);
}

.report-item.sub-item {
  padding: 8px 12px 8px 40px;
  border: none;
  border-bottom: 1px solid var(--border-light);
  border-radius: 0;
}

.report-item.sub-item:last-child {
  border-bottom: none;
}

.report-item.standalone {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
}

.report-info {
  flex: 1;
  min-width: 0;
}

.report-name {
  font-weight: 500;
  color: var(--text-primary);
  font-size: var(--font-size-base);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.report-meta {
  display: flex;
  gap: var(--spacing-sm);
  margin-top: 4px;
}

.report-date {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
}

.report-source {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
  padding: 2px 6px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-sm);
}

.sub-connection {
  font-weight: 500;
  color: var(--text-primary);
  font-size: 13px;
}

.sub-type {
  color: var(--text-secondary);
  font-size: 12px;
  margin-left: 8px;
  padding: 1px 6px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-sm);
}

.report-metrics-compact {
  display: flex;
  gap: 10px;
  font-family: var(--font-family-mono);
  font-size: 12px;
  color: var(--text-secondary);
}

.report-metrics {
  display: flex;
  gap: var(--spacing-lg);
}

.metric {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.metric-label {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.metric-value {
  font-size: var(--font-size-md);
  font-weight: 600;
  color: var(--text-primary);
  font-family: var(--font-family-mono);
}

.report-status {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.delete-btn {
  background: none;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  padding: 0;
  flex-shrink: 0;
}

.delete-btn:hover {
  color: var(--danger);
  border-color: var(--danger);
  background-color: var(--danger-bg);
}

.delete-btn.small {
  width: 20px;
  height: 20px;
  font-size: 12px;
}

.view-btn.small {
  padding: 2px 8px;
  border: 1px solid var(--primary);
  border-radius: var(--radius-sm);
  background: var(--primary-light);
  color: var(--primary);
  font-size: 10px;
  font-weight: 600;
  cursor: pointer;
  flex-shrink: 0;
}

.view-btn.small:hover {
  background: var(--primary);
  color: white;
}

.status-text {
  font-size: 11px;
  color: var(--text-muted);
  flex-shrink: 0;
}

.status-text.error-text {
  color: var(--danger);
}

.btn-danger {
  padding: 8px 16px;
  border: 1px solid var(--danger);
  border-radius: var(--radius-md);
  background-color: var(--danger);
  color: white;
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-danger:hover {
  opacity: 0.9;
}

.btn-danger-outline {
  padding: 8px 16px;
  border: 1px solid var(--danger);
  border-radius: var(--radius-md);
  background-color: transparent;
  color: var(--danger);
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-danger-outline:hover {
  background-color: var(--danger-bg);
}

.select-all-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  cursor: pointer;
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
  color: var(--text-muted);
}

.select-all-row:hover {
  background-color: var(--bg-hover);
}

.select-all-label {
  user-select: none;
}

.checkbox {
  width: 16px;
  height: 16px;
  cursor: pointer;
  flex-shrink: 0;
  accent-color: var(--primary);
}

.status-badge {
  padding: 4px 12px;
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  flex-shrink: 0;
}

.status-badge.small {
  padding: 2px 8px;
  font-size: 10px;
}

.status-success {
  background-color: var(--success-bg);
  color: var(--success);
}

.status-error {
  background-color: var(--danger-bg);
  color: var(--danger);
}

.status-warning {
  background-color: var(--warning-bg);
  color: var(--warning);
}

.status-info {
  background-color: var(--primary-light);
  color: var(--primary);
}

.status-default {
  background-color: var(--bg-secondary);
  color: var(--text-muted);
}

/* Pagination */
.pagination {
  display: flex;
  justify-content: center;
  padding: var(--spacing-md) 0;
}

.pagination-info {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
}

/* Detail Section */
.detail-section {
  width: 400px;
  flex-shrink: 0;
  border-left: 1px solid var(--border-light);
  overflow: hidden;
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.empty-state-icon {
  font-size: 48px;
  margin-bottom: var(--spacing-lg);
  opacity: 0.5;
}

.empty-state-title {
  font-size: var(--font-size-lg);
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: var(--spacing-sm);
}

.empty-state-description {
  font-size: var(--font-size-md);
  color: var(--text-muted);
  max-width: 400px;
}

/* Notice Toast */
.notice-toast {
  position: fixed;
  bottom: 20px;
  right: 20px;
  padding: 12px 20px;
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1000;
}

.notice-info {
  background-color: var(--primary);
  color: white;
}

.notice-success {
  background-color: var(--success);
  color: white;
}

.notice-warning {
  background-color: var(--warning);
  color: white;
}

.notice-error {
  background-color: var(--danger);
  color: white;
}

.notice-close {
  background: none;
  border: none;
  color: inherit;
  font-size: 18px;
  cursor: pointer;
  padding: 0;
  line-height: 1;
  opacity: 0.8;
}

.notice-close:hover {
  opacity: 1;
}
</style>
