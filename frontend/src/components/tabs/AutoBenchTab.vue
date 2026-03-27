<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
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

const connectionStore = useConnectionStore()
const appStore = useAppStore()

const draft = ref(createAutoBenchWizardDraft())
const activeConnectionFilter = ref('all')
const isCreating = ref(false)
const isStarting = ref(false)
const createError = ref('')
const startError = ref('')
const createdSuiteId = ref('')
const suiteStatus = ref(null)
const isPolling = ref(false)
let pollTimeout = null
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
  // Can only start if status is draft or ready
  return ['draft', 'ready'].includes(suiteStatus.value.status)
})

const isSuiteRunning = computed(() => {
  return suiteStatus.value?.status === 'running'
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

// Load connections on mount
onMounted(async () => {
  await connectionStore.fetchConnections()
})

onUnmounted(() => {
  if (pollTimeout) {
    clearTimeout(pollTimeout)
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
      // Fetch initial status
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
      // Start polling for status
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

  // Stop polling if too many consecutive errors
  if (consecutiveErrors >= MAX_CONSECUTIVE_ERRORS) {
    stopPolling()
    return
  }

  // Continue polling if running, with backoff on errors
  if (suiteStatus.value?.status === 'running') {
    const delay = consecutiveErrors > 0 ? 1000 * Math.min(consecutiveErrors, 5) : 1000
    pollTimeout = setTimeout(pollSuiteStatus, delay)
  } else {
    stopPolling()
  }
}

function viewReport(reportId) {
  if (!reportId) return
  // Navigate to Reports tab and show the report
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
  stopPolling()
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
          Automated database benchmark suite. Select connections and profiles to create and run benchmark suites.
        </p>
      </div>
      <div class="header-actions">
        <button class="primary-action" type="button" :disabled="!canCreateSuite" @click="handleCreateSuite">
          {{ isCreating ? 'Creating...' : 'Create Suite' }}
        </button>
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

    <div v-if="createdSuiteId && suiteStatus" class="suite-status-panel">
      <div class="status-header">
        <h3>Suite: {{ suiteStatus.name || createdSuiteId }}</h3>
        <span :class="['status-badge', getStatusClass(suiteStatus.status)]">
          {{ getStatusLabel(suiteStatus.status) }}
        </span>
      </div>

      <div class="status-metrics">
        <div class="metric">
          <span class="metric-label">Total Items</span>
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
        <h4>Items ({{ suiteStatus.items.length }})</h4>
        <div class="items-header">
          <span>Connection</span>
          <span>Type</span>
          <span>Status</span>
          <span>Report</span>
        </div>
        <div v-for="item in suiteStatus.items" :key="item.id" class="item-row">
          <span class="item-connection">{{ item.connection_id }}</span>
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

      <div v-if="['success', 'partial_success', 'failed'].includes(suiteStatus.status)" class="suite-actions">
        <button class="primary-action" @click="goToReports">
          View All Reports
        </button>
      </div>
    </div>

    <div v-if="!createdSuiteId" class="autobench-grid">
      <section class="autobench-section autobench-wizard" aria-labelledby="autobench-wizard-title">
        <div class="section-header">
          <h2 id="autobench-wizard-title">Wizard</h2>
          <p>Select connections and profiles to create a benchmark suite.</p>
        </div>

        <div class="wizard-groups">
          <section class="wizard-group" aria-labelledby="autobench-connections-title">
            <div class="wizard-group-header">
              <h3 id="autobench-connections-title">Connections</h3>
              <span class="wizard-chip">{{ draft.selectedConnectionIds.length }} selected</span>
            </div>

            <div v-if="connectionStore.loading" class="loading-state">
              Loading connections...
            </div>
            <div v-else-if="realConnections.length === 0" class="empty-state">
              No connections available. Create connections first.
            </div>
            <template v-else>
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
              <label
                v-for="connection in filteredConnections"
                :key="connection.id"
                class="wizard-option-card"
              >
                <input
                  type="checkbox"
                  :checked="draft.selectedConnectionIds.includes(connection.id)"
                  @change="toggleConnectionSelection(connection.id)"
                >
                <span class="wizard-option-copy">
                  <strong>{{ connection.label }}</strong>
                  <small class="wizard-option-meta">{{ connection.databaseType }}</small>
                  <small>{{ connection.detail }}</small>
                </span>
              </label>
            </template>
            <p v-if="wizardValidation.connectionError" class="wizard-validation">{{ wizardValidation.connectionError }}</p>
          </section>

          <section class="wizard-group" aria-labelledby="autobench-benchmark-types-title">
            <div class="wizard-group-header">
              <h3 id="autobench-benchmark-types-title">Benchmark Types</h3>
              <span class="wizard-chip">{{ draft.selectedProfiles.length }} selected</span>
            </div>

            <label
              v-for="profile in profileOptions"
              :key="profile.id"
              class="wizard-option-card wizard-option-card-profile"
            >
              <input
                type="checkbox"
                :checked="draft.selectedProfiles.includes(profile.id)"
                @change="toggleProfileSelection(profile.id)"
              >
              <span class="wizard-option-copy">
                <strong>{{ profile.label }}</strong>
                <small class="wizard-option-meta">{{ profile.scope }}</small>
                <small>{{ profile.description }}</small>
              </span>
            </label>
            <p class="wizard-order">Selected order: {{ selectedProfileSummary }}</p>
            <p v-if="wizardValidation.profileError" class="wizard-validation">{{ wizardValidation.profileError }}</p>
          </section>

          <section class="wizard-group">
            <div class="wizard-group-header">
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
        </div>
      </section>
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

.suite-status-panel {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.status-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.status-header h3 {
  margin: 0;
  color: var(--text-primary);
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

.autobench-grid {
  display: grid;
  gap: 16px;
}

.autobench-section {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.section-header h2 {
  font-size: 18px;
  color: var(--text-primary);
  margin: 0;
}

.section-header p {
  margin-top: 8px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.wizard-groups {
  margin-top: 18px;
  display: grid;
  gap: 18px;
}

.wizard-group {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
  background: var(--bg-secondary);
}

.wizard-group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.wizard-group-header h3 {
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
  margin-top: 14px;
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

.wizard-option-card {
  margin-top: 12px;
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  background: var(--bg-primary);
  cursor: pointer;
}

.wizard-option-card-profile {
  align-items: center;
}

.wizard-option-copy {
  display: grid;
  gap: 4px;
}

.wizard-option-copy strong {
  color: var(--text-primary);
  font-size: 14px;
}

.wizard-option-copy small {
  color: var(--text-secondary);
  line-height: 1.4;
}

.wizard-option-meta {
  text-transform: uppercase;
  letter-spacing: 0.04em;
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
