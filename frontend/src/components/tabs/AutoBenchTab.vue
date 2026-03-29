<script setup>
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import {
  connectionFilterOptions,
  createAutoBenchWizardDraft,
  describeSelectedProfiles,
  policySummaryItems,
  profileOptions,
  toggleDraftConnectionSelection,
  toggleDraftProfileSelection,
  validateAutoBenchWizardDraft
} from './autobenchWizardDraft.mjs'
import * as AutoBenchBinding from '../../../wailsjs/go/bindings/AutoBenchBinding'
import { useConnectionStore } from '../../stores/connection'
import { useAppStore } from '../../stores/app'
import { useAutoBenchStore } from '../../stores/autobench'

const connectionStore = useConnectionStore()
const appStore = useAppStore()
const autobenchStore = useAutoBenchStore()

const draft = ref(createAutoBenchWizardDraft())
const activeConnectionFilter = ref('all')
const isCreating = ref(false)
const isStarting = ref(false)
const createError = ref('')
const startError = ref('')
const createdSuiteId = ref('')
const suiteStatus = ref(null)
const isPolling = ref(false)
const elapsedSeconds = ref(0)
let pollTimeout = null
let elapsedInterval = null
let consecutiveErrors = 0
const MAX_CONSECUTIVE_ERRORS = 5

// Real connections from store
const realConnections = computed(() => {
  return connectionStore.connections.map(conn => ({
    id: conn.id,
    label: conn.name,
    databaseType: conn.type,
    detail: `${conn.host}:${conn.port}`
  }))
})

const wizardValidation = computed(() => validateAutoBenchWizardDraft(draft.value))
const filteredConnections = computed(() => {
  if (activeConnectionFilter.value === 'all') {
    return realConnections.value
  }
  return realConnections.value.filter(conn => conn.databaseType === activeConnectionFilter.value)
})
const selectedProfileSummary = computed(() => describeSelectedProfiles(draft.value.selectedProfiles))

const canCreateSuite = computed(() => {
  return !isCreating.value && wizardValidation.value.canCreateSuite && !createdSuiteId.value
})

const canStartSuite = computed(() => {
  if (!createdSuiteId.value || isStarting.value) return false
  if (!suiteStatus.value) return true
  return ['draft', 'ready'].includes(suiteStatus.value.status)
})

const isSuiteRunning = computed(() => {
  return suiteStatus.value?.status === 'running'
})

const connNameMap = computed(() => {
  const map = {}
  for (const conn of connectionStore.connections) {
    map[conn.id] = conn.name
  }
  return map
})

const currentItem = computed(() => {
  if (!suiteStatus.value?.items) return null
  return suiteStatus.value.items.find(item => item.status === 'running') || null
})

const suiteSummary = computed(() => {
  if (!suiteStatus.value) return null
  const s = suiteStatus.value
  return {
    status: s.status,
    name: s.name,
    total: s.total_items,
    pending: s.pending_items,
    running: s.running_items,
    completed: s.completed_items,
    progress: s.total_items > 0 ? Math.round((s.completed_items / s.total_items) * 100) : 0
  }
})

// Elapsed timer management
watch(isSuiteRunning, (running) => {
  if (running) {
    elapsedSeconds.value = 0
    if (elapsedInterval) clearInterval(elapsedInterval)
    elapsedInterval = setInterval(() => {
      elapsedSeconds.value++
    }, 1000)
  } else {
    if (elapsedInterval) {
      clearInterval(elapsedInterval)
      elapsedInterval = null
    }
  }
})

function formatElapsed(secs) {
  const m = Math.floor(secs / 60)
  const s = secs % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

// Load connections on mount, recover running suite from store
onMounted(async () => {
  await connectionStore.fetchConnections()

  // Recover from Pinia store if a suite was active before tab switch
  if (autobenchStore.createdSuiteId) {
    createdSuiteId.value = autobenchStore.createdSuiteId
    if (autobenchStore.suiteStatus) {
      suiteStatus.value = autobenchStore.suiteStatus
    }
    // Always fetch fresh status from backend
    const ok = await fetchSuiteStatus()
    if (ok && suiteStatus.value?.status === 'running') {
      startPolling()
      // Estimate elapsed from started_at if available
      if (suiteStatus.value.started_at) {
        try {
          const startMs = new Date(suiteStatus.value.started_at).getTime()
          const nowMs = Date.now()
          elapsedSeconds.value = Math.max(0, Math.floor((nowMs - startMs) / 1000))
        } catch { /* ignore */ }
      }
      // Restart elapsed counter
      if (elapsedInterval) clearInterval(elapsedInterval)
      elapsedInterval = setInterval(() => {
        elapsedSeconds.value++
      }, 1000)
    }
  }
})

onUnmounted(() => {
  if (pollTimeout) {
    clearTimeout(pollTimeout)
  }
  if (elapsedInterval) {
    clearInterval(elapsedInterval)
  }
})

function toggleConnectionSelection(connectionId) {
  draft.value = toggleDraftConnectionSelection(draft.value, connectionId)
}

function toggleProfileSelection(profileId) {
  draft.value = toggleDraftProfileSelection(draft.value, profileId)
}

async function handleCreateSuite() {
  if (!canCreateSuite.value) return

  isCreating.value = true
  createError.value = ''

  try {
    const result = await AutoBenchBinding.CreateSuite({
      name: 'AutoBench Suite',
      connection_ids: draft.value.selectedConnectionIds,
      profile_types: draft.value.selectedProfiles,
    })

    if (result.error) {
      createError.value = result.error
    } else {
      createdSuiteId.value = result.suite_id
      autobenchStore.setSuiteId(result.suite_id)
      await fetchSuiteStatus()
    }
  } catch (err) {
    createError.value = String(err)
  } finally {
    isCreating.value = false
  }
}

async function handleStartSuite() {
  if (!canStartSuite.value) return

  isStarting.value = true
  startError.value = ''

  try {
    const result = await AutoBenchBinding.StartSuite(createdSuiteId.value)

    if (result.error) {
      startError.value = result.error
    } else {
      startPolling()
    }
  } catch (err) {
    startError.value = String(err)
  } finally {
    isStarting.value = false
  }
}

async function fetchSuiteStatus() {
  if (!createdSuiteId.value) return false

  try {
    const result = await AutoBenchBinding.GetSuiteStatus(createdSuiteId.value)
    if (!result.error) {
      suiteStatus.value = result
      autobenchStore.setSuiteStatus(result)
      consecutiveErrors = 0
      return true
    }
    consecutiveErrors++
    return false
  } catch {
    consecutiveErrors++
    return false
  }
}

function startPolling() {
  if (isPolling.value) return
  isPolling.value = true
  pollSuiteStatus()
}

function stopPolling() {
  isPolling.value = false
  if (pollTimeout) {
    clearTimeout(pollTimeout)
    pollTimeout = null
  }
}

async function pollSuiteStatus() {
  if (!createdSuiteId.value || !isPolling.value) return

  await fetchSuiteStatus()

  if (consecutiveErrors >= MAX_CONSECUTIVE_ERRORS) {
    stopPolling()
    return
  }

  if (suiteStatus.value?.status === 'running') {
    const delay = consecutiveErrors > 0 ? 1000 * Math.min(consecutiveErrors, 5) : 1000
    pollTimeout = setTimeout(pollSuiteStatus, delay)
  } else {
    stopPolling()
  }
}

function viewReport(reportId) {
  if (!reportId) return
  appStore.setActiveTab('reports')
}

function goToReports() {
  appStore.setActiveTab('reports')
}

function resetSuite() {
  createdSuiteId.value = ''
  suiteStatus.value = null
  createError.value = ''
  startError.value = ''
  elapsedSeconds.value = 0
  stopPolling()
  autobenchStore.resetSuite()
}

function getStatusLabel(status) {
  const labels = {
    draft: 'Draft',
    ready: 'Ready',
    running: 'Running',
    success: 'Success',
    partial_success: 'Partial Success',
    failed: 'Failed',
    cancelled: 'Cancelled'
  }
  return labels[status] || status
}

function getStatusClass(status) {
  if (['success'].includes(status)) return 'status-success'
  if (['failed', 'cancelled'].includes(status)) return 'status-error'
  if (['running'].includes(status)) return 'status-running'
  if (['partial_success'].includes(status)) return 'status-warning'
  return ''
}
</script>

<template>
  <section class="autobench-page">
    <header class="page-header">
      <div>
        <h1 class="page-title">AutoBench</h1>
        <p class="page-subtitle">
          Create and run benchmark suites. Monitor execution progress here.
        </p>
      </div>
      <div class="header-actions">
        <button v-if="createdSuiteId && !isSuiteRunning" class="primary-action start-action" type="button" :disabled="!canStartSuite" @click="handleStartSuite">
          {{ isStarting ? 'Starting...' : 'Start Suite' }}
        </button>
        <button v-if="isSuiteRunning" class="primary-action running-action" type="button" disabled>
          Running...
        </button>
        <button v-if="createdSuiteId && !isSuiteRunning" class="secondary-action" type="button" @click="resetSuite">
          New Suite
        </button>
      </div>
    </header>

    <div v-if="createError || startError" class="error-banner">
      <span>{{ createError || startError }}</span>
    </div>

    <!-- Active Run Section -->
    <div v-if="createdSuiteId && suiteStatus" class="active-run-section">
      <div class="run-strip">
        <span :class="['status-badge', getStatusClass(suiteStatus.status)]">
          {{ getStatusLabel(suiteStatus.status) }}
        </span>
        <span class="run-name">{{ suiteStatus.name || createdSuiteId }}</span>
        <span class="run-progress">
          {{ suiteStatus.completed_items }}/{{ suiteStatus.total_items }}
        </span>
        <span v-if="currentItem" class="run-current">
          Running: {{ connNameMap[currentItem.connection_id] || currentItem.connection_id }} - {{ currentItem.profile_type }}
        </span>
        <span v-if="isSuiteRunning" class="run-elapsed">{{ formatElapsed(elapsedSeconds) }}</span>
      </div>

      <div class="run-detail">
        <div class="status-metrics">
          <div class="metric">
            <span class="metric-label">Total</span>
            <span class="metric-value">{{ suiteStatus.total_items }}</span>
          </div>
          <div class="metric">
            <span class="metric-label">Pending</span>
            <span class="metric-value">{{ suiteStatus.pending_items }}</span>
          </div>
          <div class="metric">
            <span class="metric-label">Running</span>
            <span class="metric-value">{{ suiteStatus.running_items }}</span>
          </div>
          <div class="metric">
            <span class="metric-label">Completed</span>
            <span class="metric-value">{{ suiteStatus.completed_items }}</span>
          </div>
        </div>

        <div class="progress-bar-container">
          <div class="progress-bar" :style="{ width: suiteSummary.progress + '%' }"></div>
          <span class="progress-label">{{ suiteSummary.progress }}%</span>
        </div>

        <div v-if="suiteStatus.items && suiteStatus.items.length > 0" class="items-list">
          <div class="items-header">
            <span>Connection</span>
            <span>Type</span>
            <span>Status</span>
            <span>Report</span>
          </div>
          <div v-for="item in suiteStatus.items" :key="item.id" class="item-row">
            <span class="item-connection">{{ connNameMap[item.connection_id] || item.connection_id }}</span>
            <span class="item-type">{{ item.profile_type }}</span>
            <span :class="['item-status', getStatusClass(item.status)]">{{ item.status }}</span>
            <span class="item-report">
              <button v-if="item.report_id" class="link-button" @click="viewReport(item.report_id)">
                View Report
              </button>
              <span v-else-if="item.error_message" class="error-text" :title="item.error_message">
                Error
              </span>
              <span v-else class="muted">-</span>
            </span>
          </div>
        </div>
      </div>

      <div v-if="['success', 'partial_success', 'failed'].includes(suiteStatus.status)" class="suite-actions">
        <button class="primary-action" @click="goToReports">
          View All Reports
        </button>
      </div>
    </div>

    <!-- Wizard Two-Column Layout -->
    <div v-if="!createdSuiteId" class="wizard-columns">
      <div class="wizard-col-left">
        <div class="section-header-row">
          <h3>Connections</h3>
          <span class="wizard-chip">{{ draft.selectedConnectionIds.length }}</span>
        </div>

        <div class="filter-block">
          <div class="filter-options">
            <button
              v-for="option in connectionFilterOptions"
              :key="option.id"
              type="button"
              :class="['filter-pill', { active: activeConnectionFilter === option.id }]"
              @click="activeConnectionFilter = option.id"
            >
              {{ option.label }}
            </button>
          </div>
        </div>

        <div v-if="connectionStore.loading" class="loading-state">
          Loading connections...
        </div>
        <div v-else-if="realConnections.length === 0" class="empty-state">
          No connections available. Create connections first.
        </div>
        <template v-else>
          <label
            v-for="connection in filteredConnections"
            :key="connection.id"
            class="conn-row"
            :class="{ selected: draft.selectedConnectionIds.includes(connection.id) }"
          >
            <input
              type="checkbox"
              :checked="draft.selectedConnectionIds.includes(connection.id)"
              @change="toggleConnectionSelection(connection.id)"
            >
            <span class="conn-row-name">{{ connection.label }}</span>
            <span class="conn-row-type">{{ connection.databaseType }}</span>
            <span class="conn-row-host">{{ connection.detail }}</span>
          </label>
        </template>
        <p v-if="wizardValidation.connectionError" class="wizard-validation">{{ wizardValidation.connectionError }}</p>
      </div>

      <div class="wizard-col-right">
        <section>
          <div class="section-header-row">
            <h3>Benchmark Types</h3>
            <span class="wizard-chip">{{ draft.selectedProfiles.length }}</span>
          </div>

          <div class="profile-toggles">
            <button
              v-for="profile in profileOptions"
              :key="profile.id"
              type="button"
              :class="['profile-toggle', { active: draft.selectedProfiles.includes(profile.id) }]"
              @click="toggleProfileSelection(profile.id)"
            >
              {{ profile.label }}
            </button>
          </div>
          <p class="wizard-order">Selected order: {{ selectedProfileSummary }}</p>
          <p v-if="wizardValidation.profileError" class="wizard-validation">{{ wizardValidation.profileError }}</p>
        </section>

        <section>
          <div class="section-header-row">
            <h3>Execution Policy</h3>
            <span class="wizard-chip">Default</span>
          </div>
          <dl class="policy-summary">
            <div v-for="item in policySummaryItems" :key="item.label" class="policy-row">
              <dt>{{ item.label }}</dt>
              <dd>{{ item.value }}</dd>
            </div>
          </dl>
        </section>

        <div class="wizard-actions">
          <button class="primary-action" type="button" :disabled="!canCreateSuite" @click="handleCreateSuite">
            {{ isCreating ? 'Creating...' : 'Create Suite' }}
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.autobench-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.page-title {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.page-subtitle {
  margin-top: 8px;
  color: var(--text-secondary);
  max-width: 720px;
  line-height: 1.5;
}

.header-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.primary-action {
  border: 1px solid var(--primary);
  border-radius: var(--radius-md);
  padding: 10px 14px;
  background: var(--primary);
  color: white;
  cursor: pointer;
  box-shadow: none;
  font-weight: 500;
}

.primary-action:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.secondary-action {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 10px 14px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  cursor: pointer;
  box-shadow: none;
  font-weight: 500;
}

.start-action {
  background: var(--success);
  border-color: var(--success);
}

.running-action {
  background: var(--warning);
  border-color: var(--warning);
}

.error-banner {
  background: var(--danger-bg);
  border: 1px solid var(--danger);
  border-radius: var(--radius-md);
  padding: 12px 16px;
  color: var(--danger);
}

/* Active Run Section */
.active-run-section {
  border: 1px solid var(--primary);
  border-left: 4px solid var(--primary);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
  background: var(--bg-primary);
}

.run-strip {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.run-name {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 16px;
}

.run-progress {
  font-family: var(--font-family-mono);
  font-size: 14px;
  color: var(--text-secondary);
}

.run-current {
  color: var(--primary);
  font-size: 13px;
}

.run-elapsed {
  font-family: var(--font-family-mono);
  font-size: 14px;
  color: var(--text-secondary);
  margin-left: auto;
}

.run-detail {
  margin-top: 12px;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 500;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
}

.status-success {
  background: var(--success-bg);
  color: var(--success);
  border-color: var(--success);
}

.status-error {
  background: var(--danger-bg);
  color: var(--danger);
  border-color: var(--danger);
}

.status-running {
  background: var(--primary-bg);
  color: var(--primary);
  border-color: var(--primary);
}

.status-warning {
  background: var(--warning-bg);
  color: var(--warning);
  border-color: var(--warning);
}

.status-metrics {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

.metric {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.metric-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.metric-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
}

.progress-bar-container {
  position: relative;
  height: 24px;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  overflow: hidden;
  margin-bottom: 20px;
}

.progress-bar {
  height: 100%;
  background: var(--primary);
  transition: width 0.3s ease;
}

.progress-label {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  font-weight: 500;
  color: var(--text-primary);
}

.items-list {
  margin-top: 16px;
}

.items-list h4 {
  margin: 0 0 12px 0;
  color: var(--text-primary);
}

.items-header {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr;
  gap: 12px;
  padding: 8px 12px;
  background: var(--bg-secondary);
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 500;
}

.item-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 1fr;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  margin-top: 8px;
  font-size: 13px;
}

.item-connection {
  color: var(--text-primary);
  font-weight: 500;
}

.item-type {
  color: var(--text-secondary);
}

.item-status {
  text-transform: capitalize;
}

.link-button {
  background: none;
  border: none;
  color: var(--primary);
  cursor: pointer;
  padding: 0;
  font-size: 13px;
  text-decoration: underline;
}

.link-button:hover {
  color: var(--primary-dark);
}

.error-text {
  color: var(--danger);
}

.muted {
  color: var(--text-muted);
}

.suite-actions {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}

/* Two-column wizard layout */
.wizard-columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-xl);
}

@media (max-width: 768px) {
  .wizard-columns {
    grid-template-columns: 1fr;
  }
}

.wizard-col-left,
.wizard-col-right {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.section-header-row h3 {
  font-size: 15px;
  color: var(--text-primary);
  margin: 0;
}

.wizard-chip {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  font-size: 12px;
}

.filter-block {
  margin-top: 4px;
}

.filter-options {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.filter-pill {
  border: 1px solid var(--border-color);
  border-radius: 999px;
  background: var(--bg-primary);
  color: var(--text-secondary);
  padding: 6px 10px;
  cursor: pointer;
}

.filter-pill.active {
  color: var(--primary);
  border-color: var(--primary);
}

/* Compact connection row */
.conn-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 10px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  cursor: pointer;
  margin-bottom: 4px;
  background: var(--bg-primary);
}

.conn-row:hover {
  background: var(--bg-hover);
}

.conn-row.selected {
  border-color: var(--primary);
  background: var(--primary-light);
}

.conn-row-name {
  font-weight: 500;
  min-width: 80px;
  color: var(--text-primary);
  font-size: 14px;
}

.conn-row-type {
  padding: 1px 6px;
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  background: var(--bg-secondary);
  color: var(--text-secondary);
  text-transform: uppercase;
}

.conn-row-host {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-sm);
  color: var(--text-muted);
}

/* Profile toggle pills */
.profile-toggles {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 4px;
}

.profile-toggle {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  cursor: pointer;
  background: var(--bg-primary);
  color: var(--text-secondary);
  font-size: 14px;
}

.profile-toggle.active {
  border-color: var(--primary);
  background: var(--primary-light);
  color: var(--primary);
}

.wizard-order {
  margin-top: 12px;
  color: var(--text-secondary);
  line-height: 1.5;
  font-size: 13px;
}

.wizard-validation {
  margin-top: 10px;
  color: var(--danger);
  font-size: 13px;
}

.loading-state,
.empty-state {
  margin-top: 12px;
  padding: 16px;
  text-align: center;
  color: var(--text-secondary);
  background: var(--bg-primary);
  border-radius: var(--radius-md);
}

.policy-summary {
  margin-top: 14px;
  display: grid;
  gap: 10px;
}

.policy-row {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 10px;
  border-bottom: 1px dashed var(--border-color);
}

.policy-row dt {
  color: var(--text-secondary);
}

.policy-row dd {
  color: var(--text-primary);
  font-weight: 600;
  text-align: right;
}

.wizard-actions {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .status-metrics {
    grid-template-columns: repeat(2, 1fr);
  }

  .items-header,
  .item-row {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
