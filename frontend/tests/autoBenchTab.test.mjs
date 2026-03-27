import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const autoBenchSource = fs.readFileSync(path.resolve(__dirname, '../src/components/tabs/AutoBenchTab.vue'), 'utf8')

test('T12.2: autobench renders suite status panel with metrics', () => {
  // Suite status panel container
  assert.match(autoBenchSource, /suite-status-panel/)
  // Status header with name and badge
  assert.match(autoBenchSource, /status-header/)
  assert.match(autoBenchSource, /status-badge/)
  // Status metrics grid
  assert.match(autoBenchSource, /status-metrics/)
  assert.match(autoBenchSource, /metric-label/)
  assert.match(autoBenchSource, /metric-value/)
  // Progress bar
  assert.match(autoBenchSource, /progress-bar-container/)
  assert.match(autoBenchSource, /progress-bar/)
  assert.match(autoBenchSource, /progress-label/)
  // Items list with status
  assert.match(autoBenchSource, /items-list/)
  assert.match(autoBenchSource, /item-row/)
  assert.match(autoBenchSource, /item-status/)
})

test('T12.2: autobench displays item-level status with connection, type, and report', () => {
  // Items header
  assert.match(autoBenchSource, /items-header/)
  assert.match(autoBenchSource, />Connection</)
  assert.match(autoBenchSource, />Type</)
  assert.match(autoBenchSource, />Status</)
  assert.match(autoBenchSource, />Report</)
  // Item row fields
  assert.match(autoBenchSource, /item-connection/)
  assert.match(autoBenchSource, /item-type/)
  assert.match(autoBenchSource, /item-report/)
})

test('T12.2: autobench has progress bar with percentage display', () => {
  // Progress container and bar
  assert.match(autoBenchSource, /class="progress-bar-container"/)
  assert.match(autoBenchSource, /class="progress-bar"/)
  assert.match(autoBenchSource, /:style=.*width.*progress/)
  // Progress percentage display
  assert.match(autoBenchSource, /suiteSummary\.progress/)
})

test('T12.2: autobench has status badge styling for success/error/running states', () => {
  assert.match(autoBenchSource, /getStatusClass/)
  assert.match(autoBenchSource, /status-success/)
  assert.match(autoBenchSource, /status-error/)
  assert.match(autoBenchSource, /status-running/)
  assert.match(autoBenchSource, /status-warning/)
})

test('T12.3: autobench has viewReport function for navigation to report detail', () => {
  assert.match(autoBenchSource, /function viewReport/)
  assert.match(autoBenchSource, /appStore\.setActiveTab\('reports'\)/)
})

test('T12.3: autobench has goToReports function for navigation to reports tab', () => {
  assert.match(autoBenchSource, /function goToReports/)
})

test('T12.3: autobench renders View Report button when report_id exists', () => {
  assert.match(autoBenchSource, /v-if="item\.report_id"/)
  assert.match(autoBenchSource, /View Report/)
  assert.match(autoBenchSource, /@click="viewReport\(item\.report_id\)"/)
})

test('T12.3: autobench renders View All Reports button when suite completes', () => {
  assert.match(autoBenchSource, /suite-actions/)
  assert.match(autoBenchSource, /View All Reports/)
  assert.match(autoBenchSource, /@click="goToReports"/)
})

test('T12.1: autobench uses real connections from store', () => {
  assert.match(autoBenchSource, /useConnectionStore/)
  assert.match(autoBenchSource, /realConnections/)
  assert.match(autoBenchSource, /connectionStore\.connections/)
  assert.match(autoBenchSource, /connectionStore\.fetchConnections/)
})

test('T12.1: autobench loads connections on mount', () => {
  assert.match(autoBenchSource, /onMounted/)
  assert.match(autoBenchSource, /await connectionStore\.fetchConnections/)
})

test('M10/T10.2: autobench has CreateSuite API binding', () => {
  assert.match(autoBenchSource, /AutoBenchBinding/)
  assert.match(autoBenchSource, /AutoBenchBinding\.CreateSuite/)
  assert.match(autoBenchSource, /handleCreateSuite/)
  assert.match(autoBenchSource, /canCreateSuite/)
})

test('M11/T11.3: autobench has StartSuite API binding', () => {
  assert.match(autoBenchSource, /AutoBenchBinding\.StartSuite/)
  assert.match(autoBenchSource, /handleStartSuite/)
  assert.match(autoBenchSource, /canStartSuite/)
})

test('M11/T11.3: autobench has GetSuiteStatus API binding for polling', () => {
  assert.match(autoBenchSource, /AutoBenchBinding\.GetSuiteStatus/)
  assert.match(autoBenchSource, /fetchSuiteStatus/)
  assert.match(autoBenchSource, /pollSuiteStatus/)
  assert.match(autoBenchSource, /isPolling/)
})

test('T12.2: autobench has computed suiteSummary with progress calculation', () => {
  assert.match(autoBenchSource, /const suiteSummary = computed/)
  assert.match(autoBenchSource, /total_items/)
  assert.match(autoBenchSource, /completed_items/)
  assert.match(autoBenchSource, /pending_items/)
  assert.match(autoBenchSource, /running_items/)
})

test('T12.2: autobench button states for create/start/running', () => {
  // Create Suite button
  assert.match(autoBenchSource, /:disabled="!canCreateSuite"/)
  // Start Suite button
  assert.match(autoBenchSource, /:disabled="!canStartSuite"/)
  // Running state
  assert.match(autoBenchSource, /isSuiteRunning/)
  // Reset/New Suite button
  assert.match(autoBenchSource, /resetSuite/)
  assert.match(autoBenchSource, /New Suite/)
})

test('T12.2: autobench has error handling for create and start operations', () => {
  assert.match(autoBenchSource, /createError/)
  assert.match(autoBenchSource, /startError/)
  assert.match(autoBenchSource, /error-banner/)
})

test('T12.2: autobench cleans up polling on unmount', () => {
  assert.match(autoBenchSource, /onUnmounted/)
  assert.match(autoBenchSource, /clearTimeout/)
})

test('T12.2: autobench displays connection type filter with database type options', () => {
  assert.match(autoBenchSource, /connectionFilterOptions/)
  assert.match(autoBenchSource, /activeConnectionFilter/)
  assert.match(autoBenchSource, /filteredConnections/)
})

test('T12.2: autobench uses wizard draft for connection and profile selection', () => {
  assert.match(autoBenchSource, /createAutoBenchWizardDraft/)
  assert.match(autoBenchSource, /toggleDraftConnectionSelection/)
  assert.match(autoBenchSource, /toggleDraftProfileSelection/)
  assert.match(autoBenchSource, /validateAutoBenchWizardDraft/)
  assert.match(autoBenchSource, /profileOptions/)
  assert.match(autoBenchSource, /selectedProfiles/)
  assert.match(autoBenchSource, /selectedConnectionIds/)
})
