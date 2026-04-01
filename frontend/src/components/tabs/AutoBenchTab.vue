<script setup>
import { computed, ref, onMounted, onUnmounted, onActivated, watch } from 'vue'
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

const TERMINAL_SUITE_STATUSES = ['success', 'partial_success', 'failed', 'cancelled']

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

const isSuitePaused = ref(false)

const isSuiteRunning = computed(() => {
  return suiteStatus.value?.status === 'running'
})

const isSuiteTerminal = computed(() => {
  if (!suiteStatus.value || !createdSuiteId.value) return false
  return TERMINAL_SUITE_STATUSES.includes(suiteStatus.value.status)
})

const connNameMap = computed(() => {
  const map = {}
  for (const conn of connectionStore.connections) {
    map[conn.id] = conn.name
  }
  return map
})

const connCapabilitiesMap = computed(() => {
  const map = {}
  for (const conn of connectionStore.connections) {
    const caps = []
    if (conn.ssh_enabled) caps.push('SSH')
    if (conn.winrm_enabled) caps.push('WinRM')
    // Only show AI if there are assistants with a configured provider AND valid api_key or model
    if (conn.ai_assistants && conn.ai_assistants.some(a => a.provider && (a.api_key || a.model))) caps.push('AI')
    map[conn.id] = caps
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

// Recover elapsed time from store timestamp
function recoverElapsed() {
  if (autobenchStore.startedAtTimestamp) {
    elapsedSeconds.value = Math.max(0, Math.floor((Date.now() - autobenchStore.startedAtTimestamp) / 1000))
  }
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
    // Check pause state
    isSuitePaused.value = autobenchStore.isPolling === false && suiteStatus.value?.status === 'running'
    // Always fetch fresh status from backend
    const ok = await fetchSuiteStatus()
    if (ok && suiteStatus.value?.status === 'running') {
      startPolling()
      // Recover elapsed from store timestamp
      recoverElapsed()
      // Restart elapsed counter
      if (elapsedInterval) clearInterval(elapsedInterval)
      elapsedInterval = setInterval(() => {
        elapsedSeconds.value++
      }, 1000)
    }
  }
})

// Refresh data when switching back to this tab (KeepAlive re-activation)
onActivated(async () => {
  await connectionStore.fetchConnections()
  if (createdSuiteId.value) {
    const ok = await fetchSuiteStatus()
    if (ok && suiteStatus.value?.status === 'running') {
      startPolling()
      recoverElapsed()
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
  await startSuiteAndPoll(createdSuiteId.value)
}

async function startSuiteAndPoll(suiteId) {
  isStarting.value = true
  startError.value = ''

  try {
    const result = await AutoBenchBinding.StartSuite(suiteId)

    if (result.error) {
      startError.value = result.error
    } else {
      autobenchStore.setStartedAtTimestamp(Date.now())
      elapsedSeconds.value = 0
      startPolling()
      if (elapsedInterval) clearInterval(elapsedInterval)
      elapsedInterval = setInterval(() => {
        elapsedSeconds.value++
      }, 1000)
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

function goToReports() {
  appStore.setActiveTab('reports')
}

async function handleReRunSuite() {
  if (!suiteStatus.value) return

  // Extract the original configuration from the completed suite
  const originalConnections = []
  const originalProfiles = new Set()
  for (const item of suiteStatus.value.items || []) {
    if (item.connection_id && !originalConnections.includes(item.connection_id)) {
      originalConnections.push(item.connection_id)
    }
    if (item.profile_type) originalProfiles.add(item.profile_type)
  }

  if (originalConnections.length === 0 || originalProfiles.size === 0) return

  // Save suite name before resetSuite() clears suiteStatus to null
  const suiteName = suiteStatus.value.name || 'Re-run Suite'

  // Reset state
  resetSuite()

  // Create new suite with the same configuration
  isCreating.value = true
  createError.value = ''

  try {
    const result = await AutoBenchBinding.CreateSuite({
      name: suiteName,
      connection_ids: originalConnections,
      profile_types: [...originalProfiles],
    })

    if (result.error) {
      createError.value = result.error
      return
    }

    createdSuiteId.value = result.suite_id
    autobenchStore.setSuiteId(result.suite_id)

    // Immediately start the new suite (reuses handleStartSuite logic)
    await startSuiteAndPoll(result.suite_id)
  } catch (err) {
    createError.value = String(err)
  } finally {
    isCreating.value = false
  }
}

function resetSuite() {
  createdSuiteId.value = ''
  suiteStatus.value = null
  createError.value = ''
  startError.value = ''
  elapsedSeconds.value = 0
  isSuitePaused.value = false
  stopPolling()
  autobenchStore.resetSuite()
}

async function handlePauseSuite() {
  if (!createdSuiteId.value) return
  try {
    const result = await AutoBenchBinding.PauseSuite(createdSuiteId.value)
    if (result.error) {
      startError.value = result.error
    } else {
      isSuitePaused.value = true
      stopPolling()
    }
  } catch (err) {
    startError.value = String(err)
  }
}

async function handleResumeSuite() {
  if (!createdSuiteId.value) return
  try {
    const result = await AutoBenchBinding.ResumeSuite(createdSuiteId.value)
    if (result.error) {
      startError.value = result.error
    } else {
      isSuitePaused.value = false
      // Re-sync elapsed from store timestamp
      recoverElapsed()
      startPolling()
      if (elapsedInterval) clearInterval(elapsedInterval)
      elapsedInterval = setInterval(() => {
        elapsedSeconds.value++
      }, 1000)
    }
  } catch (err) {
    startError.value = String(err)
  }
}

async function handleStopSuite() {
  if (!createdSuiteId.value) return
  if (!confirm('Stop the entire suite? All remaining items will be cancelled.')) return
  try {
    const result = await AutoBenchBinding.StopSuite(createdSuiteId.value)
    if (result.error) {
      startError.value = result.error
    } else {
      stopPolling()
      await fetchSuiteStatus()
    }
  } catch (err) {
    startError.value = String(err)
  }
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

function formatDurationMs(ms) {
  const s = Math.round(ms / 1000)
  if (s < 60) return `${s}s`
  const m = Math.floor(s / 60)
  return `${m}m ${s % 60}s`
}

function formatSuiteDuration() {
  if (!suiteStatus.value?.started_at) return '-'
  const start = new Date(suiteStatus.value.started_at).getTime()
  const end = suiteStatus.value.ended_at ? new Date(suiteStatus.value.ended_at).getTime() : Date.now()
  return formatDurationMs(end - start)
}

function formatItemDuration(item) {
  if (!item.started_at) return '-'
  const start = new Date(item.started_at).getTime()
  const end = item.ended_at ? new Date(item.ended_at).getTime() : Date.now()
  return formatDurationMs(end - start)
}

function formatPhaseInfo(item) {
  if (item.status === 'pending') return '-'
  if (item.status === 'success' || item.status === 'failed' || item.status === 'skipped') {
    if (!item.phase_timings?.length) return item.status
    return item.phase_timings.map(p => `${p.phase}: ${formatDurationMs(p.duration_ms)}`).join(', ')
  }
  if (item.phase_status) return `Phase: ${item.phase_status}`
  return item.status
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
        <button v-if="createdSuiteId && !isSuiteRunning && !isSuitePaused" class="primary-action start-action" type="button" :disabled="!canStartSuite" @click="handleStartSuite">
          {{ isStarting ? 'Starting...' : 'Start Suite' }}
        </button>
        <button v-if="isSuiteRunning && !isSuitePaused" class="primary-action warning-action" type="button" @click="handlePauseSuite">
          Pause
        </button>
        <button v-if="isSuitePaused" class="primary-action start-action" type="button" @click="handleResumeSuite">
          Resume
        </button>
        <button v-if="isSuiteRunning || isSuitePaused" class="primary-action danger-action" type="button" @click="handleStopSuite">
          Stop
        </button>
        <button v-if="createdSuiteId && !isSuiteRunning && !isSuitePaused" class="secondary-action" type="button" @click="resetSuite">
          New Suite
        </button>
        <button
          v-if="isSuiteTerminal"
          class="primary-action start-action"
          type="button"
          :disabled="isCreating || isStarting"
          @click="handleReRunSuite"
        >
          {{ isCreating || isStarting ? 'Re-running...' : 'Re-run' }}
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
          {{ getStatusLabel(suiteStatus.status) }}{{ isSuitePaused ? ' (Paused)' : '' }}
        </span>
        <span class="run-name">{{ suiteStatus.name || createdSuiteId }}</span>
        <span class="run-progress">
          {{ suiteStatus.completed_items }}/{{ suiteStatus.total_items }}
        </span>
        <span v-if="currentItem && !isSuitePaused" class="run-current">
          Running: {{ connNameMap[currentItem.connection_id] || currentItem.connection_id }} - {{ currentItem.profile_type }}
        </span>
        <span v-if="isSuiteRunning || isSuitePaused" class="run-elapsed">{{ formatElapsed(elapsedSeconds) }}</span>
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
          <div class="metric">
            <span class="metric-label">Duration</span>
            <span class="metric-value">{{ formatSuiteDuration() }}</span>
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
            <span>Duration</span>
            <span>Phase</span>
            <span>Flags</span>
          </div>
          <div v-for="item in suiteStatus.items" :key="item.id" class="item-row">
            <span class="item-connection">{{ connNameMap[item.connection_id] || item.connection_id }}</span>
            <span class="item-type">{{ item.profile_type }}</span>
            <span :class="['item-status', getStatusClass(item.status)]">{{ item.status }}</span>
            <span class="item-duration">{{ formatItemDuration(item) }}</span>
            <span class="item-phase" :title="formatPhaseInfo(item)">{{ formatPhaseInfo(item) }}</span>
            <span class="item-flags">
              <span v-for="cap in (connCapabilitiesMap[item.connection_id] || [])" :key="cap" :class="['cap-badge', 'cap-' + cap.toLowerCase()]">{{ cap }}</span>
              <span v-if="!(connCapabilitiesMap[item.connection_id] || []).length" class="muted">-</span>
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

.warning-action {
  background: var(--warning);
  border-color: var(--warning);
}

.danger-action {
  background: var(--danger);
  border-color: var(--danger);
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
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
}

.run-strip {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-color);
}

.run-name {
  font-weight: 700;
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
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}

.metric {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px 12px;
  background: var(--bg-secondary);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
}

.metric-label {
  font-size: 11px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.metric-value {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
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
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.items-list h4 {
  margin: 0 0 12px 0;
  color: var(--text-primary);
}

.items-header {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 0.8fr 1.2fr 0.7fr;
  gap: 12px;
  padding: 8px 12px;
  background: var(--bg-secondary);
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.3px;
  border-bottom: 2px solid var(--border-color);
}

.item-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr 0.8fr 1.2fr 0.7fr;
  gap: 12px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-color);
  border-radius: 0;
  margin-top: 0;
  font-size: 13px;
  transition: background 0.15s;
}

.item-row:hover {
  background: var(--bg-hover);
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

.item-duration {
  color: var(--text-secondary);
  font-variant-numeric: tabular-nums;
}

.item-phase {
  color: var(--text-secondary);
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-flags {
  display: flex;
  gap: 4px;
  align-items: center;
  flex-wrap: nowrap;
}

.cap-badge {
  display: inline-block;
  font-size: 10px;
  font-weight: 600;
  padding: 1px 5px;
  border-radius: 3px;
  line-height: 1.4;
  text-transform: uppercase;
}
.cap-ssh {
  background: #e0f2fe;
  color: #0369a1;
  border: 1px solid #7dd3fc;
}
.cap-winrm {
  background: #fef3c7;
  color: #92400e;
  border: 1px solid #fcd34d;
}
.cap-ai {
  background: #fce7f3;
  color: #9d174d;
  border: 1px solid #f9a8d4;
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
  display: flex;
  gap: 12px;
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
  padding: 16px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.section-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 4px;
}

.section-header-row h3 {
  font-size: 15px;
  color: var(--text-primary);
  margin: 0;
  font-weight: 700;
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

.error-banner {
  border-left: 3px solid var(--danger);
}
</style>
